package model

import (
	"fmt"
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
	Project []string
	Action  RepoAction
}

type RepoAction struct {
	Required []string
	Allowed  []string
}

func LoadRepos(path string) (Repos, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %v", path, err)
	}

	var x Repos
	err = toml.Unmarshal(data, &x)
	if err != nil {
		return nil, fmt.Errorf("toml unmarshal: %v", err)
	}

	return x, nil
}
