package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTeams(t *testing.T) {
	x, err := LoadTeams("testdata/teams.toml")
	if err != nil {
		t.Fatal("load user", err)
	}

	assert.ElementsMatch(t, []string{"specs"}, x["specs"].Repos)
	assert.ElementsMatch(t, []string{"example"}, x["specs"].Members[RoleLeader])
}
