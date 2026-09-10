package domain

import (
	"testing"
	"time"
)

func TestTickRemainingThenOvertime(t *testing.T) {
	s := State{
		SelectedID: "a",
		Items: []Item{{
			ID:        "a",
			Remaining: 2 * time.Second,
			Running:   true,
		}},
	}
	if s.Tick() != 0 {
		t.Fatal("first tick must not report ended")
	}
	if s.Items[0].Remaining != time.Second {
		t.Fatalf("remaining after 1s = %v, want 1s", s.Items[0].Remaining)
	}
	if s.Items[0].Invested != time.Second {
		t.Fatalf("invested after 1s = %v, want 1s", s.Items[0].Invested)
	}
	if s.Tick() != 1 {
		t.Fatal("tick that hits zero must report ended")
	}
	if s.Items[0].Remaining != 0 {
		t.Fatalf("remaining after 2s = %v, want 0", s.Items[0].Remaining)
	}
	if s.Items[0].Overtime != 0 {
		t.Fatalf("overtime must stay 0 on the tick that hits zero, got %v", s.Items[0].Overtime)
	}
	if s.Tick() != 0 {
		t.Fatal("overtime ticks must not report ended")
	}
	if s.Items[0].Overtime != time.Second {
		t.Fatalf("overtime after 1s past zero = %v, want 1s", s.Items[0].Overtime)
	}
	if s.Items[0].Invested != 3*time.Second {
		t.Fatalf("invested after 3 ticks = %v, want 3s", s.Items[0].Invested)
	}
	s.TickBy(29 * time.Minute)
	if FormatHHMM(s.Items[0].Overtime) != "00:29" {
		t.Fatalf("overtime display = %s, want 00:29", FormatHHMM(s.Items[0].Overtime))
	}
}

func TestPauseDoesNotTick(t *testing.T) {
	s := State{
		Items: []Item{{ID: "a", Remaining: time.Minute, Running: false}},
	}
	s.Tick()
	if s.Items[0].Remaining != time.Minute {
		t.Fatalf("paused remaining changed: %v", s.Items[0].Remaining)
	}
	s.Items[0].Overtime = time.Minute
	s.Items[0].Invested = time.Minute
	s.Tick()
	if s.Items[0].Overtime != time.Minute {
		t.Fatalf("paused overtime changed: %v", s.Items[0].Overtime)
	}
	if s.Items[0].Invested != time.Minute {
		t.Fatalf("paused invested changed: %v", s.Items[0].Invested)
	}
}

func TestSpaceStartsWithoutStoppingOthers(t *testing.T) {
	s := State{
		SelectedID: "a",
		Items: []Item{
			{ID: "a", Remaining: 5 * time.Minute},
			{ID: "b", Remaining: 10 * time.Minute},
		},
	}
	s.Space()
	if !s.Items[0].Running {
		t.Fatal("space on a must start a")
	}
	s.Tick()
	if s.Items[0].Remaining != 5*time.Minute-time.Second {
		t.Fatalf("a must tick, remaining=%v", s.Items[0].Remaining)
	}
	investedA := s.Items[0].Invested

	s.SelectByIndex(1)
	s.Space()
	if !s.Items[0].Running || !s.Items[1].Running {
		t.Fatal("space on b must start b and leave a running")
	}
	s.Tick()
	if s.Items[0].Remaining != 5*time.Minute-2*time.Second {
		t.Fatalf("a must keep ticking, remaining=%v", s.Items[0].Remaining)
	}
	if s.Items[0].Invested != investedA+time.Second {
		t.Fatalf("a invested must keep growing, got %v", s.Items[0].Invested)
	}
	if s.Items[1].Remaining != 10*time.Minute-time.Second {
		t.Fatalf("b must tick, remaining=%v", s.Items[1].Remaining)
	}

	s.Space()
	if s.Items[1].Running {
		t.Fatal("space on green b must pause b")
	}
	if !s.Items[0].Running {
		t.Fatal("pausing b must leave a running")
	}
	frozenB := s.Items[1].Remaining
	s.Tick()
	if s.Items[1].Remaining != frozenB {
		t.Fatalf("paused b must freeze, remaining=%v", s.Items[1].Remaining)
	}
	if s.Items[0].Remaining != 5*time.Minute-3*time.Second {
		t.Fatalf("a must keep ticking after b pauses, remaining=%v", s.Items[0].Remaining)
	}
}

func TestSpaceIgnoresDone(t *testing.T) {
	s := State{
		SelectedID: "a",
		Items:      []Item{{ID: "a", Done: true, Remaining: time.Minute}},
	}
	s.Space()
	if s.Items[0].Running {
		t.Fatal("space on a done item must be ignored")
	}
}

