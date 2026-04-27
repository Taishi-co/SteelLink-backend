package sms

import (
	"fmt"
	"os"
)

type Provider interface {
	SendCode(phone string, code string, purpose string) error
}

type DevProvider struct{}

func (p *DevProvider) SendCode(phone, code, purpose string) error {
	fmt.Printf("[SMS DEV] phone=%s code=%s purpose=%s\n", phone, code, purpose)
	return nil
}

type ProviderConfig struct {
	Provider string
}

func NewProvider(cfg ProviderConfig) Provider {
	switch os.Getenv("SMS_PROVIDER") {
	case "tencent":
		return &TencentProvider{}
	default:
		return &DevProvider{}
	}
}

type TencentProvider struct{}

func (p *TencentProvider) SendCode(phone, code, purpose string) error {
	secretID := os.Getenv("SMS_SECRET_ID")
	secretKey := os.Getenv("SMS_SECRET_KEY")
	if secretID == "" || secretKey == "" {
		return fmt.Errorf("tencent SMS credentials not configured")
	}
	return fmt.Errorf("tencent SMS integration not yet implemented, use SMS_PROVIDER=dev")
}
