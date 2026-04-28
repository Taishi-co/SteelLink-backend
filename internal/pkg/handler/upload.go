package handler

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/storage"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

var (
	cosClientOnce   sync.Once
	cosClient       *storage.Client
	localClientOnce sync.Once
	localClient     *storage.LocalClient
)

func getCOSClient(cfg config.StorageConfig) *storage.Client {
	cosClientOnce.Do(func() {
		cosClient = storage.NewClient(cfg)
	})
	return cosClient
}

func getLocalClient(cfg *config.Config) *storage.LocalClient {
	localClientOnce.Do(func() {
		baseDir := os.Getenv("STORAGE_LOCAL_DIR")
		if baseDir == "" {
			baseDir = "./uploads"
		}
		baseURL := os.Getenv("STORAGE_LOCAL_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:%d/uploads", cfg.Server.Port)
		}
		localClient = storage.NewLocalClient(baseDir, baseURL, cfg.Storage.MaxFileSizeBytes())
	})
	return localClient
}

func UploadFile(c *gin.Context) {
	purpose := storage.Purpose(strings.TrimSpace(c.PostForm("purpose")))
	if purpose == "" {
		purpose = storage.Purpose(strings.TrimSpace(c.Query("purpose")))
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	cfg := config.Get()
	var result *storage.UploadResult

	switch strings.TrimSpace(cfg.Storage.Provider) {
	case "", "tencent_cos":
		result, err = getCOSClient(cfg.Storage).Upload(c.Request.Context(), purpose, file)
	case "local":
		result, err = getLocalClient(cfg).Upload(c.Request.Context(), purpose, file)
	default:
		err = fmt.Errorf("unsupported storage provider: %s", cfg.Storage.Provider)
	}

	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}
