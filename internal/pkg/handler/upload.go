package handler

import (
	"fmt"
	"os"
	"strings"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/internal/pkg/storage"
	"github.com/Taishi-co/SteelLink-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

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

	if cfg.Storage.Provider != "" && cfg.Storage.SecretID != "" && cfg.Storage.SecretKey != "" {
		result, err = storage.NewClient(cfg.Storage).Upload(c.Request.Context(), purpose, file)
	} else {
		baseDir := os.Getenv("STORAGE_LOCAL_DIR")
		if baseDir == "" {
			baseDir = "./uploads"
		}
		baseURL := os.Getenv("STORAGE_LOCAL_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:%d/uploads", cfg.Server.Port)
		}
		result, err = storage.NewLocalClient(baseDir, baseURL, cfg.Storage.MaxFileSizeBytes()).Upload(c.Request.Context(), purpose, file)
	}

	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}
