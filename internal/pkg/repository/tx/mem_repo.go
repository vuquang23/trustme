package tx

import (
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
)

type MemRepository struct {
	data sync.Map
}

func NewMemRepository() *MemRepository {
	return &MemRepository{
		data: sync.Map{},
	}
}

func (t *MemRepository) SaveTx(address string, tx *types.Transaction) error {
	txs, _ := t.GetTxs(address)
	txs = append(txs, tx)
	t.data.Store(address, txs)
	return nil
}

func (t *MemRepository) GetTxs(address string) ([]*types.Transaction, error) {
	txs, ok := t.data.Load(address)
	if !ok {
		txs = []*types.Transaction{}
	}
	return txs.([]*types.Transaction), nil
}
