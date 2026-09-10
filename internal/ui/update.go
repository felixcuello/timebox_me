package ui

import (
	"fmt"
	"os"
	"time"

	key "charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"timebox_me/internal/store"
)

type tickMsg time.Time

type savedMsg struct{ err error }

// Init starts the one-second clock tick.
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

const bellGap = 200 * time.Millisecond

// bellEndedCmd rings two BELs per item that just hit zero, sequentially on stderr.
func bellEndedCmd(n int) tea.Cmd {
	if n <= 0 {
		return nil
	}
	return func() tea.Msg {
		for i := 0; i < n; i++ {
			if i > 0 {
				time.Sleep(bellGap)
			}
			fmt.Fprint(os.Stderr, "\a")
			time.Sleep(bellGap)
			fmt.Fprint(os.Stderr, "\a")
		}
		return nil
	}
}

// Update is the root message handler.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		m.applyLayout()
		return m, nil

	case tickMsg:
		cmds := []tea.Cmd{tickCmd()}
		if m.state.AnyRunning() {
			ended := m.state.Tick()
			m.markDirty()
			if ended > 0 {
				cmds = append(cmds, bellEndedCmd(ended))
			}
		}
		it := m.clockItem()
		if it != nil && it.Running && it.Remaining <= 0 {
			m.blinkVisible = !m.blinkVisible
		} else {
			m.blinkVisible = true
		}
		if cmd := m.persistCmd(false); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case savedMsg:
		m.saveInFlight = false
		m.lastPersist = time.Now()
		if msg.err != nil {
			m.persistErr = "save: " + msg.err.Error()
			m.dirty = true
		} else {
			m.persistErr = ""
		}
		if m.needSave {
			m.needSave = false
			m.dirty = true
			return m, m.persistCmd(true)
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	if m.focus == focusNotes {
		var cmd tea.Cmd
		m.notes, cmd = m.notes.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		m.syncNotesIntoState()
		if m.mode == modeEdit {
			m.commitEditFields()
		}
		return m, tea.Sequence(m.persistCmd(true), tea.Quit)
	}

	if m.deleteConfirm {
		return m.handleDeleteModalKey(msg)
	}

	if m.focus == focusNotes {
		return m.handleNotesKey(msg)
	}
	if m.mode == modeEdit {
		return m.handleEditKey(msg)
	}
	return m.handleNavigateKey(msg)
}

func (m Model) handleDeleteModalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.closeDeleteModal()
		return m, nil
	case key.Matches(msg, m.keys.Left), key.Matches(msg, m.keys.ShiftTab):
		m.deleteChoice = cycleConfirm(m.deleteChoice, -1)
		return m, nil
	case key.Matches(msg, m.keys.Right), key.Matches(msg, m.keys.Tab):
		m.deleteChoice = cycleConfirm(m.deleteChoice, 1)
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		return m.applyDeleteChoice()
	}
	return m, nil
}

func (m Model) applyDeleteChoice() (tea.Model, tea.Cmd) {
	switch m.deleteChoice {
	case confirmYes:
		m.state.DeleteSelected()
		m.ensureSelectedVisible()
		m.closeDeleteModal()
		m.markDirty()
		return m, m.persistCmd(true)
	case confirmDone:
		m.state.MarkSelectedDone()
		m.closeDeleteModal()
		m.markDirty()
		return m, m.persistCmd(true)
	default:
		m.closeDeleteModal()
		return m, nil
	}
}

func (m Model) handleNavigateKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "q":
		m.syncNotesIntoState()
		return m, tea.Sequence(m.persistCmd(true), tea.Quit)
	case key.Matches(msg, m.keys.Tab), key.Matches(msg, m.keys.ShiftTab):
		return m, m.enterNotes()
	case key.Matches(msg, m.keys.Enter):
		m.startEdit()
		return m, nil
	case key.Matches(msg, m.keys.Space):
		m.state.Space()
		m.blinkVisible = true
		m.markDirty()
		return m, m.persistCmd(true)
	case key.Matches(msg, m.keys.New):
		m.state.AddAtTop()
		m.ensureSelectedVisible()
		m.startEdit()
		m.markDirty()
		return m, m.persistCmd(true)
	case key.Matches(msg, m.keys.Delete):
		m.openDeleteModal()
		return m, nil
	case key.Matches(msg, m.keys.MoveDown):
		m.state.MoveSelected(1)
		m.ensureSelectedVisible()
		m.markDirty()
		return m, m.persistCmd(true)
	case key.Matches(msg, m.keys.MoveUp):
		m.state.MoveSelected(-1)
		m.ensureSelectedVisible()
		m.markDirty()
		return m, m.persistCmd(true)
	case key.Matches(msg, m.keys.MarkDone):
		m.state.ToggleDone()
		m.markDirty()
		return m, m.persistCmd(true)
	case key.Matches(msg, m.keys.Up):
		m.state.SelectDelta(-1)
		m.ensureSelectedVisible()
		m.markDirty()
		return m, nil
	case key.Matches(msg, m.keys.Down):
		m.state.SelectDelta(1)
		m.ensureSelectedVisible()
		m.markDirty()
		return m, nil
	}
	return m, nil
}

