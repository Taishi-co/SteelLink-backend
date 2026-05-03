package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/internal/modules/aiprocurement/domain/entity"
	"go.uber.org/zap"
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model               string        `json:"model"`
	Messages            []chatMessage `json:"messages"`
	Temperature         float64       `json:"temperature,omitempty"`
	TopP                float64       `json:"top_p,omitempty"`
	MaxCompletionTokens int           `json:"max_completion_tokens,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type anthropicMessageRequest struct {
	Model       string        `json:"model"`
	System      string        `json:"system,omitempty"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
}

type anthropicMessageResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type parsedInstructionPayload struct {
	ProductName       string   `json:"product_name"`
	Category          string   `json:"category"`
	Quantity          float64  `json:"quantity"`
	Unit              string   `json:"unit"`
	Usage             string   `json:"usage"`
	FinancingRequired bool     `json:"financing_required"`
	Keywords          []string `json:"keywords"`
	Confidence        float64  `json:"confidence"`
}

func parseInstructionWithModel(ctx context.Context, prompt string) (entity.ParsedInstruction, bool, error) {
	modelCfg := config.GetAIProcurement()
	if !modelCfg.ModelParsingEnabled {
		logAIProcurementInfo("AI procurement model parser skipped",
			zap.String("reason", "model_parsing_disabled"),
		)
		return entity.ParsedInstruction{}, false, nil
	}

	apiKey := modelCfg.APIKey()
	if apiKey == "" {
		logAIProcurementWarn("AI procurement model parser skipped",
			zap.String("reason", "api_key_empty"),
			zap.String("provider", modelCfg.Provider),
			zap.String("api_key_env", modelCfg.APIKeyEnv),
			zap.String("model", modelCfg.Model),
		)
		return entity.ParsedInstruction{}, false, nil
	}
	if modelCfg.BaseURL == "" || modelCfg.Model == "" {
		logAIProcurementWarn("AI procurement model parser skipped",
			zap.String("reason", "model_config_incomplete"),
			zap.String("provider", modelCfg.Provider),
			zap.String("base_url", modelCfg.BaseURL),
			zap.String("model", modelCfg.Model),
		)
		return entity.ParsedInstruction{}, false, nil
	}
	requestURL := modelParserRequestURL(modelCfg.BaseURL)

	timeout := time.Duration(modelCfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	startedAt := time.Now()
	logAIProcurementInfo("AI procurement model parser started",
		zap.String("provider", modelCfg.Provider),
		zap.String("base_url", requestURL),
		zap.String("model", modelCfg.Model),
		zap.Int("timeout_sec", int(timeout.Seconds())),
		zap.Int("prompt_len", len([]rune(prompt))),
	)
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	payload, err := buildModelParserPayload(modelCfg, prompt)
	if err != nil {
		return entity.ParsedInstruction{}, true, err
	}

	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, requestURL, bytes.NewReader(payload))
	if err != nil {
		return entity.ParsedInstruction{}, true, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logAIProcurementWarn("AI procurement model parser failed",
			zap.String("provider", modelCfg.Provider),
			zap.String("model", modelCfg.Model),
			zap.Duration("latency", time.Since(startedAt)),
			zap.Error(err),
		)
		return entity.ParsedInstruction{}, true, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return entity.ParsedInstruction{}, true, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("model parser returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		logAIProcurementWarn("AI procurement model parser failed",
			zap.String("provider", modelCfg.Provider),
			zap.String("model", modelCfg.Model),
			zap.Int("status", resp.StatusCode),
			zap.Duration("latency", time.Since(startedAt)),
			zap.Error(err),
		)
		return entity.ParsedInstruction{}, true, err
	}

	content, err := extractModelParserContent(requestURL, body)
	if err != nil {
		logAIProcurementWarn("AI procurement model parser failed",
			zap.String("provider", modelCfg.Provider),
			zap.String("model", modelCfg.Model),
			zap.Int("status", resp.StatusCode),
			zap.Duration("latency", time.Since(startedAt)),
			zap.Error(err),
		)
		return entity.ParsedInstruction{}, true, err
	}

	content = normalizeJSONContent(content)
	var parsed parsedInstructionPayload
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		logAIProcurementWarn("AI procurement model parser failed",
			zap.String("provider", modelCfg.Provider),
			zap.String("model", modelCfg.Model),
			zap.Int("status", resp.StatusCode),
			zap.Duration("latency", time.Since(startedAt)),
			zap.Error(err),
		)
		return entity.ParsedInstruction{}, true, err
	}

	result := entity.ParsedInstruction{
		ProductName:       strings.TrimSpace(parsed.ProductName),
		Category:          strings.TrimSpace(parsed.Category),
		Quantity:          parsed.Quantity,
		Unit:              strings.TrimSpace(parsed.Unit),
		Usage:             strings.TrimSpace(parsed.Usage),
		FinancingRequired: parsed.FinancingRequired,
		Keywords:          compactStrings(parsed.Keywords),
		Confidence:        parsed.Confidence,
	}
	normalizeParsedInstruction(&result, prompt)
	logAIProcurementInfo("AI procurement model parser succeeded",
		zap.String("provider", modelCfg.Provider),
		zap.String("model", modelCfg.Model),
		zap.Int("status", resp.StatusCode),
		zap.Duration("latency", time.Since(startedAt)),
		zap.String("product_name", result.ProductName),
		zap.String("category", result.Category),
		zap.Float64("confidence", result.Confidence),
	)
	return result, true, nil
}

