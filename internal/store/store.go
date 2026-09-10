package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"timebox_me/internal/domain"
)

const (
	dirPerm  = 0o700
	filePerm = 0o600
)

type fileState struct {
	SelectedID     string     `json:"selected_id"`
	Items          []fileItem `json:"items"`
	LegacyActiveID string     `json:"active_id,omitempty"`
	LegacyRunning  bool       `json:"running,omitempty"`
}

type fileItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	RemainingNS int64  `json:"remaining_ns"`
	OvertimeNS  int64  `json:"overtime_ns"`
	InvestedNS  int64  `json:"invested_ns"`
	Notes       string `json:"notes"`
	Done        bool   `json:"done"`
	Running     bool   `json:"running"`
}

// DefaultPath returns ~/.timebox_me/state.json, creating the directory if needed.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	dir := filepath.Join(home, ".timebox_me")
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return "", fmt.Errorf("create store directory: %w", err)
	}
	return filepath.Join(dir, "state.json"), nil
}

// Load reads state from path. A missing file returns an empty State.
func Load(path string) (domain.State, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return domain.State{}, nil
	}
	if err != nil {
		return domain.State{}, fmt.Errorf("read store: %w", err)
	}
	var fs fileState
	if err := json.Unmarshal(data, &fs); err != nil {
		return domain.State{}, fmt.Errorf("parse store: %w", err)
	}
	st := fromFile(fs)
	st.Normalize()
	return st, nil
}

// Save writes st to path atomically (temp file, then rename).
func Save(path string, st domain.State) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return fmt.Errorf("create store directory: %w", err)
	}
	data, err := json.MarshalIndent(toFile(st), "", "  ")
	if err != nil {
		return fmt.Errorf("encode store: %w", err)
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, filePerm); err != nil {
		return fmt.Errorf("write store temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace store: %w", err)
	}
	return nil
}

func toFile(st domain.State) fileState {
	items := make([]fileItem, len(st.Items))
	for i, it := range st.Items {
		items[i] = fileItem{
			ID:          it.ID,
			Title:       it.Title,
			RemainingNS: it.Remaining.Nanoseconds(),
			OvertimeNS:  it.Overtime.Nanoseconds(),
			InvestedNS:  it.Invested.Nanoseconds(),
			Notes:       it.Notes,
			Done:        it.Done,
			Running:     it.Running,
		}
	}
	return fileState{
		SelectedID: st.SelectedID,
		Items:      items,
	}
}

func fromFile(fs fileState) domain.State {
	items := make([]domain.Item, len(fs.Items))
	for i, it := range fs.Items {
		items[i] = domain.Item{
			ID:        it.ID,
			Title:     it.Title,
			Remaining: time.Duration(it.RemainingNS),
			Overtime:  time.Duration(it.OvertimeNS),
			Invested:  time.Duration(it.InvestedNS),
			Notes:     it.Notes,
			Done:      it.Done,
			Running:   it.Running,
		}
	}
	applyLegacyRunner(fs, items)
	return domain.State{
		SelectedID: fs.SelectedID,
		Items:      items,
	}
}

// applyLegacyRunner starts the old single runner when the file has no per-item running flags.
func applyLegacyRunner(fs fileState, items []domain.Item) {
	for i := range items {
		if items[i].Running {
			return
		}
	}
	if !fs.LegacyRunning {
		return
	}
	for i := range items {
		if items[i].ID == fs.LegacyActiveID && !items[i].Done {
			items[i].Running = true
			return
		}
	}
}
