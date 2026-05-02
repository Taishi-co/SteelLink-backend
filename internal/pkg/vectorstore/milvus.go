package vectorstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
)

type MilvusClient struct {
	endpoint   string
	database   string
	username   string
	password   string
	collection string
	httpClient *http.Client
}

type ProductHit struct {
	ProductID    uint64
	ProductNo    string
	ProductName  string
	Category     string
	EnterpriseID uint64
	Score        float64
}

type searchRequest struct {
	DBName       string      `json:"dbName,omitempty"`
	Collection   string      `json:"collectionName"`
	Data         [][]float32 `json:"data"`
	Limit        int         `json:"limit"`
	OutputFields []string    `json:"outputFields,omitempty"`
}

type searchResponse struct {
	Code    int                      `json:"code"`
	Message string                   `json:"message"`
	Data    []map[string]interface{} `json:"data"`
}

func NewMilvusClient(cfg config.VectorSearchConfig) *MilvusClient {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint != "" && !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}
	return &MilvusClient{
		endpoint:   strings.TrimRight(endpoint, "/"),
		database:   cfg.Database,
		username:   cfg.Username,
		password:   cfg.Password(),
		collection: cfg.ProductCollection,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *MilvusClient) SearchProducts(ctx context.Context, vector []float32, topK int) ([]ProductHit, error) {
	if c.endpoint == "" || c.collection == "" {
		return nil, fmt.Errorf("milvus config is incomplete")
	}
	if len(vector) == 0 {
		return nil, fmt.Errorf("search vector is empty")
	}
	if topK <= 0 {
		topK = 10
	}

	payload, err := json.Marshal(searchRequest{
		DBName:     c.database,
		Collection: c.collection,
		Data:       [][]float32{vector},
		Limit:      topK,
		OutputFields: []string{
			"product_id",
			"product_no",
			"product_name",
			"category",
			"enterprise_id",
		},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/v2/vectordb/entities/search", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)

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
		return nil, fmt.Errorf("milvus search returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed searchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("milvus search failed: %s", parsed.Message)
	}

	hits := make([]ProductHit, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		hit := ProductHit{
			ProductID:    uint64FromAny(firstNonNil(item["product_id"], item["id"])),
			ProductNo:    stringFromAny(item["product_no"]),
			ProductName:  stringFromAny(item["product_name"]),
			Category:     stringFromAny(item["category"]),
			EnterpriseID: uint64FromAny(item["enterprise_id"]),
			Score:        floatFromAny(firstNonNil(item["score"], item["distance"])),
		}
		if hit.ProductID == 0 {
			continue
		}
		hits = append(hits, hit)
	}
	return hits, nil
}

func (c *MilvusClient) setAuth(req *http.Request) {
	switch {
	case c.username != "" && c.password != "":
		req.SetBasicAuth(c.username, c.password)
	case c.password != "":
		req.Header.Set("Authorization", "Bearer "+c.password)
	}
}

func (c *MilvusClient) EndpointHost() string {
	if c.endpoint == "" {
		return ""
	}
	parsed, err := url.Parse(c.endpoint)
	if err != nil {
		return c.endpoint
	}
	return parsed.Host
}

func firstNonNil(values ...interface{}) interface{} {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func stringFromAny(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}

func uint64FromAny(value interface{}) uint64 {
	switch v := value.(type) {
	case float64:
		if v > 0 {
			return uint64(v)
		}
	case int:
		if v > 0 {
			return uint64(v)
		}
	case int64:
		if v > 0 {
			return uint64(v)
		}
	case uint64:
		return v
	case json.Number:
		if n, err := strconv.ParseUint(v.String(), 10, 64); err == nil {
			return n
		}
	case string:
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func floatFromAny(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		n, _ := strconv.ParseFloat(v.String(), 64)
		return n
	case string:
		n, _ := strconv.ParseFloat(v, 64)
		return n
	default:
		return 0
	}
}
