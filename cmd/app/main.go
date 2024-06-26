package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"

	"github.com/vuquang23/trustme/internal/pkg/api"
	"github.com/vuquang23/trustme/internal/pkg/config"
	"github.com/vuquang23/trustme/internal/pkg/parser"
	"github.com/vuquang23/trustme/internal/pkg/repository/subscriber"
	"github.com/vuquang23/trustme/internal/pkg/repository/tx"
	"github.com/vuquang23/trustme/internal/pkg/server"
	"github.com/vuquang23/trustme/pkg/logger"
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

					ctx, cancel := context.WithCancel(context.Background())
					go func() {
						sigs := make(chan os.Signal, 1)
						signal.Notify(sigs, unix.SIGTERM, unix.SIGINT)
						<-sigs
						cancel()
					}()

					// eth clients
					rpcClient, err := ethclient.Dial("https://ethereum-rpc.publicnode.com")
					if err != nil {
						return err
					}

					wsClient, err := ethclient.Dial("wss://ethereum-rpc.publicnode.com")
					if err != nil {
						return err
					}

					// repositories
					subscriberRepo := subscriber.NewMemRepository()
					txRepo := tx.NewMemRepository()

					// parser
					parser := parser.New(rpcClient, wsClient, subscriberRepo, txRepo)

					// http server
					engine := server.GinEngine(conf.Http, conf.Log, logger.LoggerBackendZap)
					api.SetupRoute(engine, parser)

					// run goroutines
					var errGroup errgroup.Group
					errGroup.Go(func() error { return parser.Run(ctx) })
					errGroup.Go(func() error {
						return server.Run(ctx, conf.Http.BindAddress, engine)
					})

					err = errGroup.Wait()
					if err != nil && err != context.Canceled {
						return err
					}

					logger.Info(ctx, "shutdown!")

					return nil
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
