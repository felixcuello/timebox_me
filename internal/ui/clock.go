package ui

import (
	"strings"
	"time"

	lipgloss "charm.land/lipgloss/v2"
)

var glyphs = map[rune][5]string{
	'0': {
		"████",
		"█  █",
		"█  █",
		"█  █",
		"████",
	},
	'1': {
		" ██ ",
		"  █ ",
		"  █ ",
		"  █ ",
		" ███",
	},
	'2': {
		"████",
		"   █",
		"████",
		"█   ",
		"████",
	},
	'3': {
		"████",
		"   █",
		"████",
		"   █",
		"████",
	},
	'4': {
		"█  █",
		"█  █",
		"████",
		"   █",
		"   █",
	},
	'5': {
		"████",
		"█   ",
		"████",
		"   █",
		"████",
	},
	'6': {
		"████",
		"█   ",
		"████",
		"█  █",
		"████",
	},
	'7': {
		"████",
		"   █",
		"  █ ",
		" █  ",
		" █  ",
	},
	'8': {
		"████",
		"█  █",
		"████",
		"█  █",
		"████",
	},
	'9': {
		"████",
		"█  █",
		"████",
		"   █",
		"████",
	},
	':': {
		"    ",
		" ██ ",
		"    ",
		" ██ ",
		"    ",
	},
	'-': {
		"    ",
		"    ",
		"████",
		"    ",
		"    ",
	},
}

// bigClock renders hh:mm:ss (or --:--:--) as 5-row block digits.
func bigClock(text string) string {
	var rows [5]strings.Builder
	for i, r := range text {
		g, ok := glyphs[r]
		if !ok {
			g = glyphs['-']
		}
		for row := 0; row < 5; row++ {
			if i > 0 {
				rows[row].WriteByte(' ')
			}
			rows[row].WriteString(g[row])
		}
	}
	out := make([]string, 5)
	for i := range rows {
		out[i] = rows[i].String()
	}
	return strings.Join(out, "\n")
}

// placeCenter centers content in a w by h box, clipping if needed.
func placeCenter(w, h int, content string) string {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

// clockWarnRemaining is true when the big clock must be red (under one minute, including zero).
func clockWarnRemaining(remaining time.Duration) bool {
	return remaining < time.Minute
}
