package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadRepos(t *testing.T) {
	x, err := LoadRepos("testdata/repos.toml", []string{"test", "testx", "abc"})
	if err != nil {
		t.Fatal("load repos", err)
	}

	assert.ElementsMatch(t, []string{"test-project"}, x["test"].Project)
	assert.ElementsMatch(t, []string{"next-root"}, x["testx"].Project)
	assert.ElementsMatch(t, []string{"root"}, x["abc"].Project)
}

func TestRepos_ParsedProjects(t *testing.T) {
	x, err := LoadRepos("testdata/repos.toml", []string{"test", "testx", "abc"})
	if err != nil {
		t.Fatal("load repos", err)
	}
	p := x.ParsedProjects()

	assert.ElementsMatch(t, []string{"abc"}, p["root"])
}
