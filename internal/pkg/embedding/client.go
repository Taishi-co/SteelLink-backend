package embedding

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
)

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

type request struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type response struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func NewClient(cfg config.VectorSearchConfig) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.EmbeddingBaseURL, "/"),
		apiKey:  cfg.EmbeddingAPIKey(),
		model:   cfg.EmbeddingModel,
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

	payload, err := json.Marshal(request{Model: c.model, Input: []string{text}})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
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
	if len(parsed.Data) == 0 || len(parsed.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("embedding service returned empty embedding")
	}
	return parsed.Data[0].Embedding, nil
}
