package utils

import (
	"context"
	"github.com/google/go-github/v35/github"
	"go.uber.org/zap"
)

func Repos(org string) ([]string, error) {
	ctx := context.Background()

	logger, _ := zap.NewDevelopment()

	client := github.NewClient(nil)

	opt := &github.RepositoryListByOrgOptions{Type: "public"}
	repos, _, err := client.Repositories.ListByOrg(ctx, org, opt)
	if err != nil {
		logger.Error("list repos", zap.Error(err))
		return nil, err
	}

	rs := make([]string, 0)

	for _, v := range repos {
		if v.GetArchived() {
			logger.Info("ignore archived repo", zap.String("repo", v.GetName()))
		}
		rs = append(rs, v.GetName())
		logger.Info("repo", zap.String("repo", v.GetName()))
	}
	return rs, nil
}
