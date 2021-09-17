package model

import (
	"fmt"
	"github.com/gobwas/glob"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Repos map[string]Repo

func (r Repos) ParsedProjects() map[string][]string {
	m := make(map[string][]string)

	for name, repo := range r {
		for _, project := range repo.Project {
			m[project] = append(m[project], name)
		}
	}
	return m
}

type Repo struct {
	Name    string
	Project []string
	Action  RepoAction

	// We use from to record where the repo config from.
	// If from is shorter than name, we can overwrite the data safely.
	from string
}

type RepoAction struct {
	Required []string
	Allowed  []string

	required map[string]struct{}
	allowed  map[string]struct{}
}

func (ra *RepoAction) IsRequired(name string) bool {
	_, ok := ra.required[name]
	return ok
}

func (ra *RepoAction) IsAllowed(name string) bool {
	_, ok := ra.allowed[name]
	return ok
}

// LoadRepos will following the neatest overwrite rule.
func LoadRepos(path string, githubRepos []string) (Repos, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %v", path, err)
	}

	var x Repos
	err = toml.Unmarshal(data, &x)
	if err != nil {
		return nil, fmt.Errorf("toml unmarshal: %v", err)
	}

	// Parsed into repos
	repos := make(Repos)
	for patternName, repo := range x {
		repo.Action.required = make(map[string]struct{})
		for _, v := range repo.Action.Required {
			repo.Action.required[v] = struct{}{}
		}

		repo.Action.allowed = make(map[string]struct{})
		for _, v := range repo.Action.Allowed {
			repo.Action.allowed[v] = struct{}{}
		}

		g := glob.MustCompile(patternName)
		for _, repoName := range githubRepos {
			if !g.Match(repoName) {
				continue
			}
			// If the repo is not recorded, we can save directly.
			if _, ok := repos[repoName]; !ok {
				repo.from = patternName
				repo.Name = repoName
				repos[repoName] = repo
				continue
			}
			// If repos from is shorted than current matched repo, we need to update it.
			if len(repos[repoName].from) < len(patternName) {
				repo.from = patternName
				repo.Name = repoName
				repos[repoName] = repo
				continue
			}
			continue
		}
	}

	return repos, nil
}
