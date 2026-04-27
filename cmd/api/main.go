package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	_ "github.com/Taishi-co/SteelLink-backend/docs" // swagger docs
	aiProcurementPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/infrastructure/persistence"
	aiProcurementRouter "github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/infrastructure/router"
	marketpersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/market/infrastructure/persistence"
	marketRouter "github.com/Taishi-co/SteelLink-backend/internal/modules/market/infrastructure/router"
	orderEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	orderRouter "github.com/Taishi-co/SteelLink-backend/internal/modules/order/infrastructure/router"
	productEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/product/domain/entity"
	productRouter "github.com/Taishi-co/SteelLink-backend/internal/modules/product/infrastructure/router"
	supplyDemandPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/infrastructure/persistence"
	supplyDemandRouter "github.com/Taishi-co/SteelLink-backend/internal/modules/supplydemand/infrastructure/router"
	tradepersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/trade/infrastructure/persistence"
	tradeRouter "github.com/Taishi-co/SteelLink-backend/internal/modules/trade/infrastructure/router"
	userPersistence "github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/persistence"
	userRouter "github.com/Taishi-co/SteelLink-backend/internal/modules/user/infrastructure/router"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/database"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/handler"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/middleware"
	"github.com/Taishi-co/SteelLink-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	configPath = flag.String("config", "config/config.yaml", "config file path")
	version    = "1.0.0"
)

// @title SteelLink API
// @version 1.0
// @description B2B钢材交易平台API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@steellink.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
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
	logger.Info("Connecting to MySQL...")
	if err := database.InitMySQL(&cfg.Database); err != nil {
		logger.Fatal("Failed to init MySQL", zap.Error(err))
	}
	defer database.CloseMySQL()
	switch os.Getenv("AUTO_MIGRATE_SCOPE") {
	case "market":
		if err := marketpersistence.AutoMigrate(); err != nil {
			logger.Fatal("Failed to auto migrate market tables", zap.Error(err))
		}
	case "skip":
		logger.Info("Skipping database auto migration")
	default:
		if err := userPersistence.AutoMigrate(); err != nil {
			logger.Fatal("Failed to migrate user tables", zap.Error(err))
		}
		if err := tradepersistence.AutoMigrate(); err != nil {
			logger.Fatal("Failed to auto migrate trade tables", zap.Error(err))
		}
		if err := supplyDemandPersistence.AutoMigrate(); err != nil {
			logger.Fatal("Failed to migrate supply demand tables", zap.Error(err))
		}
		if err := aiProcurementPersistence.AutoMigrate(); err != nil {
			logger.Fatal("Failed to migrate AI procurement tables", zap.Error(err))
		}
		if err := marketpersistence.AutoMigrate(); err != nil {
			logger.Fatal("Failed to auto migrate market tables", zap.Error(err))
		}
		if err := normalizeLegacyJSONColumns(database.GetDB()); err != nil {
			logger.Fatal("Failed to normalize legacy JSON columns", zap.Error(err))
		}
		if err := database.GetDB().AutoMigrate(&productEntity.Product{}, &orderEntity.Order{}); err != nil {
			logger.Fatal("Failed to migrate business tables", zap.Error(err))
		}
	}

	// Initialize redis
	logger.Info("Connecting to Redis...")
	if err := database.InitRedis(&cfg.Redis); err != nil {
		logger.Fatal("Failed to init Redis", zap.Error(err))
	}
	defer database.CloseRedis()

	// Setup Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	ginRouter := gin.New()
	ginRouter.Use(
		middleware.Recovery(),
		middleware.Logger(),
		middleware.CORS(),
	)

	// Swagger documentation
	ginRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoints (no auth required)
	ginRouter.GET("/health", handler.Health)
	ginRouter.GET("/ready", handler.Ready)
	ginRouter.GET("/live", handler.Live)

	// API v1 routes
	v1 := ginRouter.Group("/api/v1")
	{
		// Health check under API
		v1.GET("/health", handler.Health)
		v1.POST("/uploads", middleware.AuthRequired(), handler.UploadFile)
	}

	// 用户模块路由
	logger.Info("Registering user routes...")
	userRouter.RegisterRoutes(v1)
	logger.Info("User routes registered")
	logger.Info("Registering product routes...")
	productRouter.RegisterRoutes(v1)
	logger.Info("Product routes registered")

	logger.Info("Registering order routes...")
	orderRouter.RegisterRoutes(v1)
	logger.Info("Order routes registered")
	logger.Info("Registering market routes...")
	marketRouter.RegisterRoutes(v1)
	logger.Info("Market routes registered")

	logger.Info("Registering supply demand routes...")
	supplyDemandRouter.RegisterRoutes(v1)
	logger.Info("Supply demand routes registered")

	logger.Info("Registering AI procurement routes...")
	aiProcurementRouter.RegisterRoutes(v1)
	logger.Info("AI procurement routes registered")

	logger.Info("Registering trade routes...")
	tradeRouter.RegisterRoutes(v1.Group("/trade"))
	logger.Info("Trade routes registered")

	if cfg.Storage.Provider == "" || cfg.Storage.SecretID == "" {
		localDir := os.Getenv("STORAGE_LOCAL_DIR")
		if localDir == "" {
			localDir = "./uploads"
		}
		os.MkdirAll(localDir, 0755)
		ginRouter.Static("/uploads", localDir)
		logger.Info("Local file storage enabled", zap.String("dir", localDir))
	}

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      ginRouter,
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

