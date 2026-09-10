package ui

import lipgloss "charm.land/lipgloss/v2"

var (
	colorAccent = lipgloss.Color("25")
	colorMuted  = lipgloss.Color("244")
	colorFaint  = lipgloss.Color("240")
	colorWhite  = lipgloss.Color("15")
	colorActive = lipgloss.Color("46")
	colorClock  = lipgloss.Color("15")
	colorSelBg  = lipgloss.Color("236")
	colorErr    = lipgloss.Color("196")
)

type styles struct {
	paneFocused   lipgloss.Style
	paneBlurred   lipgloss.Style
	idleItem      lipgloss.Style
	activeItem    lipgloss.Style
	overtimeItem  lipgloss.Style
	doneItem      lipgloss.Style
	selectedBg    lipgloss.Style
	hint          lipgloss.Style
	overtime      lipgloss.Style
	clock         lipgloss.Style
	clockWarn     lipgloss.Style
	clockBlink    lipgloss.Style
	err           lipgloss.Style
	msgBold       lipgloss.Style
	msgItalic     lipgloss.Style
	msgStrike     lipgloss.Style
	cursor        lipgloss.Style
	deleteBox     lipgloss.Style
	confirmActive lipgloss.Style
	confirmIdle   lipgloss.Style
}

func newStyles() styles {
	pane := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	return styles{
		paneFocused:   pane.BorderForeground(colorAccent),
		paneBlurred:   pane.BorderForeground(colorFaint),
		idleItem:      lipgloss.NewStyle().Foreground(colorWhite),
		activeItem:    lipgloss.NewStyle().Foreground(colorActive).Bold(true),
		overtimeItem:  lipgloss.NewStyle().Foreground(colorErr).Bold(true),
		doneItem:      lipgloss.NewStyle().Foreground(colorMuted),
		selectedBg:    lipgloss.NewStyle().Background(colorSelBg),
		hint:          lipgloss.NewStyle().Foreground(colorMuted).Italic(true),
		overtime:      lipgloss.NewStyle().Foreground(colorMuted),
		clock:         lipgloss.NewStyle().Foreground(colorClock).Bold(true),
		clockWarn:     lipgloss.NewStyle().Foreground(colorErr).Bold(true),
		clockBlink:    lipgloss.NewStyle().Foreground(colorFaint),
		err:           lipgloss.NewStyle().Foreground(colorErr),
		msgBold:       lipgloss.NewStyle().Bold(true),
		msgItalic:     lipgloss.NewStyle().Italic(true),
		msgStrike:     lipgloss.NewStyle().Strikethrough(true),
		cursor:        lipgloss.NewStyle().Reverse(true),
		deleteBox:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorErr).Padding(1, 2),
		confirmActive: lipgloss.NewStyle().Reverse(true).Bold(true).Padding(0, 1),
		confirmIdle:   lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 1),
	}
}
