package domain

import (
	"testing"
	"time"
)

func TestFormatHHMM(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{0, "00:00"},
		{time.Minute, "00:01"},
		{90 * time.Minute, "01:30"},
		{time.Hour + 30*time.Minute, "01:30"},
		{59 * time.Second, "00:00"},
		{-time.Minute, "00:00"},
		{2*time.Hour + 5*time.Minute + 30*time.Second, "02:05"},
	}
	for _, tc := range tests {
		got := FormatHHMM(tc.in)
		if got != tc.want {
			t.Errorf("FormatHHMM(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatHHMMSS(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{0, "00:00:00"},
		{59 * time.Second, "00:00:59"},
		{time.Minute, "00:01:00"},
		{time.Hour + 30*time.Minute + 5*time.Second, "01:30:05"},
		{-time.Second, "00:00:00"},
	}
	for _, tc := range tests {
		got := FormatHHMMSS(tc.in)
		if got != tc.want {
			t.Errorf("FormatHHMMSS(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseHHMM(t *testing.T) {
	ok := []struct {
		in   string
		want time.Duration
	}{
		{"01:30", time.Hour + 30*time.Minute},
		{"1:30", time.Hour + 30*time.Minute},
		{"00:00", 0},
		{"  02:05  ", 2*time.Hour + 5*time.Minute},
	}
	for _, tc := range ok {
		got, err := ParseHHMM(tc.in)
		if err != nil {
			t.Errorf("ParseHHMM(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseHHMM(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}

	bad := []string{"", "abc", "1", "1:60", "-1:00", "1:2:3", "01:30x"}
	for _, in := range bad {
		if _, err := ParseHHMM(in); err == nil {
			t.Errorf("ParseHHMM(%q) must fail", in)
		}
	}
}

func TestParseHHMMRejectKeepsCallerValue(t *testing.T) {
	prev := 45 * time.Minute
	d, err := ParseHHMM("99:99")
	if err == nil {
		t.Fatal("expected error")
	}
	if d != 0 {
		t.Fatalf("on error duration must be zero value, got %v", d)
	}
	if prev != 45*time.Minute {
		t.Fatal("caller previous value must stay unchanged")
	}
}
