package md

import (
	"strings"

	lipgloss "charm.land/lipgloss/v2"
)

// Render applies WhatsApp-style *bold*, _italic_, and ~strike~ markers.
// Runs do not span newlines. Unmatched markers stay literal.
func Render(s string, boldStyle, italicStyle, strikeStyle lipgloss.Style) string {
	if s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = renderLine(line, false, false, false, boldStyle, italicStyle, strikeStyle)
	}
	return strings.Join(lines, "\n")
}

func renderLine(s string, bold, italic, strike bool, boldStyle, italicStyle, strikeStyle lipgloss.Style) string {
	if s == "" {
		return styleRun(s, bold, italic, strike, boldStyle, italicStyle, strikeStyle)
	}
	var b strings.Builder
	i := 0
	for i < len(s) {
		c := s[i]
		if isMarker(c) {
			if close := findClosing(s, c, i+1); close >= 0 {
				inner := s[i+1 : close]
				nb, ni, ns := bold, italic, strike
				switch c {
				case '*':
					nb = true
				case '_':
					ni = true
				case '~':
					ns = true
				}
				b.WriteString(renderLine(inner, nb, ni, ns, boldStyle, italicStyle, strikeStyle))
				i = close + 1
				continue
			}
		}
		j := i + 1
		for j < len(s) && !isMarker(s[j]) {
			j++
		}
		b.WriteString(styleRun(s[i:j], bold, italic, strike, boldStyle, italicStyle, strikeStyle))
		i = j
	}
	return b.String()
}

func isMarker(c byte) bool {
	return c == '*' || c == '_' || c == '~'
}

// findClosing finds the closing marker after start. Empty pairs do not match.
func findClosing(s string, open byte, start int) int {
	for i := start; i < len(s); i++ {
		c := s[i]
		if c == open {
			if i > start {
				return i
			}
			continue
		}
		if isMarker(c) && c != open {
			if close := findClosing(s, c, i+1); close >= 0 {
				i = close
			}
		}
	}
	return -1
}

func styleRun(s string, bold, italic, strike bool, boldStyle, italicStyle, strikeStyle lipgloss.Style) string {
	if s == "" {
		return s
	}
	if !bold && !italic && !strike {
		return s
	}
	st := lipgloss.NewStyle()
	if bold {
		st = st.Bold(true)
	}
	if italic {
		st = st.Italic(true)
	}
	if strike {
		st = st.Strikethrough(true)
	}
	switch {
	case bold && !italic && !strike:
		return boldStyle.Render(s)
	case italic && !bold && !strike:
		return italicStyle.Render(s)
	case strike && !bold && !italic:
		return strikeStyle.Render(s)
	default:
		return st.Render(s)
	}
}
