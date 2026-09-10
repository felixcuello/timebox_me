package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"timebox_me/internal/domain"
	"timebox_me/internal/md"
)

// View renders the two-pane layout: item list, big clock, notes.
func (m Model) View() tea.View {
	if !m.ready {
		return m.wrapView("Starting timebox...")
	}

	leftInnerH := m.height - 2
	if leftInnerH < 1 {
		leftInnerH = 1
	}
	leftInnerW := m.listWidth - 2
	if leftInnerW < 1 {
		leftInnerW = 1
	}

	leftStyle := m.styles.paneBlurred
	if m.focus == focusList {
		leftStyle = m.styles.paneFocused
	}
	left := leftStyle.Width(m.listWidth).Height(m.height).Render(m.renderList(leftInnerW, leftInnerH))

	rightW := m.width - m.listWidth
	if rightW < 1 {
		rightW = 1
	}
	clockInnerH := m.clockHeight - 2
	if clockInnerH < 1 {
		clockInnerH = 1
	}
	clockInnerW := rightW - 2
	if clockInnerW < 1 {
		clockInnerW = 1
	}

	clockStyle := m.styles.paneBlurred
	clock := clockStyle.Width(rightW).Height(m.clockHeight).Render(m.renderClock(clockInnerW, clockInnerH))

	notesOuterH := m.height - m.clockHeight
	if notesOuterH < 1 {
		notesOuterH = 1
	}
	notesStyle := m.styles.paneBlurred
	if m.focus == focusNotes {
		notesStyle = m.styles.paneFocused
	}
	notesInnerW := rightW - 2
	notesInnerH := notesOuterH - 2
	if notesInnerW < 1 {
		notesInnerW = 1
	}
	if notesInnerH < 1 {
		notesInnerH = 1
	}
	notes := notesStyle.Width(rightW).Height(notesOuterH).Render(m.renderNotes(notesInnerW, notesInnerH))

	right := lipgloss.JoinVertical(lipgloss.Left, clock, notes)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	if m.persistErr != "" {
		errLine := m.styles.err.Render(m.persistErr)
		body = lipgloss.JoinVertical(lipgloss.Left, body, errLine)
	}
	if m.deleteConfirm {
		body = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.renderDeleteModal())
	}
	return m.wrapView(body)
}

func (m Model) wrapView(content string) tea.View {
	v := tea.NewView(content)
	v.AltScreen = true
	v.WindowTitle = "timebox"
	return v
}

func (m Model) renderList(w, h int) string {
	if len(m.state.Items) == 0 {
		return m.styles.hint.Width(w).Render("n to add")
	}
	lines := make([]string, 0, h)
	end := m.listOffset + h
	if end > len(m.state.Items) {
		end = len(m.state.Items)
	}
	for i := m.listOffset; i < end; i++ {
		lines = append(lines, m.renderRow(m.state.Items[i], w))
	}
	for len(lines) < h {
		lines = append(lines, strings.Repeat(" ", w))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderRow(it domain.Item, w int) string {
	if m.mode == modeEdit && m.focus == focusList && it.ID == m.editID {
		body := clip(m.renderEditRow(), w)
		if it.ID == m.state.SelectedID {
			return m.styles.selectedBg.Width(w).Render(body)
		}
		return lipgloss.NewStyle().Width(w).Render(body)
	}

	dur := fmt.Sprintf("[%s]", domain.FormatHHMM(it.Remaining))
	title := it.Title
	st := m.itemRowStyle(it)
	titleSt := st
	if it.Done {
		titleSt = titleSt.Strikethrough(true)
	}
	prefix := dur + " "
	avail := w - runeCount(prefix)
	if avail < 0 {
		avail = 0
	}
	titleShow := title
	if runeCount(titleShow) > avail {
		titleShow = clip(titleShow, avail)
	}
	body := st.Render(dur) + " " + titleSt.Render(titleShow)
	if it.ID == m.state.SelectedID {
		return m.styles.selectedBg.Width(w).Render(body)
	}
	return lipgloss.NewStyle().Width(w).Render(body)
}

func (m Model) renderEditRow() string {
	dur := m.durationBuf
	title := m.titleBuf
	if m.editField == fieldDuration {
		dur = paintCursor(dur, m.editCursor, m.styles.cursor)
	} else {
		title = paintCursor(title, m.editCursor, m.styles.cursor)
	}
	return "[" + dur + "] " + title
}

// itemRowStyle colors a list row: green running, red overtime, gray done, else white.
func (m Model) itemRowStyle(it domain.Item) lipgloss.Style {
	if it.Done {
		return m.styles.doneItem
	}
	if it.Running {
		if it.Remaining <= 0 {
			return m.styles.overtimeItem
		}
		return m.styles.activeItem
	}
	return m.styles.idleItem
}

func (m Model) renderClock(w, h int) string {
	it := m.clockItem()
	clockText := "--:--:--"
	overtime := ""
	style := m.styles.clock
	if it != nil {
		clockText = domain.FormatHHMMSS(it.Remaining)
		if clockWarnRemaining(it.Remaining) {
			style = m.styles.clockWarn
		}
		atZero := it.Remaining <= 0
		if atZero && it.Running && !m.blinkVisible {
			style = m.styles.clockBlink
		}
		if atZero && it.Overtime > 0 {
			overtime = "[" + domain.FormatHHMM(it.Overtime) + "]"
		}
	}
	digits := style.Render(bigClock(clockText))
	parts := []string{digits}
	if overtime != "" {
		parts = append(parts, m.styles.overtime.Render(overtime))
	}
	if it != nil {
		inv := "Invested time: " + domain.FormatHHMM(it.Invested)
		parts = append(parts, m.styles.overtime.Render(inv))
	}
	block := lipgloss.JoinVertical(lipgloss.Center, parts...)
	return placeCenter(w, h, block)
}

func (m Model) renderNotes(w, h int) string {
	if m.focus == focusNotes {
		return m.notes.View()
	}
	it := m.state.Selected()
	if it == nil || it.Notes == "" {
		return m.styles.hint.Width(w).Render("notes")
	}
	rendered := md.Render(it.Notes, m.styles.msgBold, m.styles.msgItalic, m.styles.msgStrike)
	wrapped := lipgloss.NewStyle().Width(w).Render(rendered)
	lines := strings.Split(wrapped, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func clip(s string, w int) string {
	if w <= 0 {
		return ""
	}
	rs := []rune(s)
	if len(rs) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	return string(rs[:w-1]) + "…"
}

func (m Model) renderDeleteModal() string {
	q := "do you want to delete this item?"
	yes := m.confirmButton("Yes", m.deleteChoice == confirmYes)
	no := m.confirmButton("No", m.deleteChoice == confirmNo)
	done := m.confirmButton("Done", m.deleteChoice == confirmDone)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, yes, no, done)
	inner := lipgloss.JoinVertical(lipgloss.Left, q, "", buttons)
	return m.styles.deleteBox.Render(inner)
}

func (m Model) confirmButton(label string, on bool) string {
	if on {
		return m.styles.confirmActive.Render(label)
	}
	return m.styles.confirmIdle.Render(label)
}
