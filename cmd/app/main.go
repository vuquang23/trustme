package main

import (
	"log"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"github.com/vuquang23/trustme/internal/pkg/config"
	"github.com/vuquang23/trustme/internal/pkg/parser"
	"github.com/vuquang23/trustme/internal/pkg/repository/subscriber"
	"github.com/vuquang23/trustme/internal/pkg/repository/tx"
	"github.com/vuquang23/trustme/pkg/logger"
	"golang.org/x/sync/errgroup"
)

func main() {
	app := &cli.App{
		Name: "Trustme",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "internal/pkg/config/default.yaml",
				Usage:   "Configuration file",
			},
		},
		Commands: []*cli.Command{
			{
				Name: "trustme",
				Action: func(c *cli.Context) error {
					conf := config.New()
					if err := conf.Load(c.String("config")); err != nil {
						return err
					}

					// logger
					_, err := logger.Init(conf.Log, logger.LoggerBackendZap)
					if err != nil {
						return err
					}

					errGroup, ctx := errgroup.WithContext(c.Context)

					rpcClient, err := ethclient.Dial("https://ethereum-rpc.publicnode.com")
					if err != nil {
						return err
					}

					wsClient, err := ethclient.Dial("wss://ethereum-rpc.publicnode.com")
					if err != nil {
						return err
					}

					subscriberRepo := subscriber.NewMemRepository()
					txRepo := tx.NewMemRepository()

					parser := parser.New(rpcClient, wsClient, subscriberRepo, txRepo)

					errGroup.Go(func() error { return parser.Run(ctx) })

					return errGroup.Wait()
				},
			},
		},
		DefaultCommand: "trustme",
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
