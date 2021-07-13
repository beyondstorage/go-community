package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/beyondstorage/go-community/env"
	"github.com/beyondstorage/go-community/services"
)

var reportCmd = &cli.Command{
	Name:  "report",
	Usage: "maintain community reports",
	Subcommands: []*cli.Command{
		reportWeeklyCmd,
	},
}

var reportWeeklyCmd = &cli.Command{
	Name: "weekly",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "type",
			Usage:    "type of report",
			Required: true,
			Value:    "issue",
		},
		&cli.StringFlag{
			Name:     "output",
			Usage:    "destination of report",
			Required: true,
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
	Action: func(c *cli.Context) error {
		logger, _ := zap.NewDevelopment()

		g, err := services.NewGithub(
			c.String("owner"),
			c.String("token"))
		if err != nil {
			return err
		}

		ctx := context.Background()

		repos, err := g.ListRepos(ctx, c.String("owner"))
		if err != nil {
			return err
		}

		b := &strings.Builder{}

		for _, v := range repos {
			content, err := g.GenerateReport(ctx, c.String("owner"), v)
			if err != nil {
				return nil
			}
			b.WriteString(content)
		}

		if c.String("type") != "issue" {
			logger.Error("not supported report type", zap.String("type", c.String("type")))
			return errors.New("not supported type")
		}

		url, err := g.CreateWeeklyReportIssue(ctx, c.String("output"), b.String())
		if err != nil {
			return err
		}
		fmt.Printf("Create issue %s\n", url)
		return nil
	},
}
