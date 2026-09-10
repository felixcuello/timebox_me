package md

import (
	"regexp"
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"
)

var ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiEscapeRe.ReplaceAllString(s, "")
}

func formatTest(in string) string {
	bold := lipgloss.NewStyle().Bold(true)
	italic := lipgloss.NewStyle().Italic(true)
	strike := lipgloss.NewStyle().Strikethrough(true)
	return Render(in, bold, italic, strike)
}

func TestRender(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "hello", want: "hello"},
		{name: "bold", in: "hello *world*", want: "hello world"},
		{name: "italic", in: "hello _world_", want: "hello world"},
		{name: "strike", in: "hello ~world~", want: "hello world"},
		{name: "nested bold italic", in: "*_both_*", want: "both"},
		{name: "nested italic bold", in: "_*both*_", want: "both"},
		{name: "nested bold strike", in: "*~x~*", want: "x"},
		{name: "nested strike bold", in: "~*x*~", want: "x"},
		{name: "multi segment", in: "*a* and _b_ and ~c~", want: "a and b and c"},
		{name: "unmatched star", in: "a * b", want: "a * b"},
		{name: "unmatched underscore", in: "a _ b", want: "a _ b"},
		{name: "unmatched tilde", in: "a ~ b", want: "a ~ b"},
		{name: "empty pair star", in: "x**y", want: "x**y"},
		{name: "empty pair tilde", in: "x~~y", want: "x~~y"},
		{name: "newline no span", in: "*a\nb*", want: "*a\nb*"},
		{name: "strike newline no span", in: "~a\nb~", want: "~a\nb~"},
		{name: "per line", in: "*a*\n_b_\n~c~", want: "a\nb\nc"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := stripANSI(formatTest(tc.in))
			if got != tc.want {
				t.Fatalf("visible %q, want %q (raw %q)", got, tc.want, formatTest(tc.in))
			}
		})
	}
}

func TestRenderAppliesStyles(t *testing.T) {
	out := formatTest("*bold*")
	if out == "*bold*" || out == "bold" {
		t.Fatalf("expected styled output, got %q", out)
	}
	if !strings.Contains(stripANSI(out), "bold") {
		t.Fatalf("missing text: %q", out)
	}

	out = formatTest("_ital_")
	if out == "_ital_" || out == "ital" {
		t.Fatalf("expected styled italic output, got %q", out)
	}

	out = formatTest("~gone~")
	if out == "~gone~" || out == "gone" {
		t.Fatalf("expected styled strike output, got %q", out)
	}
	if !strings.Contains(stripANSI(out), "gone") {
		t.Fatalf("missing text: %q", out)
	}
}
