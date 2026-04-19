package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/steellink/backend/config"
	"github.com/steellink/backend/internal/pkg/database"
	"github.com/steellink/backend/internal/pkg/handler"
	"github.com/steellink/backend/internal/pkg/middleware"
	"github.com/steellink/backend/pkg/logger"
	"go.uber.org/zap"
)

var (
	configPath = flag.String("config", "config/config.yaml", "config file path")
	version    = "1.0.0"
)

func main() {
	flag.Parse()

	// Load .env into the process environment (optional; production can set env vars directly)
	_ = godotenv.Load()

	// Initialize config
	if err := config.Init(*configPath); err != nil {
		fmt.Printf("Failed to init config: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Get()

	// Initialize logger
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting SteelLink Backend",
		zap.String("version", version),
		zap.String("config", *configPath),
	)

	// Initialize database
	if err := database.InitMySQL(&cfg.Database); err != nil {
		logger.Fatal("Failed to init MySQL", zap.Error(err))
	}
	defer database.CloseMySQL()

	// Initialize redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		logger.Fatal("Failed to init Redis", zap.Error(err))
	}
	defer database.CloseRedis()

	// Setup Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		middleware.Recovery(),
		middleware.Logger(),
		middleware.CORS(),
	)

	// Health check endpoints (no auth required)
	router.GET("/health", handler.Health)
	router.GET("/ready", handler.Ready)
	router.GET("/live", handler.Live)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check under API
		v1.GET("/health", handler.Health)
	}

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Server starting",
			zap.Int("port", cfg.Server.Port),
			zap.String("mode", cfg.Server.Mode),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
