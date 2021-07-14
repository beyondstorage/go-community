package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:  "community",
	Usage: "community tools for open source society",
	Commands: []*cli.Command{
		teamCmd,
		reportCmd,
		trackCmd,
	},
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
