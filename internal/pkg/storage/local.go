package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type LocalClient struct {
	baseDir string
	baseURL string
	maxSize int64
}

func NewLocalClient(baseDir, baseURL string, maxSize int64) *LocalClient {
	return &LocalClient{baseDir: baseDir, baseURL: baseURL, maxSize: maxSize}
}

func (l *LocalClient) Upload(ctx context.Context, purpose Purpose, header *multipart.FileHeader) (*UploadResult, error) {
	if header == nil {
		return nil, fmt.Errorf("file is required")
	}
	if header.Size <= 0 {
		return nil, fmt.Errorf("file is empty")
	}
	if l.maxSize > 0 && header.Size > l.maxSize {
		return nil, fmt.Errorf("file exceeds %dMB limit", l.maxSize/(1<<20))
	}

	if err := ValidateExtension(purpose, header.Filename); err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	prefix := localPrefix(purpose)
	key := path.Join(prefix, time.Now().Format("2006/01/02"), randomID()+ext)

	fullPath := filepath.Join(l.baseDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	file, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload file: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}
	if ct := header.Header.Get("Content-Type"); ct != "" {
		contentType = ct
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	return &UploadResult{
		Bucket:      "local",
		Key:         key,
		URL:         strings.TrimRight(l.baseURL, "/") + "/" + key,
		Filename:    header.Filename,
		Size:        header.Size,
		ContentType: contentType,
	}, nil
}

func localPrefix(purpose Purpose) string {
	switch purpose {
	case PurposeProductImage:
		return "products/images"
	case PurposeProductQualification:
		return "products/qualifications"
	case PurposePaymentVoucher:
		return "orders/payment-vouchers"
	case PurposeReceiptVoucher:
		return "orders/receipt-vouchers"
	case PurposeIdentityCard:
		return "identity/cards"
	case PurposeEnterpriseLicense:
		return "enterprise/licenses"
	default:
		return "misc"
	}
}
