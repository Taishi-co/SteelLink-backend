package contract

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestSimulatedProviderCreate(t *testing.T) {
	t.Setenv("CONTRACT_BASE_URL", "https://example.test/contracts")
	provider := NewProvider()

	doc, err := provider.Create(context.Background(), OrderSnapshot{
		OrderNo:     "order-001",
		BuyerName:   "buyer",
		SellerName:  "seller",
		Amount:      1000,
		Currency:    "CNY",
		ConfirmedAt: time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if doc.ContractNo != "HT20260429ORDER001" {
		t.Fatalf("ContractNo = %q", doc.ContractNo)
	}
	if doc.ContractURL != "https://example.test/contracts/HT20260429ORDER001.pdf" {
		t.Fatalf("ContractURL = %q", doc.ContractURL)
	}
	if len(doc.ContractHash) != 64 {
		t.Fatalf("ContractHash length = %d", len(doc.ContractHash))
	}
}

func TestSimulatedProviderSign(t *testing.T) {
	provider := &SimulatedProvider{}
	signedAt := time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC)

	result, err := provider.Sign(context.Background(), SignRequest{
		ContractNo: "HT20260429ORDER001",
		Signer:     "buyer",
		SignedAt:   signedAt,
	})
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if !result.SignedAt.Equal(signedAt) {
		t.Fatalf("SignedAt = %s", result.SignedAt)
	}

	_, err = provider.Sign(context.Background(), SignRequest{Signer: "buyer"})
	if err == nil || !strings.Contains(err.Error(), "contract no") {
		t.Fatalf("Sign() missing contract error = %v", err)
	}
}
