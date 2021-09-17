package main

import (
	"context"
	"github.com/beyondstorage/go-community/model"
	"github.com/beyondstorage/go-community/services"
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
		&cli.StringFlag{
			Name:     "actions",
			Usage:    "the folder of our actions",
			Required: true,
			EnvVars: []string{
				env.GithubActions,
			},
		},
		&cli.StringFlag{
			Name:     "repos",
			Usage:    "the folder of our repos",
			Required: true,
			EnvVars: []string{
				env.GithubRepos,
			},
		},
	},
	Action: func(c *cli.Context) (err error) {
		ctx := context.Background()

		g, err := services.NewGithub(
			c.String("owner"),
			c.String("token"))
		if err != nil {
			return
		}

		githubRepos, err := g.ListRepos(ctx)
		if err != nil {
			return
		}

		repos, err := model.LoadRepos(c.String("repos"), githubRepos)
		if err != nil {
			return err
		}

		err = g.SyncActions(ctx, c.String("actions"), repos)
		if err != nil {
			return
		}

		return nil
	},
}
