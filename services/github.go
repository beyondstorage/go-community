package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v35/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/beyondstorage/go-community/model"
)

var (
	permissionMap = map[model.Role]string{
		model.RoleAdmin:       "admin",
		model.RoleMaintainer:  "maintain",
		model.RoleCommitter:   "push",
		model.RoleReviewer:    "triage",
		model.RoleContributor: "pull",
	}
)

type Github struct {
	owner string

	logger *zap.Logger
	client *github.Client
}

func NewGithub(owner, token string) (g *Github, err error) {
	logger, _ := zap.NewDevelopment()

	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	g = &Github{
		owner:  owner,
		logger: logger,
		client: github.NewClient(tc),
	}
	return
}

func (g *Github) ListRepos(ctx context.Context) ([]string, error) {
	opt := &github.RepositoryListByOrgOptions{
		Type: "public",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	rs := make([]string, 0)
	for {
		repos, resp, err := g.client.Repositories.ListByOrg(ctx, g.owner, opt)
		if err != nil {
			g.logger.Error("list repos", zap.Error(err))
			return nil, err
		}

		for _, v := range repos {
			if v.GetArchived() {
				g.logger.Info("ignore archived repo", zap.String("repo", v.GetName()))
			}
			rs = append(rs, v.GetName())
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	g.logger.Info("all repo has been listed")
	return rs, nil
}

func (g *Github) SyncTeam(ctx context.Context, teams model.Teams, repos model.Repos) (err error) {
	err = g.setupTeams(ctx, teams)
	if err != nil {
		return
	}

	projects := repos.ParsedProjects()
	for tn, t := range teams {
		expectRepos := make(map[string]struct{})
		for _, v := range projects[t.Project] {
			expectRepos[v] = struct{}{}
		}

		existRepos := make(map[string]struct{})
		opt := &github.ListOptions{
			PerPage: 100,
		}
		for {
			rps, resp, err := g.client.Teams.ListTeamReposBySlug(ctx, g.owner, tn, opt)
			if err != nil {
				g.logger.Error("list team githubRepos", zap.Error(err))
				return err
			}
			for _, v := range rps {
				existRepos[v.GetName()] = struct{}{}
			}
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		// Add githubRepos that in expectRepos but not in existRepos.
		for er := range expectRepos {
			_, exist := existRepos[er]
			if exist {
				continue
			}
			_, err = g.client.Teams.AddTeamRepoBySlug(
				ctx, g.owner, tn, g.owner, er,
				&github.TeamAddTeamRepoOptions{Permission: permissionMap[t.Role]})
			if err != nil {
				return fmt.Errorf("add team repo by slug: %w", err)
			}
			g.logger.Info("Added repo into team",
				zap.String("team", tn),
				zap.String("repo", er))
		}

		// Delete githubRepos that in existRepos but not in expectRepos.
		for er := range existRepos {
			_, exist := expectRepos[er]
			if exist {
				continue
			}
			_, err = g.client.Teams.RemoveTeamRepoBySlug(
				ctx, g.owner, tn, g.owner, er)
			if err != nil {
				return fmt.Errorf("remove team repo by slug: %w", err)
			}
			g.logger.Info("Removed repo into team",
				zap.String("team", tn),
				zap.String("repo", er))
		}

		expectMembers := make(map[string]struct{})
		for _, v := range t.Members {
			expectMembers[v] = struct{}{}
		}

		existMembers := make(map[string]struct{})
		teamopt := &github.TeamListTeamMembersOptions{
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}
		for {
			rps, resp, err := g.client.Teams.ListTeamMembersBySlug(ctx, g.owner, tn, teamopt)
			if err != nil {
				g.logger.Error("list team members", zap.Error(err))
				return err
			}
			for _, v := range rps {
				existMembers[v.GetLogin()] = struct{}{}
			}
			if resp.NextPage == 0 {
				break
			}
			teamopt.Page = resp.NextPage
		}

		// Add members that in expectMembers but not in existMembers.
		for m := range expectMembers {
			_, exist := existMembers[m]
			if exist {
				continue
			}
			_, _, err = g.client.Teams.AddTeamMembershipBySlug(
				ctx, g.owner, tn, m, nil)
			if err != nil {
				return fmt.Errorf("add team member by slug: %w", err)
			}
			g.logger.Info("Added member into team",
				zap.String("team", tn),
				zap.String("member", m))
		}

		// Delete members that in existMembers but not in expectMembers.
		for m := range existMembers {
			_, exist := expectMembers[m]
			if exist {
				continue
			}
			_, err = g.client.Teams.RemoveTeamMembershipBySlug(
				ctx, g.owner, tn, m)
			if err != nil {
				return fmt.Errorf("remove team member by slug: %w", err)
			}
			g.logger.Info("Removed member from team",
				zap.String("team", tn),
				zap.String("member", m))
		}
	}
	return
}

func (g *Github) SyncContributors(ctx context.Context, teams model.Teams, repos []string) (err error) {
	// All members in team.
	teamMembers := make(map[string]struct{})
	for _, team := range teams {
		for _, v := range team.Members {
			teamMembers[v] = struct{}{}
		}
	}

	// All members in org.
	existMembers := make(map[string]struct{})
	opt := &github.ListMembersOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	for {
		rps, resp, err := g.client.Organizations.ListMembers(ctx, g.owner, opt)
		if err != nil {
			g.logger.Info("list team members", zap.Error(err))
			return err
		}
		for _, v := range rps {
			existMembers[v.GetLogin()] = struct{}{}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// A map about <Github Login> -> <Github ID>
	expectMembers := make(map[string]int64)

	// List all contributors
	for _, repo := range repos {
		opt := &github.ListContributorsOptions{
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}
		for {
			contributors, resp, err := g.client.Repositories.ListContributors(ctx, g.owner, repo, opt)
			if err != nil {
				return fmt.Errorf("list contributors: %w", err)
			}
			for _, v := range contributors {
				expectMembers[v.GetLogin()] = v.GetID()
			}
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
	}

	// Add all members that not in org and team.
	for v, id := range expectMembers {
		_, exist := teamMembers[v]
		if exist {
			continue
		}
		_, exist = existMembers[v]
		if exist {
			continue
		}
		// We will ignore all bot account.
		if g.isBot(v) {
			continue
		}
		_, _, err = g.client.Organizations.CreateOrgInvitation(ctx, g.owner, &github.CreateOrgInvitationOptions{
			InviteeID: github.Int64(id),
			Role:      github.String("direct_member"),
			TeamID:    []int64{},
		})
		if err != nil {
			g.logger.Error("create invite for org",
				zap.String("login", v),
				zap.Int64("id", id))
			return fmt.Errorf("create invite: %w", err)
		}
	}
	return nil
}

func (g *Github) SyncActions(ctx context.Context, actionPath string, repos model.Repos) (err error) {
	for _, repo := range repos {
		if len(repo.Action.Required) == 0 {
			g.logger.Info("repo doesn't have required actions, ignore",
				zap.String("repo", repo.Name))
			continue
		}

		dc, err := g.listActions(ctx, repo.Name)
		if err != nil {
			return err
		}

		// fileToAdd will be filled all required actions.
		// If check passed, we remove it.
		// If check failed, we update with old files sha.
		fileToAdd := make(map[string]string)
		for _, v := range repo.Action.Required {
			fileToAdd[v] = ""
		}
		fileToRemove := make(map[string]string)

		for _, file := range dc {
			basename := strings.TrimSuffix(file.GetName(), ".yml")

			// We will keep all allowed actions untouched.
			if repo.Action.IsAllowed(basename) {
				g.logger.Info("ignore allowed actions",
					zap.String("repo", repo.Name),
					zap.String("name", basename))
				continue
			}
			// Check action required actions.
			if repo.Action.IsRequired(basename) {
				// Check file content.
				ra, err := file.GetContent()
				if err != nil {
					g.logger.Error("get repo content", zap.Error(err))
					return err
				}
				// Read local action files.
				actionFile := fmt.Sprintf("%s/%s.yml", actionPath, basename)
				bs, err := ioutil.ReadFile(actionFile)
				if err != nil {
					g.logger.Error("read local actions",
						zap.String("path", actionFile), zap.Error(err))
					return err
				}
				if ra == string(bs) {
					delete(fileToAdd, basename)
					g.logger.Info("action is in sync, ignore",
						zap.String("repo", repo.Name),
						zap.String("name", basename))
				} else {
					fileToAdd[basename] = file.GetSHA()
					g.logger.Info("action is out of sync, prepare update",
						zap.String("repo", repo.Name),
						zap.String("name", basename))
				}
				continue
			}
			// Other actions should be removed.
			fileToRemove[basename] = file.GetSHA()
		}

		if len(fileToRemove) == 0 && len(fileToAdd) == 0 {
			g.logger.Info("all actions are in sync, finished")
			continue
		}

		baseref, _, err := g.client.Git.GetRef(ctx, g.owner, repo.Name, "heads/master")
		if err != nil {
			g.logger.Error("get base ref", zap.Error(err))
			return err
		}

		newBranch := fmt.Sprintf("sync-actions-%d", time.Now().Unix())
		newref, _, err := g.client.Git.CreateRef(ctx, g.owner, repo.Name, &github.Reference{
			Ref:    github.String("heads/" + newBranch),
			Object: baseref.Object,
		})
		if err != nil {
			g.logger.Error("create new ref", zap.Error(err))
			return err
		}

		// Remove file that need to be removed.
		for filename, sha := range fileToRemove {
			sha := sha

			filePath := fmt.Sprintf(".github/workflows/%s.yml", filename)
			_, _, err = g.client.Repositories.DeleteFile(ctx, g.owner, repo.Name, filePath, &github.RepositoryContentFileOptions{
				Message:   github.String("Delete not allowed file: " + filePath),
				SHA:       &sha,
				Branch:    github.String(newBranch),
				Author:    g.getCommitter(),
				Committer: g.getCommitter(),
			})
			if err != nil {
				g.logger.Error("delete file", zap.Error(err))
				return err
			}
			g.logger.Info("remove not allowed files",
				zap.String("repo", repo.Name),
				zap.String("name", filePath))
		}

		for filename, sha := range fileToAdd {
			sha := sha

			actionFile := fmt.Sprintf("%s/%s.yml", actionPath, filename)
			bs, err := ioutil.ReadFile(actionFile)
			if err != nil {
				g.logger.Error("read local actions",
					zap.String("path", actionFile), zap.Error(err))
				return err
			}

			filePath := fmt.Sprintf(".github/workflows/%s.yml", filename)

			rcf := &github.RepositoryContentFileOptions{
				Message:   github.String("Add new file: " + filePath),
				Content:   bs,
				Branch:    github.String(newBranch),
				Author:    g.getCommitter(),
				Committer: g.getCommitter(),
			}
			if sha != "" {
				// If sha is empty, we are creating files, or we are updating files.
				rcf.SHA = github.String(sha)
			}

			_, _, err = g.client.Repositories.CreateFile(ctx, g.owner, repo.Name, filePath, rcf)
			if err != nil {
				g.logger.Error("write new files",
					zap.String("path", filePath),
					zap.String("branch", newBranch),
					zap.Error(err))
				return err
			}

			g.logger.Info("add required files",
				zap.String("repo", repo.Name),
				zap.String("path", filePath))
		}

		_, _, err = g.client.PullRequests.Create(ctx, g.owner, repo.Name, &github.NewPullRequest{
			Title:               github.String("ci: Sync github actions"),
			Head:                newref.Ref,
			Base:                baseref.Ref,
			MaintainerCanModify: github.Bool(true),
		})
		if err != nil {
			g.logger.Error("create pull request", zap.Error(err))
			return err
		}
	}

	return nil
}

func (g *Github) GenerateReportDataByRepo(ctx context.Context, org, repo string) (
	content string, users map[string]bool, stat model.Statistic, err error) {
	users = make(map[string]bool)
	events, err := g.listEvents(ctx, org, repo)
	if err != nil {
		return "", users, stat, err
	}

	b := &strings.Builder{}
	// Add front line with repo name.
	// ## [repo](https://github.com/org/repo)
	b.WriteString(fmt.Sprintf("## [%s](https://github.com/%s/%s)\n\n", repo, org, repo))

	for _, v := range events {
		// skip events committed by bot
		if g.isBot(v.GetActor().GetLogin()) {
			continue
		}
		// add user info into dict
		users[v.GetActor().GetLogin()] = true

		raw, err := v.ParsePayload()
		if err != nil {
			return "", users, stat, err
		}
		switch v.GetType() {
		case "IssuesEvent":
			e := raw.(*github.IssuesEvent)
			switch e.GetAction() {
			case "opened":
				stat.CountIssueOpen()
				b.WriteString(fmt.Sprintf("- [@%s] opened issue [%s](%s)\n",
					v.GetActor().GetLogin(),
					e.GetIssue().GetTitle(),
					e.GetIssue().GetHTMLURL()))
			case "closed":
				stat.CountIssueClose()
				b.WriteString(fmt.Sprintf("- [@%s] closed issue [%s](%s)\n",
					v.GetActor().GetLogin(),
					e.GetIssue().GetTitle(),
					e.GetIssue().GetHTMLURL()))
			default:
				g.logger.Info("ignore issue",
					zap.String("repo", repo),
					zap.String("action", e.GetAction()))
				continue
			}
		case "PullRequestEvent":
			e := raw.(*github.PullRequestEvent)
			switch e.GetAction() {
			case "opened":
				stat.CountPROpen()
				b.WriteString(fmt.Sprintf("- [@%s] opened pull request [%s](%s)\n",
					v.GetActor().GetLogin(),
					e.GetPullRequest().GetTitle(),
					e.GetPullRequest().GetHTMLURL()))
			case "closed":
				stat.CountPRClose()
				if e.GetPullRequest().GetMerged() {
					b.WriteString(fmt.Sprintf("- [@%s] merged pull request [%s](%s)\n",
						v.GetActor().GetLogin(),
						e.GetPullRequest().GetTitle(),
						e.GetPullRequest().GetHTMLURL()))
				} else {
					b.WriteString(fmt.Sprintf("- [@%s] closed pull request [%s](%s)\n",
						v.GetActor().GetLogin(),
						e.GetPullRequest().GetTitle(),
						e.GetPullRequest().GetHTMLURL()))
				}
			default:
				g.logger.Info("ignore pull request",
					zap.String("repo", repo),
					zap.String("action", e.GetAction()))
				continue
			}
		default:
			panic("invalid event type")
		}
	}

	// Add trailing empty line.
	b.WriteString("\n")

	return b.String(), users, stat, nil
}

func (g *Github) CreateWeeklyReportIssue(ctx context.Context, repo, content string) (issueURL string, err error) {
	issue, _, err := g.client.Issues.Create(ctx, g.owner, repo, &github.IssueRequest{
		Title: github.String(fmt.Sprintf("Weekly report since %s", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))),
		Body:  github.String(content),
	})
	if err != nil {
		return
	}
	return issue.GetHTMLURL(), nil
}

func (g *Github) CreateIssue(ctx context.Context, repo, title, content string) (issueURL string, err error) {
	issue, _, err := g.client.Issues.Create(ctx, g.owner, repo, &github.IssueRequest{
		Title: github.String(title),
		Body:  github.String(content),
	})
	if err != nil {
		return
	}
	return issue.GetHTMLURL(), nil
}

func (g *Github) listTeams(ctx context.Context) (teams map[string]*github.Team, err error) {
	opt := &github.ListOptions{
		PerPage: 100,
	}

	teams = make(map[string]*github.Team)
	for {
		ts, resp, err := g.client.Teams.ListTeams(ctx, g.owner, opt)
		if err != nil {
			return nil, fmt.Errorf("github list teams: %w", err)
		}
		for _, v := range ts {
			v := v
			teams[v.GetName()] = v
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return teams, nil
}

func (g *Github) listEvents(ctx context.Context, org, repo string) (es []*github.Event, err error) {
	opt := &github.ListOptions{
		PerPage: 100,
	}

	for {
		events, resp, err := g.client.Activity.ListRepositoryEvents(ctx, org, repo, opt)
		if err != nil {
			g.logger.Error("list events", zap.Error(err))
			return nil, err
		}

		for _, v := range events {
			v := v

			// Ignore all private events.
			if !v.GetPublic() {
				continue
			}

			// Ignore all events that happens before 7 days ago.
			expectedSince := time.Now().AddDate(0, 0, -7)
			createdAt := v.GetCreatedAt()
			if createdAt.Before(expectedSince) {
				continue
			}

			typ := v.GetType()
			switch typ {
			case "IssuesEvent", "PullRequestEvent":
				es = append(es, v)
			default:
				// Ignore all events except issues and PRs.
				g.logger.Debug("ignore events", zap.String("type", typ))
				continue
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Sort all events.
	sort.SliceStable(es, func(i, j int) bool {
		return es[i].GetCreatedAt().Before(es[j].GetCreatedAt())
	})
	return
}

func (g *Github) listTeamMembers(ctx context.Context, team string) (users []string, err error) {
	opt := &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		ts, resp, err := g.client.Teams.ListTeamMembersBySlug(ctx, g.owner, team, opt)
		if err != nil {
			return nil, fmt.Errorf("github list teams: %w", err)
		}
		for _, v := range ts {
			users = append(users, v.GetLogin())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return users, nil
}

func (g *Github) setupTeams(ctx context.Context, teams model.Teams) (err error) {
	for teamName := range teams {
		slug := teamName

		_, resp, err := g.client.Teams.GetTeamBySlug(ctx, g.owner, slug)
		if err == nil {
			// The team is exist, we can continue to setup next team.
			continue
		}

		if resp.StatusCode != 404 {
			// This error is not a valid github error, return directly.
			return fmt.Errorf("get team by slug %s: %v", slug, err)
		}

		privacy := "closed" // open to all team members.
		// Now we can handle the create team logic.
		_, _, err = g.client.Teams.CreateTeam(ctx, g.owner, github.NewTeam{
			Name:    slug,
			Privacy: &privacy,
		})
		if err != nil {
			return fmt.Errorf("create team slug %s: %v", slug, err)
		}
	}
	return nil
}

func (g *Github) listActions(ctx context.Context, repo string) (dc []*github.RepositoryContent, err error) {
	_, idc, _, err := g.client.Repositories.GetContents(ctx, g.owner, repo, ".github/workflows", nil)
	if err != nil {
		g.logger.Error("get folder",
			zap.String("repo", repo),
			zap.String("path", ".github/workflows"),
			zap.Error(err))
		return nil, err
	}

	dc = make([]*github.RepositoryContent, 0)
	for _, file := range idc {
		if file.GetType() != "file" || !strings.HasSuffix(file.GetName(), ".yml") {
			continue
		}
		fc, _, _, err := g.client.Repositories.GetContents(ctx, g.owner, repo, file.GetPath(), nil)
		if err != nil {
			g.logger.Error("get file",
				zap.String("repo", repo),
				zap.String("path", file.GetPath()),
				zap.Error(err))
			return nil, err
		}
		dc = append(dc, fc)
	}
	return
}

func (g *Github) getCommitter() *github.CommitAuthor {
	now := time.Now()

	return &github.CommitAuthor{
		Date: &now,
		// TODO: we need to get from env
		Name:  github.String("BeyondRobot"),
		Email: github.String("robot@beyondstorage.io"),
		Login: github.String("BeyondRobot"),
	}
}

func (g *Github) isBot(login string) bool {
	switch login {
	case "dependabot[bot]", "github-actions[bot]", "BeyondRobot", "dependabot-preview[bot]", "gitter-badger", "renovate[bot]":
		return true
	default:
		return false
	}
}
