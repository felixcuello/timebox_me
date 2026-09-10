package ui

import "testing"

func TestCycleConfirm(t *testing.T) {
	if cycleConfirm(confirmYes, 1) != confirmNo {
		t.Fatal("yes +1 must be no")
	}
	if cycleConfirm(confirmNo, 1) != confirmDone {
		t.Fatal("no +1 must be done")
	}
	if cycleConfirm(confirmDone, 1) != confirmYes {
		t.Fatal("done +1 must wrap to yes")
	}
	if cycleConfirm(confirmYes, -1) != confirmDone {
		t.Fatal("yes -1 must wrap to done")
	}
	if cycleConfirm(confirmYes, 0) != confirmYes {
		t.Fatal("delta 0 must stay")
	}
}
