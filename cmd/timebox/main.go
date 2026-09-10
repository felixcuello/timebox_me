// Command timebox is a full-screen TUI for timeboxing tasks.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"timebox_me/internal/store"
	"timebox_me/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "timebox:", err)
		os.Exit(1)
	}
}

func run() error {
	path, err := store.DefaultPath()
	if err != nil {
		return err
	}
	st, err := store.Load(path)
	if err != nil {
		return err
	}
	m := ui.New(st, path)
	prog := tea.NewProgram(m)
	final, err := prog.Run()
	if err != nil {
		return err
	}
	if fm, ok := final.(ui.Model); ok {
		if err := store.Save(path, fm.State()); err != nil {
			return err
		}
	}
	return nil
}
