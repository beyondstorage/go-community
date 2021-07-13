package main

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/beyondstorage/go-community/env"
	"github.com/beyondstorage/go-community/model"
	"github.com/beyondstorage/go-community/services"
)

var teamCmd = &cli.Command{
	Name:  "team",
	Usage: "maintain community teams",
	Subcommands: []*cli.Command{
		teamSyncCmd,
	},
}

var teamSyncCmd = &cli.Command{
	Name: "sync",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "path",
			Usage:    "path to the teams.toml",
			Required: true,
			Value:    "teams.toml",
		},
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
		g, err := services.NewGithub(
			c.String("owner"),
			c.String("token"))
		if err != nil {
			return
		}

		team, err := model.LoadTeams(c.String("path"))
		if err != nil {
			return
		}

		err = g.SyncTeam(context.Background(), team)
		if err != nil {
			return
		}
		return nil
	},
}
