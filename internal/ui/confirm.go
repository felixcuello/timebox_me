package ui

// confirmChoice is the highlighted button on the delete modal.
type confirmChoice int

const (
	confirmYes confirmChoice = iota
	confirmNo
	confirmDone
)

const confirmChoiceCount = 3

// cycleConfirm moves the highlight by delta and wraps.
func cycleConfirm(c confirmChoice, delta int) confirmChoice {
	n := (int(c) + delta) % confirmChoiceCount
	if n < 0 {
		n += confirmChoiceCount
	}
	return confirmChoice(n)
}
