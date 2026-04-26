package handler

import (
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

	result, err := storage.NewClient(config.Get().Storage).Upload(c.Request.Context(), purpose, file)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}