func (m Model) handleEditKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.commitEditFields()
		m.mode = modeNavigate
		m.markDirty()
		return m, m.persistCmd(true)
	case key.Matches(msg, m.keys.Tab):
		m.commitEditFields()
		if m.editField == fieldDuration {
			m.editField = fieldTitle
		} else {
			m.editField = fieldDuration
		}
		m.resetEditCursor()
		m.markDirty()
		return m, nil
	case key.Matches(msg, m.keys.ShiftTab):
		m.commitEditFields()
		if m.editField == fieldTitle {
			m.editField = fieldDuration
		} else {
			m.editField = fieldTitle
		}
		m.resetEditCursor()
		m.markDirty()
		return m, nil
	case key.Matches(msg, m.keys.Up):
		m.commitEditFields()
		m.state.SelectDelta(-1)
		m.ensureSelectedVisible()
		m.loadEditBuffers()
		m.markDirty()
		return m, m.persistCmd(true)
	case key.Matches(msg, m.keys.Down):
		m.commitEditFields()
		m.state.SelectDelta(1)
		m.ensureSelectedVisible()
		m.loadEditBuffers()
		m.markDirty()
		return m, m.persistCmd(true)
	case key.Matches(msg, m.keys.Left):
		if m.editField == fieldDuration {
			m.editCursor = prevDurationCursor(m.editCursor)
		} else if m.editCursor > 0 {
			m.editCursor--
		}
		return m, nil
	case key.Matches(msg, m.keys.Right):
		if m.editField == fieldDuration {
			m.editCursor = nextDurationCursor(m.editCursor)
		} else if m.editCursor < m.currentFieldLen() {
			m.editCursor++
		}
		return m, nil
	case key.Matches(msg, m.keys.Backspace):
		m.editBackspace()
		return m, nil
	}

	if m.editField == fieldDuration && key.Matches(msg, m.keys.Space) {
		return m, nil
	}

	if msg.Text != "" {
		for _, r := range msg.Text {
			m.editInsert(r)
		}
	}
	return m, nil
}

func (m Model) handleNotesKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Tab) || key.Matches(msg, m.keys.ShiftTab) {
		m.leaveNotes()
		m.markDirty()
		return m, m.persistCmd(true)
	}
	var cmd tea.Cmd
	m.notes, cmd = m.notes.Update(msg)
	m.markDirty()
	return m, cmd
}

func (m Model) currentFieldLen() int {
	if m.editField == fieldDuration {
		return runeCount(m.durationBuf)
	}
	return runeCount(m.titleBuf)
}

func (m *Model) editInsert(r rune) {
	if m.editField == fieldDuration {
		buf, cur, ok := overwriteDurationDigit(m.durationBuf, m.editCursor, r)
		if ok {
			m.durationBuf = buf
			m.editCursor = cur
			m.durationDirty = true
		}
		return
	}
	if r == '\n' || r == '\t' {
		return
	}
	m.titleBuf, m.editCursor = insertRune(m.titleBuf, m.editCursor, r)
}

func (m *Model) editBackspace() {
	if m.editField == fieldDuration {
		prev := m.durationBuf
		m.durationBuf, m.editCursor = backspaceDuration(m.durationBuf, m.editCursor)
		if m.durationBuf != prev {
			m.durationDirty = true
		}
		return
	}
	m.titleBuf, m.editCursor = deleteBefore(m.titleBuf, m.editCursor)
}

func (m *Model) persistCmd(force bool) tea.Cmd {
	m.syncNotesIntoState()
	if !m.dirty && !force {
		return nil
	}
	if m.saveInFlight {
		if force {
			m.needSave = true
		}
		return nil
	}
	if !force && time.Since(m.lastPersist) < saveDebounce {
		return nil
	}
	m.saveInFlight = true
	m.dirty = false
	path := m.path
	snap := m.state.Clone()
	return func() tea.Msg {
		return savedMsg{err: store.Save(path, snap)}
	}
}
