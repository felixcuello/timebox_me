package ui

import "testing"

func TestOverwriteDurationDigit(t *testing.T) {
	buf, cur, ok := overwriteDurationDigit("00:00", 0, '1')
	if !ok || buf != "10:00" || cur != 1 {
		t.Fatalf("first hour digit: buf=%q cur=%d ok=%v", buf, cur, ok)
	}
	buf, cur, ok = overwriteDurationDigit(buf, cur, '2')
	if !ok || buf != "12:00" || cur != 3 {
		t.Fatalf("second hour digit must skip colon: buf=%q cur=%d ok=%v", buf, cur, ok)
	}
	buf, cur, ok = overwriteDurationDigit(buf, cur, '4')
	if !ok || buf != "12:40" || cur != 4 {
		t.Fatalf("minutes tens: buf=%q cur=%d ok=%v", buf, cur, ok)
	}
	buf, cur, ok = overwriteDurationDigit(buf, cur, '4')
	if !ok || buf != "12:44" || cur != 4 {
		t.Fatalf("minutes ones: buf=%q cur=%d ok=%v", buf, cur, ok)
	}

	_, _, ok = overwriteDurationDigit("12:00", 3, '6')
	if ok {
		t.Fatal("minutes tens must reject 6")
	}
	_, _, ok = overwriteDurationDigit("12:00", 0, 'a')
	if ok {
		t.Fatal("non-digit must be rejected")
	}
	buf, cur, ok = overwriteDurationDigit("00:00", 0, '9')
	if !ok || buf != "90:00" || cur != 1 {
		t.Fatalf("hours must accept 0-9: buf=%q cur=%d ok=%v", buf, cur, ok)
	}
}

func TestBackspaceDuration(t *testing.T) {
	buf, cur := backspaceDuration("12:44", 4)
	if buf != "12:40" || cur != 3 {
		t.Fatalf("backspace ones: buf=%q cur=%d", buf, cur)
	}
	buf, cur = backspaceDuration(buf, cur)
	if buf != "12:00" || cur != 1 {
		t.Fatalf("backspace tens must skip colon: buf=%q cur=%d", buf, cur)
	}
	buf, cur = backspaceDuration(buf, cur)
	if buf != "10:00" || cur != 0 {
		t.Fatalf("backspace hour ones: buf=%q cur=%d", buf, cur)
	}
	buf, cur = backspaceDuration(buf, cur)
	if buf != "00:00" || cur != 0 {
		t.Fatalf("backspace at start: buf=%q cur=%d", buf, cur)
	}
}

func TestDurationCursorSkipsColon(t *testing.T) {
	if nextDurationCursor(1) != 3 {
		t.Fatal("right from hours ones must land on minutes tens")
	}
	if prevDurationCursor(3) != 1 {
		t.Fatal("left from minutes tens must land on hours ones")
	}
	if clampDurationCursor(2) != 1 {
		t.Fatal("colon cursor must snap to a digit")
	}
}
