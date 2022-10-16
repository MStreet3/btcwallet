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
		called                  = make(chan struct{})
		nc                      = newMockNeutrinoClient(t)
		wantNotifyReceivedCalls = 4
		wantUpdateCalls         = wantNotifyReceivedCalls - 1
	)
	t.Cleanup(cancel)

	go func() {
		defer close(sent)
		for i := 0; i < wantNotifyReceivedCalls; i++ {
			err := nc.NotifyReceived(addrs)
			require.NoError(t, err)

			// signal that NotifyReceived was called on first iteration
			if i == 0 {
				close(called)
			}
		}
	}()

	// wait for call to Update or test failure
	select {
	case <-ctx.Done():
		t.Fatal("timed out")
	case <-sent:
		rescanner := <-nc.rescannerCh
		mockRescan := rescanner.(*mockRescanner)
		require.Equal(t, wantUpdateCalls, mockRescan.updateArgs.Len())
	}
}

// TestNeutrinoClientNotifyReceivedRescan verifies concurrent calls to NotifyReceived and Rescan
// do not result in a data race and that there is no panic on replacing the Rescanner.
func TestNeutrinoClientNotifyReceivedRescan(t *testing.T) {
	var (
		ctx, cancel  = context.WithTimeout(context.Background(), 1*time.Second)
		addrs        []btcutil.Address
		sent         = make(chan struct{})
		nc           = newMockNeutrinoClient(t)
		wantRoutines = 100
	)

	t.Cleanup(cancel)

	err := nc.Start()
	require.NoError(t, err)

	go func() {
		defer close(sent)
		for i := 0; i < wantRoutines; i++ {
			if i%3 == 0 {
				go func() {
					rerr := nc.Rescan(nil, addrs, nil)
					require.NoError(t, rerr)
				}()
			}
			go func() {
				err := nc.NotifyReceived(addrs)
				require.NoError(t, err)
			}()
		}
	}()

	select {
	case <-ctx.Done():
		t.Fatal("timed out")
	case <-sent:
	}
}
