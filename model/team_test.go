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

	team := x["pmc"]
	assert.ElementsMatch(t, []string{"*"}, team.Repos)
	assert.Equal(t, RoleAdmin, team.Role)

	team = x["go-storage-maintainer"]
	assert.Equal(t, "go-storage", team.Project)
	assert.Equal(t, RoleMaintainer, team.Role)
	assert.ElementsMatch(t, []string{"test-user"}, team.Members)
}
