package store

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"timebox_me/internal/domain"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	orig := domain.State{
		SelectedID: "b",
		Items: []domain.Item{
			{
				ID:        "a",
				Title:     "Ship",
				Remaining: 90 * time.Minute,
				Overtime:  0,
				Invested:  45 * time.Minute,
				Notes:     "*bold* note",
				Running:   true,
			},
			{
				ID:        "b",
				Title:     "Review",
				Remaining: 0,
				Overtime:  29 * time.Minute,
				Notes:     "line1\nline2",
				Done:      true,
			},
		},
	}
	if err := Save(path, orig); err != nil {
		t.Fatalf("save: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if bytes.Contains(raw, []byte("active_id")) {
		t.Fatalf("saved JSON must not write active_id:\n%s", raw)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatalf("saved JSON: %v", err)
	}
	if _, ok := top["running"]; ok {
		t.Fatalf("saved JSON must not write top-level running:\n%s", raw)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.SelectedID != orig.SelectedID {
		t.Fatalf("selected mismatch: %+v", got)
	}
	if len(got.Items) != 2 {
		t.Fatalf("len=%d", len(got.Items))
	}
	if got.Items[0].Title != "Ship" || got.Items[0].Remaining != 90*time.Minute || got.Items[0].Invested != 45*time.Minute {
		t.Fatalf("item0 %+v", got.Items[0])
	}
	if !got.Items[0].Running {
		t.Fatal("item0 must load running")
	}
	if got.Items[1].Overtime != 29*time.Minute || got.Items[1].Notes != "line1\nline2" || !got.Items[1].Done {
		t.Fatalf("item1 %+v", got.Items[1])
	}
	if got.Items[0].Done {
		t.Fatal("item0 must load done=false")
	}
	if got.Items[1].Running {
		t.Fatal("done item must not load running")
	}
}

func TestLoadMissingFile(t *testing.T) {
	st, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("load missing: %v", err)
	}
	if len(st.Items) != 0 || st.SelectedID != "" {
		t.Fatalf("want empty state, got %+v", st)
	}
}

func TestLoadNormalizesDanglingIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	data := []byte(`{"active_id":"gone","running":true,"selected_id":"gone","items":[{"id":"a","title":"x"}]}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if st.Items[0].Running {
		t.Fatalf("dangling legacy runner must not start another item: %+v", st)
	}
	if st.SelectedID != "a" {
		t.Fatalf("selected must fall back to first item, got %q", st.SelectedID)
	}
	if st.Items[0].Done {
		t.Fatal("missing done field must load as false")
	}
	if st.Items[0].Invested != 0 {
		t.Fatal("missing invested field must load as 0")
	}
}

func TestLoadMigratesLegacyRunner(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	data := []byte(`{"active_id":"a","running":true,"selected_id":"b","items":[{"id":"a","title":"Ship","remaining_ns":60000000000},{"id":"b","title":"Review"}]}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !st.Items[0].Running {
		t.Fatal("legacy running+active_id must start that item")
	}
	if st.Items[1].Running {
		t.Fatal("other items must stay paused")
	}
	if st.SelectedID != "b" {
		t.Fatalf("selected=%q, want b", st.SelectedID)
	}
}
