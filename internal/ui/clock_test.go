package ui

import (
	"strings"
	"testing"
	"time"
)

func TestBigClockHasFiveRows(t *testing.T) {
	out := bigClock("01:30")
	lines := strings.Split(out, "\n")
	if len(lines) != 5 {
		t.Fatalf("rows=%d, want 5\n%s", len(lines), out)
	}
	if bigClock("--:--:--") == "" {
		t.Fatal("placeholder clock must not be empty")
	}
	if lines := strings.Split(bigClock("01:30:05"), "\n"); len(lines) != 5 {
		t.Fatalf("hh:mm:ss rows=%d, want 5", len(lines))
	}
}

func TestClockWarnRemaining(t *testing.T) {
	tests := []struct {
		in   time.Duration
		warn bool
	}{
		{time.Hour + 30*time.Minute, false},
		{time.Minute, false},
		{time.Minute - time.Second, true},
		{30 * time.Second, true},
		{0, true},
		{-time.Second, true},
	}
	for _, tc := range tests {
		got := clockWarnRemaining(tc.in)
		if got != tc.warn {
			t.Errorf("clockWarnRemaining(%v) = %v, want %v", tc.in, got, tc.warn)
		}
	}
}
