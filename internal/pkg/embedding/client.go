package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
)

type Client struct {
	provider   string
	baseURL    string
	apiKey     string
	groupID    string
	model      string
	httpClient *http.Client
}

type openAIRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type minimaxRequest struct {
	Model string   `json:"model"`
	Texts []string `json:"texts"`
	Type  string   `json:"type"`
}

type response struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	BaseResp struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
	Vectors json.RawMessage `json:"vectors"`
}

func NewClient(cfg config.VectorSearchConfig) *Client {
	return &Client{
		provider: strings.ToLower(strings.TrimSpace(cfg.EmbeddingProvider)),
		baseURL:  strings.TrimRight(cfg.EmbeddingBaseURL, "/"),
		apiKey:   cfg.EmbeddingAPIKey(),
		groupID:  os.Getenv("MINIMAX_GROUP_ID"),
		model:    cfg.EmbeddingModel,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("embedding text is empty")
	}
	if c.baseURL == "" || c.model == "" {
		return nil, fmt.Errorf("embedding config is incomplete")
	}
	if c.apiKey == "" {
		return nil, fmt.Errorf("embedding api key is empty")
	}

	endpoint, payload, err := c.buildRequest(text)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding service returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("embedding service returned status_code=%d status_msg=%s", parsed.BaseResp.StatusCode, parsed.BaseResp.StatusMsg)
	}
	if vector := parseVectors(parsed.Vectors); len(vector) > 0 {
		return vector, nil
	}
	if len(parsed.Data) > 0 && len(parsed.Data[0].Embedding) > 0 {
		return parsed.Data[0].Embedding, nil
	}
	return nil, fmt.Errorf("embedding service returned empty embedding: %s", trimBody(body))
}

func (c *Client) buildRequest(text string) (string, []byte, error) {
	if c.provider == "minimax" {
		payload, err := json.Marshal(minimaxRequest{
			Model: c.model,
			Texts: []string{text},
			Type:  "db",
		})
		if err != nil {
			return "", nil, err
		}
		return c.minimaxEndpoint(), payload, nil
	}
	payload, err := json.Marshal(openAIRequest{Model: c.model, Input: []string{text}})
	if err != nil {
		return "", nil, err
	}
	return c.baseURL, payload, nil
}

func (c *Client) minimaxEndpoint() string {
	if c.groupID == "" {
		return c.baseURL
	}
	parsed, err := url.Parse(c.baseURL)
	if err != nil {
		return c.baseURL
	}
	query := parsed.Query()
	if query.Get("group_id") == "" {
		query.Set("group_id", c.groupID)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func parseVectors(raw json.RawMessage) []float32 {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var matrix [][]float32
	if err := json.Unmarshal(raw, &matrix); err == nil && len(matrix) > 0 && len(matrix[0]) > 0 {
		return matrix[0]
	}
	var objects []struct {
		Embedding []float32 `json:"embedding"`
		Vector    []float32 `json:"vector"`
	}
	if err := json.Unmarshal(raw, &objects); err == nil && len(objects) > 0 {
		if len(objects[0].Embedding) > 0 {
			return objects[0].Embedding
		}
		if len(objects[0].Vector) > 0 {
			return objects[0].Vector
		}
	}
	var vector []float32
	if err := json.Unmarshal(raw, &vector); err == nil && len(vector) > 0 {
		return vector
	}
	return nil
}

func trimBody(body []byte) string {
	value := strings.TrimSpace(string(body))
	if len(value) <= 500 {
		return value
	}
	return value[:500] + "..."
}
