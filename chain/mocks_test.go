package chain

import (
	"container/list"
	"errors"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/gcs"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightninglabs/neutrino"
	"github.com/lightninglabs/neutrino/headerfs"
)

var (
	ErrNotImplemented              = errors.New("not implemented")
	_                 Rescanner    = (*mockRescanner)(nil)
	_                 ChainService = (*mockChainService)(nil)
	testBestBlock                  = &headerfs.BlockStamp{
		Height: 42,
	}
)

func newMockNeutrinoClient(t *testing.T) *NeutrinoClient {
	t.Helper()
	var (
		chainParams  = &chaincfg.Params{}
		chainSvc     = &mockChainService{}
		newRescanner = func(ro ...neutrino.RescanOption) Rescanner {
			return &mockRescanner{
				updateArgs: list.New(),
			}
		}
	)
	return &NeutrinoClient{
		CS:           chainSvc,
		chainParams:  chainParams,
		newRescanner: newRescanner,
		rescanQuitCh: make(chan chan struct{}, 1),
		rescannerCh:  make(chan Rescanner, 1),
	}
}

type mockRescanner struct {
	updateArgs *list.List
}

func (m *mockRescanner) Start() <-chan error {
	return make(<-chan error)
}

func (m *mockRescanner) WaitForShutdown() {
	// no-op
}

func (m *mockRescanner) Update(opts ...neutrino.UpdateOption) error {
	m.updateArgs.PushBack(opts)
	return nil
}

type mockChainService struct{}

func (m *mockChainService) Start() error {
	return nil
}

func (m *mockChainService) GetBlock(
	chainhash.Hash,
	...neutrino.QueryOption,
) (*btcutil.Block, error) {
	return nil, ErrNotImplemented
}

func (m *mockChainService) GetBlockHeight(*chainhash.Hash) (int32, error) {
	return 0, ErrNotImplemented
}

func (m *mockChainService) BestBlock() (*headerfs.BlockStamp, error) {
	return m.getBestBlock(), nil
}

func (m *mockChainService) getBestBlock() *headerfs.BlockStamp {
	return testBestBlock
}

func (m *mockChainService) GetBlockHash(int64) (*chainhash.Hash, error) {
	return nil, ErrNotImplemented
}

func (m *mockChainService) GetBlockHeader(*chainhash.Hash) (*wire.BlockHeader, error) {
	return nil, ErrNotImplemented
}

func (m *mockChainService) IsCurrent() bool {
	return false
}

func (m *mockChainService) SendTransaction(*wire.MsgTx) error {
	return ErrNotImplemented
}

func (m *mockChainService) GetCFilter(
	chainhash.Hash,
	wire.FilterType,
	...neutrino.QueryOption,
) (*gcs.Filter, error) {
	return nil, ErrNotImplemented
}
