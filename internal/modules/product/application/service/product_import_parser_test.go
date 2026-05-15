package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Taishi-co/SteelLink-backend/config"
)

func TestParseProductImportRowsWithHeader(t *testing.T) {
	text := `品种,牌号,规格,产地,存放地,捆包号,库存,单价
热轧板卷,Q235B,3.0*1500*C,唐钢,唐山仓库A,B-001,52.5,3820
钢管,Q355B,Φ25*6*12000,宝钢,上海仓库,B-002,18,4900`

	items := parseProductImportRows(text)
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	first := items[0].Product
	if first.Category != "热轧板卷" {
		t.Fatalf("category = %q", first.Category)
	}
	if first.Spec.Material != "Q235B" {
		t.Fatalf("material = %q", first.Spec.Material)
	}
	if first.Spec.HeightSpecValue != 3 {
		t.Fatalf("height spec = %v", first.Spec.HeightSpecValue)
	}
	if first.Spec.WidthSpecValue != 1500 {
		t.Fatalf("width spec = %v", first.Spec.WidthSpecValue)
	}
	if first.Spec.LengthType != "coil" {
		t.Fatalf("length type = %q", first.Spec.LengthType)
	}
	if first.Quantity != 52.5 || first.UnitPrice != 3820 {
		t.Fatalf("quantity/unit price = %v/%v", first.Quantity, first.UnitPrice)
	}

	second := items[1].Product
	if second.Spec.HeightSpecType != "diameter" {
		t.Fatalf("pipe height spec type = %q", second.Spec.HeightSpecType)
	}
	if second.Spec.WidthSpecType != "wall_thickness" {
		t.Fatalf("pipe width spec type = %q", second.Spec.WidthSpecType)
	}
	if second.Spec.LengthValue != 12000 {
		t.Fatalf("pipe length value = %v", second.Spec.LengthValue)
	}
}

func TestParseProductImportRowsWithoutHeader(t *testing.T) {
	text := "唐钢 Q235B 热轧板卷 3.0*1500*C 库存50 单价3800"

	items := parseProductImportRows(text)
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	item := items[0].Product
	if item.Spec.Material != "Q235B" {
		t.Fatalf("material = %q", item.Spec.Material)
	}
	if item.Category != "热轧板卷" {
		t.Fatalf("category = %q", item.Category)
	}
	if item.SteelMill != "唐钢" {
		t.Fatalf("steel mill = %q", item.SteelMill)
	}
	if item.Quantity != 50 || item.UnitPrice != 3800 {
		t.Fatalf("quantity/unit price = %v/%v", item.Quantity, item.UnitPrice)
	}
}

func TestParseProductImportWithMinimaxTokenPlanVision(t *testing.T) {
	t.Setenv("MINIMAX_API_KEY", "test-key")
	previousConfig := config.GlobalConfig
	defer func() {
		config.GlobalConfig = previousConfig
	}()

	var requestPayload struct {
		Prompt   string `json:"prompt"`
		ImageURL string `json:"image_url"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization = %q", got)
		}
		if got := r.Header.Get("MM-API-Source"); got != "SteelLink-Backend" {
			t.Fatalf("MM-API-Source = %q", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"content": "{\"items\":[{\"category\":\"热轧板卷\",\"material\":\"Q235B\",\"spec\":\"3.0*1500*C\",\"steel_mill\":\"唐钢\",\"warehouse_name\":\"唐山仓库A\",\"quantity\":52.5,\"unit_price\":3820,\"confidence\":0.91,\"warnings\":[]}],\"warnings\":[]}",
			"base_resp": {"status_code": 0, "status_msg": "success"}
		}`))
	}))
	defer server.Close()

	config.GlobalConfig = &config.Config{
		AIProcurement: config.AIProcurementConfig{
			ModelParsingEnabled: true,
			Provider:            "minimax",
			APIKeyEnv:           "MINIMAX_API_KEY",
			Model:               "MiniMax-M2.7",
			VisionProvider:      "minimax_token_plan",
			VisionBaseURL:       server.URL,
			VisionAPIKeyEnv:     "MINIMAX_API_KEY",
			TimeoutSec:          5,
		},
	}

	items, warnings, usedModel, err := parseProductImportWithModel(context.Background(), productImportSource{
		Filename:   "ledger.png",
		SourceType: "image",
		IsImage:    true,
		DataURL:    "data:image/png;base64,AAAA",
	}, "优先识别库存和单价")
	if err != nil {
		t.Fatalf("parseProductImportWithModel err = %v", err)
	}
	if !usedModel {
		t.Fatalf("usedModel = false")
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v", warnings)
	}
	if requestPayload.ImageURL != "data:image/png;base64,AAAA" {
		t.Fatalf("image_url = %q", requestPayload.ImageURL)
	}
	if requestPayload.Prompt == "" {
		t.Fatalf("prompt is empty")
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	product := items[0].Product
	if product.Category != "热轧板卷" || product.Spec.Material != "Q235B" || product.SteelMill != "唐钢" {
		t.Fatalf("unexpected product = %#v", product)
	}
	if product.Spec.HeightSpecValue != 3 || product.Spec.WidthSpecValue != 1500 || product.Spec.LengthType != "coil" {
		t.Fatalf("unexpected parsed spec = %#v", product.Spec)
	}
}

func TestParseProductImportWithMinimaxTokenPlanVisionAPIError(t *testing.T) {
	t.Setenv("MINIMAX_API_KEY", "test-key")
	previousConfig := config.GlobalConfig
	defer func() {
		config.GlobalConfig = previousConfig
	}()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"base_resp":{"status_code":2049,"status_msg":"invalid api key"}}`))
	}))
	defer server.Close()

	config.GlobalConfig = &config.Config{
		AIProcurement: config.AIProcurementConfig{
			ModelParsingEnabled: true,
			VisionProvider:      "minimax_token_plan",
			VisionBaseURL:       server.URL,
			VisionAPIKeyEnv:     "MINIMAX_API_KEY",
			TimeoutSec:          5,
		},
	}

	_, _, usedModel, err := parseProductImportWithModel(context.Background(), productImportSource{
		Filename:   "ledger.png",
		SourceType: "image",
		IsImage:    true,
		DataURL:    "data:image/png;base64,AAAA",
	}, "")
	if !usedModel {
		t.Fatalf("usedModel = false")
	}
	if err == nil {
		t.Fatalf("err = nil")
	}
}
