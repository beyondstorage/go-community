package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/gobwas/glob"
	"github.com/urfave/cli/v2"

	"github.com/beyondstorage/go-community/env"
	"github.com/beyondstorage/go-community/services"
)

var trackCmd = &cli.Command{
	Name:  "track",
	Usage: "maintain community tracking issues",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "title",
			Usage: "the tracking issue title",
		},
		&cli.StringFlag{
			Name:  "path",
			Usage: "the tracking issue content path",
		},
		&cli.StringFlag{
			Name:  "repo",
			Usage: "the tracking repos, support glob style like go-service-*",
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
		owner := c.String("owner")

		g, err := services.NewGithub(
			owner,
			c.String("token"))
		if err != nil {
			return
		}

		ctx := context.Background()

		repoGlob, err := glob.Compile(c.String("repo"))
		if err != nil {
			return fmt.Errorf("invalid glob pattern: %s", c.String("repo"))
		}

		content, err := ioutil.ReadFile(c.String("path"))
		if err != nil {
			return err
		}

		repos, err := g.ListRepos(ctx)
		if err != nil {
			return err
		}

		for _, v := range repos {
			// Ignore all non-matched repos.
			if !repoGlob.Match(v) {
				continue
			}

			url, err := g.CreateIssue(ctx, v, c.String("title"), string(content))
			if err != nil {
				return err
			}
			fmt.Printf("Created issue %s\n", url)
		}
		return nil
	},
}
