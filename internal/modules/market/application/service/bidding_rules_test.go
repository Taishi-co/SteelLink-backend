package service

import (
	"testing"

	marketentity "github.com/Taishi-co/SteelLink-backend/internal/modules/market/domain/entity"
)

func TestBiddingViewScope(t *testing.T) {
	activity := marketentity.BiddingActivity{BuyerID: 10}

	tests := []struct {
		name         string
		enterpriseID uint64
		want         BiddingViewScope
	}{
		{name: "anonymous public view", enterpriseID: 0, want: BiddingViewPublic},
		{name: "buyer workspace view", enterpriseID: 10, want: BiddingViewBuyer},
		{name: "seller workspace view", enterpriseID: 20, want: BiddingViewSeller},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveBiddingViewScope(activity, tt.enterpriseID); got != tt.want {
				t.Fatalf("resolveBiddingViewScope() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBiddingQuoteVisibility(t *testing.T) {
	activity := marketentity.BiddingActivity{
		BuyerID:         10,
		Status:          marketentity.BiddingStatusPendingAward,
		SupplierName:    "中标供应商",
		AwardSupplierID: 20,
	}
	quotes := []marketentity.BiddingQuote{
		{ID: 1, BiddingID: 100, SupplierID: 20, SupplierName: "中标供应商", UnitPrice: 3900, Remark: "可提前交付", IsAwarded: true},
		{ID: 2, BiddingID: 100, SupplierID: 30, SupplierName: "其他供应商", UnitPrice: 3850, Remark: "最低价"},
		{ID: 3, BiddingID: 100, SupplierID: 40, SupplierName: "第三供应商", UnitPrice: 4000, Remark: "普通报价"},
	}

	publicItem := toBiddingItem(activity, quotes, BiddingViewPublic, 0)
	if len(publicItem.Quotes) != 0 {
		t.Fatalf("public view quotes len = %d, want 0", len(publicItem.Quotes))
	}
	if publicItem.SupplierName != "" || publicItem.AwardSupplierID != 0 {
		t.Fatalf("public pending_award view leaked supplier: name=%q id=%d", publicItem.SupplierName, publicItem.AwardSupplierID)
	}

	buyerItem := toBiddingItem(activity, quotes, BiddingViewBuyer, 10)
	if len(buyerItem.Quotes) != len(quotes) {
		t.Fatalf("buyer view quotes len = %d, want %d", len(buyerItem.Quotes), len(quotes))
	}
	if buyerItem.Quotes[0].SupplierName == "" || buyerItem.Quotes[1].SupplierName == "" {
		t.Fatalf("buyer view should include supplier identities")
	}

	sellerItem := toBiddingItem(activity, quotes, BiddingViewSeller, 30)
	if len(sellerItem.Quotes) != 2 {
		t.Fatalf("seller view quotes len = %d, want own quote plus awarded quote", len(sellerItem.Quotes))
	}
	if sellerItem.Quotes[0].SupplierID != 0 || sellerItem.Quotes[0].SupplierName != "" || sellerItem.Quotes[0].Remark != "" {
		t.Fatalf("seller view leaked awarded competitor quote: %+v", sellerItem.Quotes[0])
	}
	if sellerItem.Quotes[1].SupplierID != 30 || sellerItem.Quotes[1].SupplierName == "" {
		t.Fatalf("seller view should include own quote identity: %+v", sellerItem.Quotes[1])
	}
}

func TestCanAwardBiddingStatus(t *testing.T) {
	tests := []struct {
		status marketentity.BiddingStatus
		want   bool
	}{
		{status: marketentity.BiddingStatusBidding, want: false},
		{status: marketentity.BiddingStatusPendingAward, want: true},
		{status: marketentity.BiddingStatusAwarded, want: false},
		{status: marketentity.BiddingStatusOrderCreated, want: false},
		{status: marketentity.BiddingStatusFailed, want: false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := canAwardBiddingStatus(tt.status); got != tt.want {
				t.Fatalf("canAwardBiddingStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestNormalizeBiddingStatus(t *testing.T) {
	tests := []struct {
		status marketentity.BiddingStatus
		want   marketentity.BiddingStatus
	}{
		{status: marketentity.BiddingStatusOngoing, want: marketentity.BiddingStatusBidding},
		{status: marketentity.BiddingStatusSprint, want: marketentity.BiddingStatusBidding},
		{status: marketentity.BiddingStatusSuccess, want: marketentity.BiddingStatusAwarded},
		{status: marketentity.BiddingStatusClosed, want: marketentity.BiddingStatusClosed},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := normalizeBiddingStatus(tt.status); got != tt.want {
				t.Fatalf("normalizeBiddingStatus(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}
