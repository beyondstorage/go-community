package main

import (
	"github.com/urfave/cli/v2"
)

var generateCmd = &cli.Command{
	Name:   "generate",
	Usage:  "generate rooms",
	Action: generate,
}

func generate(context *cli.Context) error {

	return nil
}
