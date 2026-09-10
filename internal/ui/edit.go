package ui

import (
	"unicode/utf8"

	lipgloss "charm.land/lipgloss/v2"
)

func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}

func runesOf(s string) []rune {
	return []rune(s)
}

func clampCursor(n, cur int) int {
	if cur < 0 {
		return 0
	}
	if cur > n {
		return n
	}
	return cur
}

// insertRune inserts r at cursor and returns the new string and cursor.
func insertRune(s string, cur int, r rune) (string, int) {
	rs := runesOf(s)
	cur = clampCursor(len(rs), cur)
	out := make([]rune, 0, len(rs)+1)
	out = append(out, rs[:cur]...)
	out = append(out, r)
	out = append(out, rs[cur:]...)
	return string(out), cur + 1
}

// deleteBefore removes the rune before cursor.
func deleteBefore(s string, cur int) (string, int) {
	rs := runesOf(s)
	cur = clampCursor(len(rs), cur)
	if cur == 0 {
		return s, 0
	}
	out := append([]rune{}, rs[:cur-1]...)
	out = append(out, rs[cur:]...)
	return string(out), cur - 1
}

// paintCursor inverts the glyph at cur, or appends a block when cur is at the end.
func paintCursor(s string, cur int, cursorStyle lipgloss.Style) string {
	rs := runesOf(s)
	cur = clampCursor(len(rs), cur)
	if len(rs) == 0 {
		return cursorStyle.Render(" ")
	}
	if cur == len(rs) {
		return s + cursorStyle.Render(" ")
	}
	return string(rs[:cur]) + cursorStyle.Render(string(rs[cur])) + string(rs[cur+1:])
}
