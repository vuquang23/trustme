package api

import "github.com/ethereum/go-ethereum/core/types"

type IParser interface {
	// last parsed block
	GetCurrentBlock() int

	// add address to observer
	Subscribe(address string) bool

	// list of inbound or outbound transactions for an address
	GetTransactions(address string) []*types.Transaction
}
