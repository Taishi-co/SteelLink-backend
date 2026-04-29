package contract

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

type OrderSnapshot struct {
	OrderNo     string
	BuyerName   string
	SellerName  string
	Amount      float64
	Currency    string
	ConfirmedAt time.Time
}

type Document struct {
	ContractNo   string
	ContractURL  string
	ContractHash string
}

type SignRequest struct {
	ContractNo string
	Signer     string
	SignedAt   time.Time
}

type SignResult struct {
	SignedAt time.Time
}

type Provider interface {
	Create(ctx context.Context, order OrderSnapshot) (*Document, error)
	Sign(ctx context.Context, req SignRequest) (*SignResult, error)
}

type SimulatedProvider struct {
	baseURL string
}

func NewProvider() Provider {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("CONTRACT_PROVIDER"))) {
	case "", "simulated", "manual":
		return &SimulatedProvider{baseURL: contractBaseURL()}
	default:
		return &SimulatedProvider{baseURL: contractBaseURL()}
	}
}

func (p *SimulatedProvider) Create(ctx context.Context, order OrderSnapshot) (*Document, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	confirmedAt := order.ConfirmedAt
	if confirmedAt.IsZero() {
		confirmedAt = time.Now()
	}

	contractNo := fmt.Sprintf("HT%s%s", confirmedAt.Format("20060102"), sanitizeOrderNo(order.OrderNo))
	hashInput := fmt.Sprintf("%s|%s|%s|%.2f|%s|%s",
		contractNo,
		order.BuyerName,
		order.SellerName,
		order.Amount,
		order.Currency,
		confirmedAt.UTC().Format(time.RFC3339),
	)
	sum := sha256.Sum256([]byte(hashInput))

	return &Document{
		ContractNo:   contractNo,
		ContractURL:  strings.TrimRight(p.baseURL, "/") + "/" + contractNo + ".pdf",
		ContractHash: hex.EncodeToString(sum[:]),
	}, nil
}

func (p *SimulatedProvider) Sign(ctx context.Context, req SignRequest) (*SignResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if strings.TrimSpace(req.ContractNo) == "" {
		return nil, fmt.Errorf("contract no is required")
	}
	if strings.TrimSpace(req.Signer) == "" {
		return nil, fmt.Errorf("contract signer is required")
	}

	signedAt := req.SignedAt
	if signedAt.IsZero() {
		signedAt = time.Now()
	}
	return &SignResult{SignedAt: signedAt}, nil
}

func contractBaseURL() string {
	if baseURL := strings.TrimSpace(os.Getenv("CONTRACT_BASE_URL")); baseURL != "" {
		return baseURL
	}
	return "https://contracts.steellink.local/simulated"
}

func sanitizeOrderNo(orderNo string) string {
	orderNo = strings.ToUpper(strings.TrimSpace(orderNo))
	var b strings.Builder
	for _, c := range orderNo {
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			b.WriteRune(c)
		}
	}
	if b.Len() == 0 {
		return "ORDER"
	}
	return b.String()
}
