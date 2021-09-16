package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadRepos(t *testing.T) {
	x, err := LoadRepos("testdata/repos.toml")
	if err != nil {
		t.Fatal("load user", err)
	}

	assert.ElementsMatch(t, []string{"beyond-tp", "go-storage"}, x["site"].Project)
}
