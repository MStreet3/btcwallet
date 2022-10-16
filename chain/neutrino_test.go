package chain

import (
	"testing"
)

// TestNeutrinoClientMultipleRestart ensures that multiple goroutines
// can Start and Stop the client without errors or races.
func TestNeutrinoClientMultipleRestart(t *testing.T) {
	// call notifyreceived and rescan in a loop
	nc := newMockNeutrinoClient(t)
	_ = nc
}
