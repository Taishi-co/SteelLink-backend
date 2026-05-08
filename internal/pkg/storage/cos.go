package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type Purpose string

const (
	PurposeProductImage         Purpose = "product_image"
	PurposeProductQualification Purpose = "product_qualification"
	PurposePaymentVoucher       Purpose = "payment_voucher"
	PurposeReceiptVoucher       Purpose = "receipt_voucher"
	PurposeIdentityCard         Purpose = "identity_card"
	PurposeEnterpriseLicense    Purpose = "enterprise_license"
	PurposeEnterpriseServiceDoc Purpose = "enterprise_service_document"
)

type UploadResult struct {
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

type Client struct {
	cfg config.StorageConfig
}

func NewClient(cfg config.StorageConfig) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) Upload(ctx context.Context, purpose Purpose, header *multipart.FileHeader) (*UploadResult, error) {
	if err := c.validate(purpose, header); err != nil {
		return nil, err
	}

	file, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload file: %w", err)
	}
	defer file.Close()

	bucket, baseURL, prefix, err := c.target(purpose)
	if err != nil {
		return nil, err
	}

	key := path.Join(prefix, time.Now().Format("2006/01/02"), randomID()+strings.ToLower(filepath.Ext(header.Filename)))
	objectURL, err := url.Parse(bucketURL(bucket, c.cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("parse bucket url: %w", err)
	}

	cosClient := cos.NewClient(&cos.BaseURL{BucketURL: objectURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  c.cfg.SecretID,
			SecretKey: c.cfg.SecretKey,
		},
	})

	contentType := contentType(header, file)
	if seeker, ok := file.(io.Seeker); ok {
		_, _ = seeker.Seek(0, io.SeekStart)
	}
	_, err = cosClient.Object.Put(ctx, key, file, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("upload to cos: %w", err)
	}

	return &UploadResult{
		Bucket:      bucket,
		Key:         key,
		URL:         strings.TrimRight(baseURL, "/") + "/" + key,
		Filename:    header.Filename,
		Size:        header.Size,
		ContentType: contentType,
	}, nil
}

func (c *Client) validate(purpose Purpose, header *multipart.FileHeader) error {
	if c.cfg.Provider == "" {
		return fmt.Errorf("storage provider is not configured")
	}
	if c.cfg.Region == "" || c.cfg.SecretID == "" || c.cfg.SecretKey == "" {
		return fmt.Errorf("cos credentials are not configured")
	}
	if header == nil {
		return fmt.Errorf("file is required")
	}
	if header.Size <= 0 {
		return fmt.Errorf("file is empty")
	}
	if header.Size > c.cfg.MaxFileSizeBytes() {
		return fmt.Errorf("file exceeds %dMB limit", c.cfg.MaxFileSizeMB)
	}
	if err := ValidateExtension(purpose, header.Filename); err != nil {
		return err
	}
	return nil
}

func (c *Client) target(purpose Purpose) (bucket, baseURL, prefix string, err error) {
	switch purpose {
	case PurposeProductImage:
		bucket = c.cfg.PublicBucket
		baseURL = c.cfg.PublicBaseURL
		prefix = "products/images"
	case PurposeProductQualification:
		bucket = c.cfg.PrivateBucket
		baseURL = c.cfg.PrivateBaseURL
		prefix = "products/qualifications"
	case PurposePaymentVoucher:
		bucket = c.cfg.PrivateBucket
		baseURL = c.cfg.PrivateBaseURL
		prefix = "orders/payment-vouchers"
	case PurposeReceiptVoucher:
		bucket = c.cfg.PrivateBucket
		baseURL = c.cfg.PrivateBaseURL
		prefix = "orders/receipt-vouchers"
	case PurposeIdentityCard:
		bucket = c.cfg.PrivateBucket
		baseURL = c.cfg.PrivateBaseURL
		prefix = "identity/cards"
	case PurposeEnterpriseLicense:
		bucket = c.cfg.PrivateBucket
		baseURL = c.cfg.PrivateBaseURL
		prefix = "enterprise/licenses"
	case PurposeEnterpriseServiceDoc:
		bucket = c.cfg.PrivateBucket
		baseURL = c.cfg.PrivateBaseURL
		prefix = "enterprise/service-documents"
	default:
		err = fmt.Errorf("invalid upload purpose")
	}
	if err != nil {
		return "", "", "", err
	}
	if bucket == "" {
		return "", "", "", fmt.Errorf("cos bucket is not configured")
	}
	if baseURL == "" {
		baseURL = bucketURL(bucket, c.cfg.Region)
	}
	return bucket, baseURL, prefix, nil
}

func bucketURL(bucket, region string) string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region)
}

func allowedExt(ext string, allowed ...string) bool {
	for _, item := range allowed {
		if ext == item {
			return true
		}
	}
	return false
}

func contentType(header *multipart.FileHeader, file multipart.File) string {
	if value := header.Header.Get("Content-Type"); value != "" {
		return value
	}
	buffer := make([]byte, 512)
	n, _ := file.Read(buffer)
	return http.DetectContentType(buffer[:n])
}

func randomID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
