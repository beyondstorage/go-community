package main

import (
	"github.com/urfave/cli/v2"

	"github.com/beyondstorage/go-community/env"
)

var repoCmd = &cli.Command{
	Name:  "repo",
	Usage: "maintain community repos",
	Subcommands: []*cli.Command{
		repoSyncActionsCmd,
	},
}

var repoSyncActionsCmd = &cli.Command{
	Name: "sync-actions",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "owner",
			Usage:    "github organization name",
			Required: true,
			EnvVars: []string{
				env.GithubOwner,
			},
		},
		&cli.StringFlag{
			Name:     "token",
			Usage:    "github access token",
			Required: true,
			EnvVars: []string{
				env.GithubAccessToken,
			},
		},
	},
	Action: func(c *cli.Context) (err error) {
		return nil
	},
}
