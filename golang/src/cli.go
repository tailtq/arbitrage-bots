package main

import (
	"arbitrage-bot/commands"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Commands: []cli.Command{
			{
				Name: "uniswap-fetch-pools",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "pool-data-temp",
						Usage: "pool data temp (containing pool addresses)",
					},
				},
				Action: func(ctx *cli.Context) {
					if poolDataTemp := ctx.String("pool-data-temp"); poolDataTemp != "" {
						var command = commands.NewFetchUniswapPoolDataCommand()
						command.Fetch(poolDataTemp)
					} else {
						fmt.Println("Please provide both network and pool-data-temp")
					}
				},
			},
			{
				Name: "pancakeswap-fetch-pools",
				Action: func(ctx *cli.Context) {
					var command = commands.NewFetchPancakeswapPoolDataCommand()
					command.Fetch()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
