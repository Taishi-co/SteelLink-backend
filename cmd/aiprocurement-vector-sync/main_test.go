package main

import (
	"testing"
	"time"
)

func TestParseScheduleDuration(t *testing.T) {
	interval, once, err := parseSchedule("30m")
	if err != nil {
		t.Fatalf("parseSchedule returned error: %v", err)
	}
	if once {
		t.Fatal("parseSchedule returned once=true, want false")
	}
	if interval != 30*time.Minute {
		t.Fatalf("interval = %s, want 30m", interval)
	}
}

func TestParseScheduleManual(t *testing.T) {
	interval, once, err := parseSchedule("manual")
	if err != nil {
		t.Fatalf("parseSchedule returned error: %v", err)
	}
	if !once {
		t.Fatal("parseSchedule returned once=false, want true")
	}
	if interval != 0 {
		t.Fatalf("interval = %s, want 0", interval)
	}
}
