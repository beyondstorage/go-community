package main

import "github.com/urfave/cli/v2"

var syncCmd = &cli.Command{
	Name:   "sync",
	Usage:  "sync rooms",
	Action: sync,
}

func sync(context *cli.Context) error {
	return nil
}
