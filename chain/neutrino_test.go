package chain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestNeutrinoClientStartStop ensures that the client
// can sequentially Start and Stop without errors or races.
func TestNeutrinoClientStartStop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	t.Cleanup(cancel)
	nc := newMockNeutrinoClient(t)
	numRestarts := 5

	startStop := func() <-chan struct{} {
		done := make(chan struct{})
		go func() {
			defer close(done)
			err := nc.Start()
			require.NoError(t, err)
			nc.Stop()
			nc.WaitForShutdown()
		}()
		return done
	}

	for i := 0; i < numRestarts; i++ {
		done := startStop()
		select {
		case <-ctx.Done():
			t.Fatal("timed out")
		case <-done:
		}
	}
}
