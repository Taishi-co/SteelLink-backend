package service

import (
	"encoding/json"
	"testing"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
)

func TestBuildModelParserPayloadUsesAnthropicMessagesShape(t *testing.T) {
	payload, err := buildModelParserPayload(config.AIProcurementConfig{
		BaseURL:     "https://api.minimaxi.com/anthropic/v1/messages",
		Model:       "MiniMax-M2.7",
		MaxTokens:   1024,
		Temperature: 0.1,
	}, "采购螺纹钢")
	if err != nil {
		t.Fatalf("buildModelParserPayload returned error: %v", err)
	}

	var req anthropicMessageRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		t.Fatalf("payload is not anthropic request: %v", err)
	}
	if req.Model != "MiniMax-M2.7" {
		t.Fatalf("model = %q, want MiniMax-M2.7", req.Model)
	}
	if req.MaxTokens != 1024 {
		t.Fatalf("max_tokens = %d, want 1024", req.MaxTokens)
	}
	if req.System == "" {
		t.Fatal("system prompt is empty")
	}
	if len(req.Messages) != 1 || req.Messages[0].Role != "user" || req.Messages[0].Content != "采购螺纹钢" {
		t.Fatalf("messages = %+v, want one user message", req.Messages)
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatalf("payload is not json: %v", err)
	}
	if _, ok := raw["max_completion_tokens"]; ok {
		t.Fatal("anthropic payload must not use max_completion_tokens")
	}
}

func TestExtractModelParserContentReadsAnthropicTextContent(t *testing.T) {
	body := []byte(`{
		"content": [
			{"type": "thinking", "thinking": "hidden"},
			{"type": "text", "text": "{\"product_name\":\"螺纹钢\",\"category\":\"钢材\",\"quantity\":10,\"unit\":\"吨\",\"usage\":\"施工\",\"financing_required\":false,\"keywords\":[\"螺纹钢\"],\"confidence\":0.91}"}
		]
	}`)

	content, err := extractModelParserContent("https://api.minimaxi.com/anthropic/v1/messages", body)
	if err != nil {
		t.Fatalf("extractModelParserContent returned error: %v", err)
	}
	if content == "" || content[0] != '{' {
		t.Fatalf("content = %q, want json text", content)
	}
}

func TestParsedInstructionPayloadAcceptsStringQuantity(t *testing.T) {
	var payload parsedInstructionPayload
	err := json.Unmarshal([]byte(`{
		"product_name": "螺纹钢",
		"category": "钢材",
		"quantity": "20",
		"unit": "吨",
		"usage": "施工",
		"financing_required": false,
		"keywords": ["螺纹钢"],
		"confidence": 0.88
	}`), &payload)
	if err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if got := float64(payload.Quantity); got != 20 {
		t.Fatalf("quantity = %v, want 20", got)
	}
}

func TestParsedInstructionPayloadAcceptsStringFinancingRequired(t *testing.T) {
	var payload parsedInstructionPayload
	err := json.Unmarshal([]byte(`{
		"product_name": "热轧板卷",
		"category": "热轧板卷",
		"quantity": 100,
		"unit": "吨",
		"usage": "企业生产采购",
		"financing_required": "否",
		"keywords": ["热轧板卷"],
		"confidence": 0.88
	}`), &payload)
	if err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if bool(payload.FinancingRequired) {
		t.Fatal("financing_required = true, want false")
	}
}

func TestNormalizeParsedInstructionUsesExplicitCategoryHint(t *testing.T) {
	parsed := parsedInstructionPayload{
		ProductName:       "热轧板卷",
		Category:          "热轧板卷",
		Quantity:          quantity(100),
		Unit:              "吨",
		Usage:             "企业生产采购",
		FinancingRequired: false,
		Keywords:          []string{"热轧板卷"},
		Confidence:        0.8,
	}
	result := toParsedInstruction(parsed)
	normalizeParsedInstruction(&result, "我想买100吨的钢，品类为热轧板卷")

	if result.ProductName != "钢" {
		t.Fatalf("product_name = %q, want 钢", result.ProductName)
	}
	if result.Category != "热轧板卷" {
		t.Fatalf("category = %q, want 热轧板卷", result.Category)
	}
}

func TestModelParserRequestURLCompletesAnthropicBasePath(t *testing.T) {
	got := modelParserRequestURL("https://api.minimaxi.com/anthropic")
	want := "https://api.minimaxi.com/anthropic/v1/messages"
	if got != want {
		t.Fatalf("modelParserRequestURL = %q, want %q", got, want)
	}
}

func toParsedInstruction(payload parsedInstructionPayload) entity.ParsedInstruction {
	return entity.ParsedInstruction{
		ProductName:           payload.ProductName,
		Category:              payload.Category,
		Quantity:              float64(payload.Quantity),
		Unit:                  payload.Unit,
		Usage:                 payload.Usage,
		FinancingRequired:     bool(payload.FinancingRequired),
		Keywords:              payload.Keywords,
		Confidence:            payload.Confidence,
		DeliveryAddress:       payload.DeliveryAddress,
		DeliveryAddressSource: payload.DeliveryAddressSource,
		DeliveryWarehouseName: payload.DeliveryWarehouseName,
	}
}
