package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadProjects(t *testing.T) {
	x, err := LoadProjects("testdata/projects.toml")
	if err != nil {
		t.Fatal("load user", err)
	}

	assert.ElementsMatch(t, []string{"test-repo"}, x["test-team"].Repos)
}
