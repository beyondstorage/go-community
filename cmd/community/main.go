package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:  "community",
	Usage: "community tools for open source society",
	Action: func(c *cli.Context) error {
		fmt.Println("boom! I say!")
		return nil
	},
	Commands: []*cli.Command{
		generateCmd,
		syncCmd,
	},
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