func normalizeLegacyJSONColumns(db *gorm.DB) error {
	if err := normalizeLegacyStatusColumns(db); err != nil {
		return err
	}

	type fix struct {
		model  interface{}
		column string
		sql    string
	}
	fixes := []fix{
		{model: &productEntity.Product{}, column: "spec", sql: "UPDATE products SET spec = '{}' WHERE spec IS NULL OR TRIM(CAST(spec AS CHAR)) = '' OR JSON_VALID(spec) = 0"},
		{model: &productEntity.Product{}, column: "warehouse_info", sql: "UPDATE products SET warehouse_info = '{}' WHERE warehouse_info IS NULL OR TRIM(CAST(warehouse_info AS CHAR)) = '' OR JSON_VALID(warehouse_info) = 0"},
		{model: &productEntity.Product{}, column: "images", sql: "UPDATE products SET images = '[]' WHERE images IS NULL OR TRIM(CAST(images AS CHAR)) = '' OR JSON_VALID(images) = 0"},
		{model: &productEntity.Product{}, column: "remark", sql: "UPDATE products SET remark = '{}' WHERE remark IS NULL OR TRIM(CAST(remark AS CHAR)) = '' OR JSON_VALID(remark) = 0"},
		{model: &orderEntity.Order{}, column: "items", sql: "UPDATE orders SET items = '[]' WHERE items IS NULL OR TRIM(CAST(items AS CHAR)) = '' OR JSON_VALID(items) = 0"},
		{model: &orderEntity.Order{}, column: "contract_info", sql: "UPDATE orders SET contract_info = NULL WHERE contract_info IS NOT NULL AND (TRIM(CAST(contract_info AS CHAR)) = '' OR JSON_VALID(contract_info) = 0)"},
		{model: &orderEntity.Order{}, column: "payment_info", sql: "UPDATE orders SET payment_info = NULL WHERE payment_info IS NOT NULL AND (TRIM(CAST(payment_info AS CHAR)) = '' OR JSON_VALID(payment_info) = 0)"},
		{model: &orderEntity.Order{}, column: "delivery_info", sql: "UPDATE orders SET delivery_info = NULL WHERE delivery_info IS NOT NULL AND (TRIM(CAST(delivery_info AS CHAR)) = '' OR JSON_VALID(delivery_info) = 0)"},
		{model: &orderEntity.Order{}, column: "remark", sql: "UPDATE orders SET remark = '{}' WHERE remark IS NULL OR TRIM(CAST(remark AS CHAR)) = '' OR JSON_VALID(remark) = 0"},
	}

	for _, f := range fixes {
		if !db.Migrator().HasColumn(f.model, f.column) {
			continue
		}
		if err := db.Exec(f.sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func normalizeLegacyStatusColumns(db *gorm.DB) error {
	productStatusType, err := getColumnDataType(db, "products", "status")
	if err != nil {
		return err
	}
	if isStringColumnType(productStatusType) {
		stmt := `
			UPDATE products
			SET status = CASE LOWER(TRIM(status))
				WHEN 'draft' THEN '0'
				WHEN 'pending' THEN '1'
				WHEN 'pending_review' THEN '1'
				WHEN 'submitted' THEN '1'
				WHEN 'approved' THEN '2'
				WHEN 'on_sale' THEN '2'
				WHEN 'onsale' THEN '2'
				WHEN 'published' THEN '2'
				WHEN 'off_sale' THEN '3'
				WHEN 'offsale' THEN '3'
				WHEN 'disabled' THEN '3'
				WHEN 'rejected' THEN '4'
				WHEN 'reject' THEN '4'
				ELSE '0'
			END
			WHERE status IS NOT NULL AND TRIM(status) <> ''
		`
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	orderStatusType, err := getColumnDataType(db, "orders", "status")
	if err != nil {
		return err
	}
	if isStringColumnType(orderStatusType) {
		stmt := `
			UPDATE orders
			SET status = CASE LOWER(TRIM(status))
				WHEN 'draft' THEN '0'
				WHEN 'pending_confirm' THEN '1'
				WHEN 'pending_contract' THEN '2'
				WHEN 'pending_payment' THEN '3'
				WHEN 'paid' THEN '4'
				WHEN 'pending_delivery' THEN '5'
				WHEN 'delivering' THEN '6'
				WHEN 'delivered' THEN '7'
				WHEN 'completed' THEN '8'
				WHEN 'cancelled' THEN '9'
				WHEN 'frozen' THEN '10'
				WHEN 'exception' THEN '11'
				ELSE '0'
			END
			WHERE status IS NOT NULL AND TRIM(status) <> ''
		`
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	return nil
}

func getColumnDataType(db *gorm.DB, tableName, columnName string) (string, error) {
	var dataType sql.NullString
	err := db.Raw(`
		SELECT DATA_TYPE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?
		LIMIT 1
	`, tableName, columnName).Scan(&dataType).Error
	if err != nil {
		return "", err
	}
	if !dataType.Valid {
		return "", nil
	}
	return dataType.String, nil
}

func isStringColumnType(dataType string) bool {
	switch dataType {
	case "char", "varchar", "text", "tinytext", "mediumtext", "longtext":
		return true
	default:
		return false
	}
}
