package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/gin-gonic/gin"
)

func TestUploadFileTencentCOSMissingCredentialsDoesNotFallbackToLocal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("STORAGE_LOCAL_DIR", t.TempDir()+"/not-used")

	config.GlobalConfig = &config.Config{
		Server: config.ServerConfig{Port: 8080},
		Storage: config.StorageConfig{
			Provider:      "tencent_cos",
			Region:        "ap-beijing",
			PublicBucket:  "steellink-public-1321917249",
			MaxFileSizeMB: 20,
		},
	}
	resetUploadClientsForTest()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("purpose", "product_image"); err != nil {
		t.Fatalf("write purpose: %v", err)
	}
	part, err := writer.CreateFormFile("file", "product.jpg")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("fake-image")); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	UploadFile(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	responseBody := rec.Body.String()
	if !strings.Contains(responseBody, "cos credentials are not configured") {
		t.Fatalf("expected missing credentials error, got %s", responseBody)
	}
	if strings.Contains(responseBody, "create directory") {
		t.Fatalf("unexpected local storage fallback error: %s", responseBody)
	}
}

func resetUploadClientsForTest() {
	cosClientOnce = syncOnceForTest()
	cosClient = nil
	localClientOnce = syncOnceForTest()
	localClient = nil
}

func syncOnceForTest() sync.Once {
	return sync.Once{}
}
