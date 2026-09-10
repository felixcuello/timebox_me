package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FormatHHMM renders d as hours and minutes, floored, zero-padded (01:30).
func FormatHHMM(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalMin := int64(d / time.Minute)
	h := totalMin / 60
	m := totalMin % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

// FormatHHMMSS renders d as hours, minutes, and seconds, floored, zero-padded (01:30:05).
func FormatHHMMSS(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSec := int64(d / time.Second)
	h := totalSec / 3600
	m := (totalSec % 3600) / 60
	sec := totalSec % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, sec)
}

// ParseHHMM parses a duration like 1:30 or 01:30 as hours and minutes.
// Minutes must be in 0-59. Invalid input returns an error; the caller must keep the previous value.
func ParseHHMM(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	h, errH := strconv.Atoi(parts[0])
	m, errM := strconv.Atoi(parts[1])
	if errH != nil || errM != nil || h < 0 || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute, nil
}
