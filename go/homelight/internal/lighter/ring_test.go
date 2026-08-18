package lighter

import "testing"

func TestNextDeviceIDWraps(t *testing.T) {
	if got := nextDeviceID(4); got != 5 {
		t.Fatalf("next after 4 = %d, want 5", got)
	}
	if got := nextDeviceID(5); got != 0 {
		t.Fatalf("next after 5 = %d, want 0", got)
	}
}

func TestAdvanceTargetSkipsSelf(t *testing.T) {
	if got := advanceTarget(0, 1); got != 2 {
		t.Fatalf("advance from 0 skipping self 1 = %d, want 2", got)
	}
}

func TestSilenceThresholdForComputer(t *testing.T) {
	if got := silenceThreshold(0); got != ringSilenceBaseTicks {
		t.Fatalf("threshold for id 0 = %d, want %d", got, ringSilenceBaseTicks)
	}
}
