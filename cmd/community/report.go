package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/beyondstorage/go-community/env"
	"github.com/beyondstorage/go-community/model"
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

		if c.String("type") != "issue" {
			logger.Error("not supported report type", zap.String("type", c.String("type")))
			return errors.New("not supported type")
		}

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

		sort.Strings(repos)
		b := &strings.Builder{}

		usernameDict := make(map[string]bool)
		statistics := make([]model.Statistic, 0, len(repos))

		for _, v := range repos {
			content, users, stat, err := g.GenerateReportDataByRepo(ctx, c.String("owner"), v)
			if err != nil {
				return nil
			}

			// skip generate content for repo whose statistic is blank
			if stat.IsBlank() {
				continue
			}

			// aggregate statistic data
			statistics = append(statistics, stat)

			// merge users into dict
			for name, exist := range users {
				usernameDict[name] = exist
			}

			b.WriteString(content)
		}

		// sum statistics as the result
		resStat := model.Statistics(statistics).Sum()

		// append users link after report content
		for username, _ := range usernameDict {
			// ## [@username]: https://github.com/username
			b.WriteString(fmt.Sprintf("[@%s]: https://github.com/%s\n", username, username))
		}

		// print statistics before report content
		result := fmt.Sprintf("%s\n%s\n", resStat.FormatPrint(), b.String())

		url, err := g.CreateWeeklyReportIssue(ctx, c.String("output"), result)
		if err != nil {
			return err
		}
		fmt.Printf("Create issue %s\n", url)
		return nil
	},
}
