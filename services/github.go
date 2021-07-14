package services

import (
	"context"
	"fmt"
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
		model.RoleLeader:      "admin",
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

func (g *Github) ListRepos(ctx context.Context, org string) ([]string, error) {
	opt := &github.RepositoryListByOrgOptions{
		Type: "public",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	rs := make([]string, 0)
	for {
		repos, resp, err := g.client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			g.logger.Error("list repos", zap.Error(err))
			return nil, err
		}

		for _, v := range repos {
			if v.GetArchived() {
				g.logger.Info("ignore archived repo", zap.String("repo", v.GetName()))
			}
			rs = append(rs, v.GetName())
			g.logger.Info("repo", zap.String("repo", v.GetName()))
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return rs, nil
}

func (g *Github) SyncTeam(ctx context.Context, teams model.Teams) (err error) {
	err = g.setupTeams(ctx, teams)
	if err != nil {
		return
	}

	for tn, t := range teams {
		for _, role := range model.ValidRoles {
			slug := model.FormatTeamSlug(tn, role)

			// Sync repos
			for _, repo := range t.Repos {
				_, err = g.client.Teams.AddTeamRepoBySlug(
					ctx, g.owner, slug, g.owner, repo,
					&github.TeamAddTeamRepoOptions{Permission: permissionMap[role]})
				if err != nil {
					return fmt.Errorf("add team repo by slug: %w", err)
				}
				g.logger.Info("Added repo into team",
					zap.String("team", slug),
					zap.String("repo", repo))
			}
			// Sync members
			expectMembers := map[string]struct{}{}
			for _, member := range t.Members[role] {
				_, _, err = g.client.Teams.AddTeamMembershipBySlug(
					ctx, g.owner, slug, member, nil)
				if err != nil {
					return fmt.Errorf("add team repo by slug: %w", err)
				}

				expectMembers[member] = struct{}{}

				g.logger.Info("Added member into team",
					zap.String("team", slug),
					zap.String("member", member))
			}
			members, err := g.listTeamMembers(ctx, slug)
			if err != nil {
				return fmt.Errorf("list team members by slug: %w", err)
			}
			for _, member := range members {
				if _, ok := expectMembers[member]; ok {
					continue
				}
				// Remove all members that not in expect members.
				_, err = g.client.Teams.RemoveTeamMembershipBySlug(ctx, g.owner, slug, member)
				if err != nil {
					return fmt.Errorf("remove team member by slug: %w", err)
				}
				g.logger.Info("Removed member from team",
					zap.String("team", slug),
					zap.String("member", member))
			}
		}
	}
	return
}

func (g *Github) GenerateReport(ctx context.Context, org, repo string) (content string, err error) {
	events, err := g.listEvents(ctx, org, repo)
	if err != nil {
		return "", err
	}

	b := &strings.Builder{}
	// Add front line with repo name.
	// ## [repo](https://github.com/org/repo)
	b.WriteString(fmt.Sprintf("## [%s](https://github.com/%s/%s)\n\n", repo, org, repo))

	for _, v := range events {
		typ := v.GetType()

		raw, err := v.ParsePayload()
		if err != nil {
			return "", err
		}
		switch typ {
		case "IssuesEvent":
			e := raw.(*github.IssuesEvent)
			switch e.GetAction() {
			case "opened":
				b.WriteString(fmt.Sprintf("- @%s opened issue %s\n",
					v.GetActor().GetLogin(),
					e.GetIssue().GetHTMLURL()))
			case "closed":
				b.WriteString(fmt.Sprintf("- @%s closed issue %s\n",
					v.GetActor().GetLogin(),
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
				b.WriteString(fmt.Sprintf("- @%s opened pull request %s\n",
					v.GetActor().GetLogin(),
					e.GetPullRequest().GetHTMLURL()))
			case "closed":
				if e.GetPullRequest().GetMerged() {
					b.WriteString(fmt.Sprintf("- @%s merged pull request %s\n",
						v.GetActor().GetLogin(),
						e.GetPullRequest().GetHTMLURL()))
				} else {
					b.WriteString(fmt.Sprintf("- @%s closed pull request %s\n",
						v.GetActor().GetLogin(),
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

	return b.String(), nil
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
		for _, role := range model.ValidRoles {
			slug := teamName + "-" + role.String()

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
	}
	return nil
}
