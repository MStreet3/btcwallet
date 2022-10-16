package chain

import (
	"context"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/stretchr/testify/require"
)

// TestNeutrinoClientSequentialStartStop ensures that the client
// can sequentially Start and Stop without errors or races.
func TestNeutrinoClientSequentialStartStop(t *testing.T) {
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

// TestNeutrinoClientNotifyReceived verifies that a call to NotifyReceived sets the client into
// the scanning state and that subsequent calls while scanning will call Update on the
// client's Rescanner
func TestNeutrinoClientNotifyReceived(t *testing.T) {
	var (
		ctx, cancel             = context.WithTimeout(context.Background(), 1*time.Second)
		addrs                   []btcutil.Address
		sent                    = make(chan struct{})
		read                    = make(chan struct{})
		called                  = make(chan struct{})
		nc                      = newMockNeutrinoClient(t)
		wantNotifyReceivedCalls = 4
		wantUpdateCalls         = wantNotifyReceivedCalls - 1
		gotUC                   = 0
	)
	t.Cleanup(cancel)

	go func() {
		defer close(sent)
		for i := 0; i < wantNotifyReceivedCalls; i++ {
			err := nc.NotifyReceived(addrs)
			require.NoError(t, err)
			require.True(t, nc.scanning)

			// signal that NotifyReceived was called on first iteration
			if i == 0 {
				close(called)
			}
		}
	}()

	// wait until called, then type cast and read from private channel
	<-called
	mockRescan := nc.rescan.(*mockRescanner)
	go func() {
		defer close(read)
		for {
			select {
			case <-ctx.Done():
				return
			case <-mockRescan.updateCh:
				gotUC++
				if gotUC == wantUpdateCalls {
					return
				}
			}
		}
	}()

	// wait for call to Update or test failure
	select {
	case <-ctx.Done():
		t.Fatal("timed out")
	case <-read:
		require.Equal(t, wantUpdateCalls, gotUC)
	}
}
