package service

import (
	"strings"
	"testing"
	"time"

	orderEntity "github.com/Taishi-co/SteelLink-backend/internal/modules/order/domain/entity"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/order/ui/dto"
)

func TestBuildOrderTimelineFinancingDisbursementEnablesLogistics(t *testing.T) {
	now := time.Now()
	order := &orderEntity.Order{
		ID:        1,
		OrderNo:   "ORD-TIMELINE-1",
		FinanceID: 10,
		Status:    orderEntity.OrderStatusPaid,
		CreatedAt: now,
		PaymentInfo: &orderEntity.PaymentInfo{
			PaymentMethod: "financing_disbursement",
			VoucherURL:    "https://example.com/disbursement.pdf",
			PaymentTime:   now,
			UploadedAt:    now,
		},
	}
	events := []orderEntity.OrderEvent{{EventType: "financing_disbursed", CreatedAt: now}}

	items := buildOrderTimeline(order, events)
	disburseIndex := timelineIndex(items, "financing_disburse")
	logisticsIndex := timelineIndex(items, "logistics_ready")
	if disburseIndex < 0 || logisticsIndex < 0 {
		t.Fatalf("expected disbursement and logistics-ready items, got %#v", items)
	}
	if disburseIndex >= logisticsIndex {
		t.Fatalf("expected disbursement before logistics, got %d >= %d", disburseIndex, logisticsIndex)
	}
	if items[logisticsIndex].Status != "current" {
		t.Fatalf("expected logistics-ready current, got %s", items[logisticsIndex].Status)
	}
	if timelineIndex(items, "financing_pledge") >= 0 {
		t.Fatal("did not expect pledge before logistics delivery")
	}
	paymentIndex := timelineIndex(items, "payment")
	if paymentIndex < 0 || !strings.Contains(items[paymentIndex].Description, "付款凭证已同步") {
		t.Fatalf("expected payment item to mention synced voucher, got %#v", items[paymentIndex])
	}
}

func TestBuildOrderTimelineFinancingDeliveredWaitsForFinanceReceipt(t *testing.T) {
	now := time.Now()
	order := &orderEntity.Order{
		ID:        2,
		OrderNo:   "ORD-TIMELINE-2",
		FinanceID: 20,
		Status:    orderEntity.OrderStatusDelivered,
		CreatedAt: now,
		UpdatedAt: now,
		PaymentInfo: &orderEntity.PaymentInfo{
			PaymentMethod: "financing_disbursement",
			VoucherURL:    "https://example.com/disbursement.pdf",
			PaymentTime:   now,
			UploadedAt:    now,
		},
	}
	events := []orderEntity.OrderEvent{
		{EventType: "financing_disbursed", CreatedAt: now},
		{EventType: "logistics_delivered", CreatedAt: now},
	}

	items := buildOrderTimeline(order, events)
	receiptIndex := timelineIndex(items, "receipt")
	if receiptIndex < 0 {
		t.Fatalf("expected receipt item, got %#v", items)
	}
	receipt := items[receiptIndex]
	if receipt.Status != "current" {
		t.Fatalf("expected receipt current, got %s", receipt.Status)
	}
	if receipt.Label != "资方确认监管收货并质押" {
		t.Fatalf("expected finance receipt label, got %s", receipt.Label)
	}
	if !strings.Contains(receipt.Description, "完成货权质押") {
		t.Fatalf("expected pledge completion description, got %s", receipt.Description)
	}
}

func TestBuildOrderTimelineFinancingReceiptCompletesPledge(t *testing.T) {
	now := time.Now()
	order := &orderEntity.Order{
		ID:        3,
		OrderNo:   "ORD-TIMELINE-3",
		FinanceID: 30,
		Status:    orderEntity.OrderStatusReceived,
		CreatedAt: now,
		ReceiptAt: &now,
		PaymentInfo: &orderEntity.PaymentInfo{
			PaymentMethod: "financing_disbursement",
			VoucherURL:    "https://example.com/disbursement.pdf",
			PaymentTime:   now,
			UploadedAt:    now,
		},
	}
	events := []orderEntity.OrderEvent{
		{EventType: "financing_disbursed", CreatedAt: now},
		{EventType: "finance_received", CreatedAt: now},
	}

	items := buildOrderTimeline(order, events)
	receiptIndex := timelineIndex(items, "receipt")
	if receiptIndex < 0 {
		t.Fatalf("expected receipt item, got %#v", items)
	}
	receipt := items[receiptIndex]
	if receipt.Status != "completed" {
		t.Fatalf("expected receipt completed, got %s", receipt.Status)
	}
	if !strings.Contains(receipt.Description, "质押已完成") {
		t.Fatalf("expected completed pledge description, got %s", receipt.Description)
	}
}

func timelineIndex(items []dto.TimelineItem, key string) int {
	for i, item := range items {
		if item.Key == key {
			return i
		}
	}
	return -1
}
