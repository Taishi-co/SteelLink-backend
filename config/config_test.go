package config

import "testing"

func TestApplyAdminAIProcurementConfigVisionFields(t *testing.T) {
	target := AIProcurementConfig{
		RuntimeConfig: RuntimeConfig{
			Enabled:     true,
			RedisKey:    "custom_ai_procurement_config",
			CacheTTLSec: 12,
		},
	}
	source := aiProcurementAdminConfig{Enabled: true}
	source.Model.ParsingEnabled = true
	source.Model.Provider = "minimax"
	source.Model.BaseURL = "https://api.minimax.io/v1/chat/completions/"
	source.Model.APIKeyEnv = "MINIMAX_API_KEY"
	source.Model.VisionProvider = "minimax_token_plan"
	source.Model.VisionBaseURL = "https://api.minimax.io/v1/coding_plan/vlm/"
	source.Model.VisionAPIKeyEnv = "MINIMAX_TOKEN_PLAN_API_KEY"
	source.Model.Model = "MiniMax-M2.7"

	applyAdminAIProcurementConfig(&target, source)

	if !target.ModelParsingEnabled {
		t.Fatalf("ModelParsingEnabled = false")
	}
	if target.VisionProvider != "minimax_token_plan" {
		t.Fatalf("VisionProvider = %q", target.VisionProvider)
	}
	if target.VisionBaseURL != "https://api.minimax.io/v1/coding_plan/vlm" {
		t.Fatalf("VisionBaseURL = %q", target.VisionBaseURL)
	}
	if target.VisionAPIKeyEnv != "MINIMAX_TOKEN_PLAN_API_KEY" {
		t.Fatalf("VisionAPIKeyEnv = %q", target.VisionAPIKeyEnv)
	}
	if target.RuntimeConfig.RedisKey != "custom_ai_procurement_config" || target.RuntimeConfig.CacheTTLSec != 12 {
		t.Fatalf("RuntimeConfig was not preserved: %#v", target.RuntimeConfig)
	}
}

func TestApplyAdminAIProcurementConfigDefaultsVisionAPIKeyEnv(t *testing.T) {
	target := AIProcurementConfig{}
	source := aiProcurementAdminConfig{Enabled: true}
	source.Model.ParsingEnabled = true

	applyAdminAIProcurementConfig(&target, source)

	if target.VisionProvider != "minimax_token_plan" {
		t.Fatalf("VisionProvider = %q", target.VisionProvider)
	}
	if target.VisionBaseURL != "https://api.minimax.io/v1/coding_plan/vlm" {
		t.Fatalf("VisionBaseURL = %q", target.VisionBaseURL)
	}
	if target.VisionAPIKeyEnv != "MINIMAX_TOKEN_PLAN_API_KEY" {
		t.Fatalf("VisionAPIKeyEnv = %q", target.VisionAPIKeyEnv)
	}
}
