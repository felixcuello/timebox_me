package domain

import "time"

// Item is one timeboxed task with remaining budget, overtime, and notes.
type Item struct {
	ID        string
	Title     string
	Remaining time.Duration
	Overtime  time.Duration
	Invested  time.Duration
	Notes     string
	Done      bool
	Running   bool
}

// State is the full session: items and which row is selected.
type State struct {
	// SelectedID is the highlighted list row (notes and the big clock follow this item).
	SelectedID string
	Items      []Item
}

// Clone returns a deep copy so async save commands cannot race with later edits.
func (s State) Clone() State {
	out := s
	if s.Items != nil {
		out.Items = make([]Item, len(s.Items))
		copy(out.Items, s.Items)
	}
	return out
}

// Normalize repairs dangling IDs after load or delete and stops clocks on done items.
func (s *State) Normalize() {
	if s.itemIndex(s.SelectedID) < 0 {
		if len(s.Items) > 0 {
			s.SelectedID = s.Items[0].ID
		} else {
			s.SelectedID = ""
		}
	}
	for i := range s.Items {
		if s.Items[i].Done {
			s.Items[i].Running = false
		}
	}
}

func (s *State) itemIndex(id string) int {
	if id == "" {
		return -1
	}
	for i := range s.Items {
		if s.Items[i].ID == id {
			return i
		}
	}
	return -1
}

// SelectedIndex returns the selected row index, or -1 when the list is empty.
func (s *State) SelectedIndex() int {
	return s.itemIndex(s.SelectedID)
}

// Selected returns the highlighted item, or nil when the list is empty.
func (s *State) Selected() *Item {
	i := s.itemIndex(s.SelectedID)
	if i < 0 {
		return nil
	}
	return &s.Items[i]
}

// AnyRunning reports whether at least one item is ticking.
func (s *State) AnyRunning() bool {
	for i := range s.Items {
		if s.Items[i].Running {
			return true
		}
	}
	return false
}

// SelectByIndex highlights the item at i. Out-of-range values are ignored.
func (s *State) SelectByIndex(i int) {
	if i < 0 || i >= len(s.Items) {
		return
	}
	s.SelectedID = s.Items[i].ID
}

// SelectDelta moves the highlight by delta (+1 down, -1 up) and clamps to the list.
func (s *State) SelectDelta(delta int) {
	if len(s.Items) == 0 {
		return
	}
	i := s.SelectedIndex()
	if i < 0 {
		s.SelectedID = s.Items[0].ID
		return
	}
	i += delta
	if i < 0 {
		i = 0
	}
	if i >= len(s.Items) {
		i = len(s.Items) - 1
	}
	s.SelectedID = s.Items[i].ID
}

// Space toggles the selected item's clock. Done items are ignored. Other runners stay as they are.
func (s *State) Space() {
	sel := s.Selected()
	if sel == nil || sel.Done {
		return
	}
	sel.Running = !sel.Running
}

// AddAtTop inserts a zero-duration empty item at the top of the list and selects it.
func (s *State) AddAtTop() *Item {
	item := Item{ID: NewID()}
	s.Items = append([]Item{item}, s.Items...)
	s.SelectedID = item.ID
	return &s.Items[0]
}

// MoveSelected swaps the selected item with its neighbor by delta (+1 down, -1 up).
// Out-of-range moves are ignored. SelectedID stays on the same item.
func (s *State) MoveSelected(delta int) {
	i := s.SelectedIndex()
	if i < 0 || delta == 0 {
		return
	}
	j := i + delta
	if j < 0 || j >= len(s.Items) {
		return
	}
	s.Items[i], s.Items[j] = s.Items[j], s.Items[i]
}

// ToggleDone flips Done on the selected item.
// Marking a running item done stops that clock only. Undoing done does not resume it.
func (s *State) ToggleDone() {
	it := s.Selected()
	if it == nil {
		return
	}
	it.Done = !it.Done
	if it.Done {
		it.Running = false
	}
}

// MarkSelectedDone sets Done on the selected item without toggling it off.
func (s *State) MarkSelectedDone() {
	it := s.Selected()
	if it == nil {
		return
	}
	it.Done = true
	it.Running = false
}

// DeleteSelected removes the highlighted item. Other runners keep ticking.
func (s *State) DeleteSelected() {
	i := s.SelectedIndex()
	if i < 0 {
		return
	}
	s.Items = append(s.Items[:i], s.Items[i+1:]...)
	if len(s.Items) == 0 {
		s.SelectedID = ""
		return
	}
	if i >= len(s.Items) {
		i = len(s.Items) - 1
	}
	s.SelectedID = s.Items[i].ID
}

// CommitDuration sets remaining from hh:mm text. On parse error the previous remaining is kept.
// A successful parse with remaining greater than zero clears overtime.
func (s *State) CommitDuration(id, text string) {
	i := s.itemIndex(id)
	if i < 0 {
		return
	}
	d, err := ParseHHMM(text)
	if err != nil {
		return
	}
	s.Items[i].Remaining = d
	if d > 0 {
		s.Items[i].Overtime = 0
	}
}

// CommitTitle sets the item title.
func (s *State) CommitTitle(id, title string) {
	i := s.itemIndex(id)
	if i < 0 {
		return
	}
	s.Items[i].Title = title
}

// CommitNotes sets the item notes.
func (s *State) CommitNotes(id, notes string) {
	i := s.itemIndex(id)
	if i < 0 {
		return
	}
	s.Items[i].Notes = notes
}

// Tick applies one second to every running item.
// It returns how many items reached zero remaining on this tick.
func (s *State) Tick() int {
	return s.TickBy(time.Second)
}

// TickBy applies step to every running item. Used by Tick and by tests.
// It returns how many items reached zero remaining on this step.
func (s *State) TickBy(step time.Duration) int {
	if step <= 0 {
		return 0
	}
	ended := 0
	for i := range s.Items {
		if !s.Items[i].Running {
			continue
		}
		if tickItem(&s.Items[i], step) {
			ended++
		}
	}
	return ended
}

// tickItem advances one running item. It returns true when remaining just reached zero.
func tickItem(item *Item, step time.Duration) bool {
	item.Invested += step
	if item.Remaining > 0 {
		item.Remaining -= step
		if item.Remaining < 0 {
			item.Remaining = 0
		}
		return item.Remaining == 0
	}
	item.Overtime += step
	return false
}
