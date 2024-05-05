package parser

import (
	"github.com/ethereum/go-ethereum/core/types"
)

type ISubscriberRepository interface {
	Create(address string) error
	IsSubscriber(address string) bool
}

type ITxRepository interface {
	SaveTx(address string, tx *types.Transaction) error
	GetTxs(address string) ([]*types.Transaction, error)
}