func buildModelParserPayload(modelCfg config.AIProcurementConfig, prompt string) ([]byte, error) {
	systemPrompt := modelParserSystemPrompt()
	if isAnthropicMessagesEndpoint(modelCfg.BaseURL) {
		maxTokens := modelCfg.MaxTokens
		if maxTokens <= 0 {
			maxTokens = 1200
		}
		return json.Marshal(anthropicMessageRequest{
			Model:       modelCfg.Model,
			System:      systemPrompt,
			Messages:    []chatMessage{{Role: "user", Content: prompt}},
			MaxTokens:   maxTokens,
			Temperature: modelCfg.Temperature,
			TopP:        modelCfg.TopP,
		})
	}
	return json.Marshal(chatCompletionRequest{
		Model: modelCfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature:         modelCfg.Temperature,
		TopP:                modelCfg.TopP,
		MaxCompletionTokens: modelCfg.MaxTokens,
	})
}

func extractModelParserContent(baseURL string, body []byte) (string, error) {
	if isAnthropicMessagesEndpoint(baseURL) {
		var completion anthropicMessageResponse
		if err := json.Unmarshal(body, &completion); err != nil {
			return "", err
		}
		for _, item := range completion.Content {
			if item.Type == "text" && strings.TrimSpace(item.Text) != "" {
				return item.Text, nil
			}
		}
		return "", fmt.Errorf("model parser returned no text content")
	}
	var completion chatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return "", err
	}
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("model parser returned no choices")
	}
	return completion.Choices[0].Message.Content, nil
}

func isAnthropicMessagesEndpoint(baseURL string) bool {
	normalized := strings.ToLower(strings.TrimSpace(baseURL))
	return strings.Contains(normalized, "/anthropic")
}

func modelParserRequestURL(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(strings.ToLower(trimmed), "/anthropic") {
		return trimmed + "/v1/messages"
	}
	return trimmed
}

func modelParserSystemPrompt() string {
	return strings.Join([]string{
		"你是钢材B2B平台的采购指令解析器。",
		"只输出JSON，不要输出解释、Markdown或代码块。",
		"JSON字段必须包含 product_name, category, quantity, unit, usage, financing_required, keywords, confidence。",
		"category 使用钢材或工业品类短名称；quantity 是数字；confidence 取值0到1。",
	}, "\n")
}

func normalizeJSONContent(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizeParsedInstruction(parsed *entity.ParsedInstruction, prompt string) {
	fallback := parseInstructionByRule(prompt)
	if parsed.ProductName == "" {
		parsed.ProductName = fallback.ProductName
	}
	if parsed.Category == "" {
		parsed.Category = fallback.Category
	}
	if parsed.Quantity <= 0 {
		parsed.Quantity = fallback.Quantity
	}
	if parsed.Unit == "" {
		parsed.Unit = fallback.Unit
	}
	if parsed.Usage == "" {
		parsed.Usage = fallback.Usage
	}
	if len(parsed.Keywords) == 0 {
		parsed.Keywords = fallback.Keywords
	}
	if parsed.Confidence <= 0 || parsed.Confidence > 1 {
		parsed.Confidence = fallback.Confidence
	}
}
