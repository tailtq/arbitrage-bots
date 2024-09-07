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
				Name: "pool-data-fetch",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "pool-data-temp",
						Usage: "pool data temp (containing pool addresses)",
					},
				},
				Action: func(ctx *cli.Context) error {
					var poolDataTemp = ctx.String("pool-data-temp")

					if poolDataTemp == "" {
						fmt.Println("Please provide both network and pool-data-temp")
						return nil
					}

					commands.FetchUniswapPoolData(poolDataTemp)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
