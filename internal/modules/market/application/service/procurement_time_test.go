package service

import (
	"testing"
	"time"
)

func TestParseMarketTimeUsesShanghaiLocationForPlainDateTime(t *testing.T) {
	parsed, err := parseMarketTime("2026-05-31 18:30")
	if err != nil {
		t.Fatalf("parseMarketTime returned error: %v", err)
	}
	_, offset := parsed.Zone()
	if offset != 8*60*60 {
		t.Fatalf("plain datetime offset = %d, want +08:00", offset)
	}
	if parsed.Year() != 2026 || parsed.Month() != time.May || parsed.Day() != 31 || parsed.Hour() != 18 || parsed.Minute() != 30 {
		t.Fatalf("plain datetime parsed as %s", parsed.Format(time.RFC3339))
	}
}

func TestParseMarketTimeKeepsRFC3339Offset(t *testing.T) {
	parsed, err := parseMarketTime("2026-05-31T18:30:00+08:00")
	if err != nil {
		t.Fatalf("parseMarketTime returned error: %v", err)
	}
	_, offset := parsed.Zone()
	if offset != 8*60*60 {
		t.Fatalf("RFC3339 offset = %d, want +08:00", offset)
	}
}