func TestTickTwoItemsHitZeroTogether(t *testing.T) {
	s := State{
		Items: []Item{
			{ID: "a", Remaining: time.Second, Running: true},
			{ID: "b", Remaining: time.Second, Running: true},
		},
	}
	if n := s.Tick(); n != 2 {
		t.Fatalf("ended=%d, want 2", n)
	}
}

func TestCommitDurationInvalidKeepsPrevious(t *testing.T) {
	s := State{
		Items: []Item{{ID: "a", Remaining: 45 * time.Minute, Overtime: time.Minute}},
	}
	s.CommitDuration("a", "nope")
	if s.Items[0].Remaining != 45*time.Minute {
		t.Fatalf("remaining = %v, want 45m", s.Items[0].Remaining)
	}
	s.CommitDuration("a", "01:30")
	if s.Items[0].Remaining != time.Hour+30*time.Minute {
		t.Fatalf("remaining = %v, want 1h30m", s.Items[0].Remaining)
	}
	if s.Items[0].Overtime != 0 {
		t.Fatalf("overtime must clear when remaining is set above zero, got %v", s.Items[0].Overtime)
	}
}

func TestAddAtTop(t *testing.T) {
	s := State{}
	s.AddAtTop()
	if len(s.Items) != 1 || s.SelectedID == "" {
		t.Fatalf("first add: %+v", s)
	}
	first := s.SelectedID
	s.AddAtTop()
	if len(s.Items) != 2 {
		t.Fatalf("len=%d", len(s.Items))
	}
	if s.Items[0].ID != s.SelectedID {
		t.Fatal("newest item must be first and selected")
	}
	if s.Items[1].ID != first {
		t.Fatal("older item must move down")
	}
	s.Items[0].Running = true
	s.Items[1].Running = true
	s.DeleteSelected()
	if len(s.Items) != 1 {
		t.Fatalf("len=%d", len(s.Items))
	}
	if s.Items[0].Running != true {
		t.Fatal("deleting one runner must leave the other running")
	}
	s.DeleteSelected()
	if len(s.Items) != 0 || s.SelectedID != "" {
		t.Fatalf("empty after last delete: %+v", s)
	}
}

func TestMoveSelected(t *testing.T) {
	s := State{
		SelectedID: "a",
		Items: []Item{
			{ID: "a"},
			{ID: "b"},
			{ID: "c"},
		},
	}
	s.MoveSelected(-1)
	if s.Items[0].ID != "a" || s.SelectedID != "a" {
		t.Fatalf("move up at top must no-op: %+v", s.Items)
	}
	s.MoveSelected(1)
	if s.Items[0].ID != "b" || s.Items[1].ID != "a" || s.SelectedID != "a" {
		t.Fatalf("move down: %+v selected=%s", s.Items, s.SelectedID)
	}
	s.SelectByIndex(2)
	s.MoveSelected(1)
	if s.Items[2].ID != "c" {
		t.Fatalf("move down at bottom must no-op: %+v", s.Items)
	}
	s.MoveSelected(-1)
	if s.Items[1].ID != "c" || s.Items[2].ID != "a" || s.SelectedID != "c" {
		t.Fatalf("move up: %+v selected=%s", s.Items, s.SelectedID)
	}
}

func TestToggleDoneStopsOnlyThatRunner(t *testing.T) {
	s := State{
		SelectedID: "a",
		Items: []Item{
			{ID: "a", Running: true},
			{ID: "b", Running: true},
		},
	}
	s.ToggleDone()
	if !s.Items[0].Done {
		t.Fatal("must mark done")
	}
	if s.Items[0].Running {
		t.Fatal("marking a runner done must stop that clock")
	}
	if !s.Items[1].Running {
		t.Fatal("the other runner must keep ticking")
	}
	s.ToggleDone()
	if s.Items[0].Done {
		t.Fatal("toggle must clear done")
	}
	if s.Items[0].Running {
		t.Fatal("undoing done must not auto-resume")
	}
	s.SelectByIndex(1)
	s.MarkSelectedDone()
	if !s.Items[1].Done {
		t.Fatal("MarkSelectedDone must set done")
	}
	if s.Items[1].Running {
		t.Fatal("MarkSelectedDone must stop that runner")
	}
}

func TestNormalizeStopsRunningOnDone(t *testing.T) {
	s := State{
		SelectedID: "a",
		Items:      []Item{{ID: "a", Done: true, Running: true}},
	}
	s.Normalize()
	if s.Items[0].Running {
		t.Fatal("done items must not stay running")
	}
}

func TestCloneIndependent(t *testing.T) {
	s := State{Items: []Item{{ID: "a", Title: "x"}}}
	c := s.Clone()
	c.Items[0].Title = "y"
	if s.Items[0].Title != "x" {
		t.Fatal("clone must not share the items slice")
	}
}
