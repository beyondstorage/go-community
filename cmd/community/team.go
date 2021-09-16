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
			Name:     "teams",
			Usage:    "path to the teams.toml",
			Required: true,
			Value:    "teams.toml",
		},
		&cli.StringFlag{
			Name:     "repos",
			Usage:    "path to the repos.toml",
			Required: true,
			Value:    "repos.toml",
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
		ctx := context.Background()

		g, err := services.NewGithub(
			c.String("owner"),
			c.String("token"))
		if err != nil {
			return
		}

		team, err := model.LoadTeams(c.String("teams"))
		if err != nil {
			return
		}

		repos, err := model.LoadRepos(c.String("repos"))
		if err != nil {
			return err
		}

		githubRepos, err := g.ListRepos(ctx)
		if err != nil {
			return
		}

		err = g.SyncTeam(ctx, team, repos, githubRepos)
		if err != nil {
			return
		}

		err = g.SyncContributors(ctx, team, githubRepos)
		if err != nil {
			return
		}
		return nil
	},
}
