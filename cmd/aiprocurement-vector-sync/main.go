package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type syncOptions struct {
	endpoint     string
	token        string
	tokenEnv     string
	target       string
	schedule     string
	batchSize    int
	limit        int
	createSchema bool
	timeout      time.Duration
}

type syncRequest struct {
	Target       string `json:"target"`
	BatchSize    int    `json:"batch_size"`
	Limit        int    `json:"limit"`
	CreateSchema bool   `json:"create_schema"`
}

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func main() {
	opts := syncOptions{}
	flag.StringVar(&opts.endpoint, "endpoint", "", "backend vector sync endpoint, defaults to AI_PROCUREMENT_VECTOR_SYNC_ENDPOINT")
	flag.StringVar(&opts.token, "token", "", "backend vector sync token; prefer token env in production")
	flag.StringVar(&opts.tokenEnv, "token-env", "AI_PROCUREMENT_VECTOR_SYNC_TOKEN", "environment variable that stores backend vector sync token")
	flag.StringVar(&opts.target, "target", "products", "sync target: products, suppliers, finance, all")
	flag.StringVar(&opts.schedule, "schedule", "manual", "schedule: manual, minute, hour, day")
	flag.IntVar(&opts.batchSize, "batch-size", 100, "products per backend batch")
	flag.IntVar(&opts.limit, "limit", 0, "max products to sync, 0 means no limit")
	flag.BoolVar(&opts.createSchema, "create-schema", true, "ask backend to create pgvector extension/table/indexes when possible")
	flag.DurationVar(&opts.timeout, "timeout", 30*time.Minute, "HTTP request timeout")
	flag.Parse()

	opts.endpoint = resolveEndpoint(opts.endpoint)
	if opts.token == "" && opts.tokenEnv != "" {
		opts.token = os.Getenv(opts.tokenEnv)
	}
	if opts.endpoint == "" {
		fmt.Fprintln(os.Stderr, "backend vector sync endpoint is required; set --endpoint or AI_PROCUREMENT_VECTOR_SYNC_ENDPOINT")
		os.Exit(1)
	}
	if opts.token == "" {
		fmt.Fprintf(os.Stderr, "backend vector sync token is required; set --token or %s\n", opts.tokenEnv)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := runScheduled(ctx, opts); err != nil {
		fmt.Fprintf(os.Stderr, "vector sync failed: %v\n", err)
		os.Exit(1)
	}
}

func resolveEndpoint(value string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	if envValue := strings.TrimSpace(os.Getenv("AI_PROCUREMENT_VECTOR_SYNC_ENDPOINT")); envValue != "" {
		return envValue
	}
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("STEELLINK_BACKEND_BASE_URL")), "/")
	if baseURL == "" {
		return ""
	}
	return baseURL + "/api/v1/internal/ai-procurements/vector-sync"
}

func runScheduled(ctx context.Context, opts syncOptions) error {
	interval, once, err := parseSchedule(opts.schedule)
	if err != nil {
		return err
	}
	for {
		startedAt := time.Now()
		if err := callBackend(ctx, opts); err != nil {
			return err
		}
		fmt.Printf("vector sync request finished schedule=%s target=%s duration=%s\n", opts.schedule, opts.target, time.Since(startedAt).Round(time.Millisecond))
		if once {
			return nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}

func parseSchedule(value string) (time.Duration, bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "manual", "once":
		return 0, true, nil
	case "minute", "minutely":
		return time.Minute, false, nil
	case "hour", "hourly":
		return time.Hour, false, nil
	case "day", "daily":
		return 24 * time.Hour, false, nil
	default:
		return 0, false, fmt.Errorf("unsupported schedule %q", value)
	}
}

func callBackend(ctx context.Context, opts syncOptions) error {
	body, err := json.Marshal(syncRequest{
		Target:       opts.target,
		BatchSize:    opts.batchSize,
		Limit:        opts.limit,
		CreateSchema: opts.createSchema,
	})
	if err != nil {
		return err
	}
	reqCtx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, opts.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AI-Procurement-Sync-Token", opts.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("backend returned %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	var parsed apiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return fmt.Errorf("decode backend response failed: %w", err)
	}
	if parsed.Code != 0 {
		return fmt.Errorf("backend returned code=%d message=%s", parsed.Code, parsed.Message)
	}
	fmt.Printf("backend response: %s\n", strings.TrimSpace(string(parsed.Data)))
	return nil
}
