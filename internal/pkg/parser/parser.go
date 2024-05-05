package parser

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/sync/errgroup"

	"github.com/vuquang23/trustme/pkg/logger"
)

type Parser struct {
	rpcClient *ethclient.Client
	wsClient  *ethclient.Client

	currentBlock atomic.Int64

	subscriberRepo ISubscriberRepository
	txRepo         ITxRepository

	blockHashChan chan common.Hash
}

func New(rpcClient, wsClient *ethclient.Client, subscriberRepo ISubscriberRepository, txRepo ITxRepository) *Parser {
	return &Parser{
		rpcClient:      rpcClient,
		wsClient:       wsClient,
		subscriberRepo: subscriberRepo,
		txRepo:         txRepo,
		blockHashChan:  make(chan common.Hash, 10),
	}
}

func (p *Parser) Run(ctx context.Context) error {
	errgroup, ctx := errgroup.WithContext(ctx)

	errgroup.Go(func() error { return p.listenBlocks(ctx) })
	errgroup.Go(func() error { return p.handleBlocks(ctx) })

	return errgroup.Wait()
}

func (p *Parser) listenBlocks(ctx context.Context) error {
	f := func() error {
		headers := make(chan *types.Header)
		sub, err := p.wsClient.SubscribeNewHead(ctx, headers)
		if err != nil {
			return err
		}

		for {
			select {
			case err := <-sub.Err():
				return err
			case header := <-headers:
				p.currentBlock.Store(header.Number.Int64())
				p.blockHashChan <- header.Hash()
			}
		}
	}
	for {
		logger.Info(ctx, "listen new blocks...")

		if err := f(); err != nil {
			logger.Errorf(ctx, err.Error())
		}

		time.Sleep(3 * time.Second)
	}
}

func (p *Parser) handleBlocks(ctx context.Context) error {
	for h := range p.blockHashChan {
		logger.WithFields(ctx, logger.Fields{"hash": h.Hex()}).Info("new block")

		if err := p.handleBlock(ctx, h); err != nil {
			logger.WithFields(ctx, logger.Fields{
				"hash":     h.Hex(),
				"errorMsg": err.Error(),
			}).Warn("failed to handle block")
		}
	}

	return nil
}

func (p *Parser) handleBlock(ctx context.Context, hash common.Hash) error {
	block, err := p.rpcClient.BlockByHash(ctx, hash)
	if err != nil {
		return err
	}

	logger.WithFields(ctx, logger.Fields{
		"hash": hash.Hex(),
	}).Info("get block successfully")

	for _, tx := range block.Transactions() {
		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			return err
		}

		var (
			subscriber string
			fromStr    = strings.ToLower(from.Hex())
			toStr      = strings.ToLower(tx.To().Hex())
		)

		if p.subscriberRepo.IsSubscriber(fromStr) {
			subscriber = fromStr
		} else if p.subscriberRepo.IsSubscriber(toStr) {
			subscriber = toStr
		}

		if subscriber == "" {
			continue
		}

		if err := p.txRepo.SaveTx(subscriber, tx); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) GetCurrentBlock() int {
	return int(p.currentBlock.Load())
}

func (p *Parser) Subscribe(address string) bool {
	if p.subscriberRepo.IsSubscriber(address) {
		return false
	}

	p.subscriberRepo.Create(address)

	return true
}

func (p *Parser) GetTransactions(address string) []*types.Transaction {
	txs, _ := p.txRepo.GetTxs(address)
	return txs
}
