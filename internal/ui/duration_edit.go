package ui

import (
	"time"

	"timebox_me/internal/domain"
)

const (
	durationMaskLen = 5
	durationColonAt = 2
	minutesTensAt   = 3
)

// durationMaskFromRemaining loads remaining into a fixed hh:mm editor mask (hours 00-99).
func durationMaskFromRemaining(d time.Duration) string {
	maxMask := 99*time.Hour + 59*time.Minute
	if d > maxMask {
		d = maxMask
	}
	s := domain.FormatHHMM(d)
	rs := []rune(s)
	if len(rs) == durationMaskLen && rs[durationColonAt] == ':' {
		return s
	}
	return "00:00"
}

// nextDurationCursor moves right in hh:mm, skipping the colon.
func nextDurationCursor(cur int) int {
	cur = clampDurationCursor(cur)
	switch cur {
	case 0:
		return 1
	case 1:
		return 3
	case 3:
		return 4
	default:
		return 4
	}
}

// prevDurationCursor moves left in hh:mm, skipping the colon.
func prevDurationCursor(cur int) int {
	cur = clampDurationCursor(cur)
	switch cur {
	case 4:
		return 3
	case 3:
		return 1
	case 1:
		return 0
	default:
		return 0
	}
}

// clampDurationCursor keeps the editor cursor on a digit, never on the colon.
func clampDurationCursor(cur int) int {
	if cur <= 0 {
		return 0
	}
	if cur == durationColonAt {
		return 1
	}
	if cur >= 4 {
		return 4
	}
	return cur
}

// isDurationDigitSlot reports whether cur is one of the four hh:mm digit indexes.
func isDurationDigitSlot(cur int) bool {
	return cur == 0 || cur == 1 || cur == 3 || cur == 4
}

// overwriteDurationDigit writes one digit into the hh:mm mask.
// Minutes tens (index 3) only accept 0-5. The colon is never overwritten.
func overwriteDurationDigit(buf string, cur int, digit rune) (string, int, bool) {
	if digit < '0' || digit > '9' {
		return buf, cur, false
	}
	rs := []rune(buf)
	if len(rs) != durationMaskLen || rs[durationColonAt] != ':' {
		rs = []rune("00:00")
	}
	cur = clampDurationCursor(cur)
	if !isDurationDigitSlot(cur) {
		return string(rs), cur, false
	}
	if cur == minutesTensAt && digit > '5' {
		return string(rs), cur, false
	}
	rs[cur] = digit
	return string(rs), nextDurationCursor(cur), true
}

// backspaceDuration sets the current digit to 0 and moves left.
func backspaceDuration(buf string, cur int) (string, int) {
	rs := []rune(buf)
	if len(rs) != durationMaskLen || rs[durationColonAt] != ':' {
		return "00:00", 0
	}
	cur = clampDurationCursor(cur)
	if isDurationDigitSlot(cur) {
		rs[cur] = '0'
	}
	return string(rs), prevDurationCursor(cur)
}
