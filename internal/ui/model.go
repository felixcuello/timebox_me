package ui

import (
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

	"timebox_me/internal/domain"
)

type focusZone int

const (
	focusList focusZone = iota
	focusNotes
)

type listMode int

const (
	modeNavigate listMode = iota
	modeEdit
)

type editField int

const (
	fieldDuration editField = iota
	fieldTitle
)

const (
	saveDebounce   = 5 * time.Second
	minListWidth   = 24
	minRightWidth  = 36
	minClockHeight = 9
	minNotesHeight = 6
	defaultClockH  = 13
)

// Model is the root Bubble Tea model for the timebox TUI.
type Model struct {
	path   string
	state  domain.State
	styles styles
	keys   keyMap
	notes  textarea.Model

	width, height int
	ready         bool
	listWidth     int
	clockHeight   int
	listOffset    int

	focus         focusZone
	mode          listMode
	editField     editField
	editID        string
	durationBuf   string
	titleBuf      string
	editCursor    int
	durationDirty bool

	dirty         bool
	saveInFlight  bool
	needSave      bool
	lastPersist   time.Time
	persistErr    string
	blinkVisible  bool
	deleteConfirm bool
	deleteChoice  confirmChoice
}

// New constructs the TUI from loaded state and the JSON path used for saves.
func New(st domain.State, path string) Model {
	st.Normalize()
	ta := textarea.New()
	ta.Placeholder = "notes"
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("enter"))

	return Model{
		path:         path,
		state:        st,
		styles:       newStyles(),
		keys:         defaultKeys(),
		notes:        ta,
		focus:        focusList,
		mode:         modeNavigate,
		blinkVisible: true,
		lastPersist:  time.Now(),
		deleteChoice: confirmYes,
	}
}

// State returns a snapshot for flushing to disk on process exit.
func (m Model) State() domain.State {
	m.syncNotesIntoState()
	if m.mode == modeEdit {
		m.commitEditFields()
	}
	return m.state.Clone()
}

func (m *Model) syncNotesIntoState() {
	if m.focus != focusNotes {
		return
	}
	if it := m.state.Selected(); it != nil {
		m.state.CommitNotes(it.ID, m.notes.Value())
	}
}

func (m *Model) applyLayout() {
	if m.width < 1 || m.height < 1 {
		return
	}
	listW := m.width / 2
	if listW < minListWidth {
		listW = minListWidth
	}
	if m.width-listW < minRightWidth {
		listW = m.width - minRightWidth
	}
	if listW < 10 {
		listW = m.width / 2
	}
	m.listWidth = listW

	clockH := defaultClockH
	if m.height-clockH < minNotesHeight {
		clockH = m.height - minNotesHeight
	}
	if clockH < minClockHeight {
		clockH = minClockHeight
	}
	if clockH > m.height-3 {
		clockH = m.height - 3
	}
	if clockH < 5 {
		clockH = m.height / 2
	}
	m.clockHeight = clockH

	rightW := m.width - m.listWidth
	notesOuterH := m.height - m.clockHeight
	innerW := rightW - 2
	innerH := notesOuterH - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 1 {
		innerH = 1
	}
	m.notes.SetWidth(innerW)
	m.notes.SetHeight(innerH)
	m.ensureSelectedVisible()
}

func (m *Model) ensureSelectedVisible() {
	innerH := m.height - 2
	if innerH < 1 {
		innerH = 1
	}
	idx := m.state.SelectedIndex()
	if idx < 0 {
		m.listOffset = 0
		return
	}
	if idx < m.listOffset {
		m.listOffset = idx
	}
	if idx >= m.listOffset+innerH {
		m.listOffset = idx - innerH + 1
	}
	if m.listOffset < 0 {
		m.listOffset = 0
	}
}

func (m *Model) startEdit() {
	it := m.state.Selected()
	if it == nil {
		return
	}
	m.mode = modeEdit
	m.editField = fieldDuration
	m.loadEditBuffers()
}

func (m *Model) loadEditBuffers() {
	it := m.state.Selected()
	if it == nil {
		m.editID = ""
		m.durationBuf = ""
		m.titleBuf = ""
		m.editCursor = 0
		return
	}
	m.editID = it.ID
	m.durationBuf = durationMaskFromRemaining(it.Remaining)
	m.titleBuf = it.Title
	m.durationDirty = false
	m.resetEditCursor()
}

func (m *Model) resetEditCursor() {
	if m.editField == fieldDuration {
		m.editCursor = 0
	} else {
		m.editCursor = runeCount(m.titleBuf)
	}
}

func (m *Model) commitEditFields() {
	if m.editID == "" {
		return
	}
	if m.durationDirty {
		m.state.CommitDuration(m.editID, m.durationBuf)
	}
	m.state.CommitTitle(m.editID, m.titleBuf)
	if it := itemByID(&m.state, m.editID); it != nil {
		m.durationBuf = domain.FormatHHMM(it.Remaining)
		m.titleBuf = it.Title
	}
	m.durationDirty = false
}

func itemByID(st *domain.State, id string) *domain.Item {
	for i := range st.Items {
		if st.Items[i].ID == id {
			return &st.Items[i]
		}
	}
	return nil
}

func (m *Model) markDirty() {
	m.dirty = true
}

func (m *Model) enterNotes() tea.Cmd {
	m.focus = focusNotes
	m.mode = modeNavigate
	if it := m.state.Selected(); it != nil {
		m.notes.SetValue(it.Notes)
	} else {
		m.notes.SetValue("")
	}
	return m.notes.Focus()
}

func (m *Model) leaveNotes() {
	m.syncNotesIntoState()
	m.notes.Blur()
	m.focus = focusList
}

func (m *Model) openDeleteModal() {
	if m.state.Selected() == nil {
		return
	}
	m.deleteConfirm = true
	m.deleteChoice = confirmYes
}

func (m *Model) closeDeleteModal() {
	m.deleteConfirm = false
	m.deleteChoice = confirmYes
}

// clockItem is the item shown on the big clock: always the selected row.
func (m Model) clockItem() *domain.Item {
	return m.state.Selected()
}
