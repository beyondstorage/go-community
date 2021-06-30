package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadUsers(t *testing.T) {
	x, err := LoadUsers("testdata/users.toml")
	if err != nil {
		t.Fatal("load user", err)
	}

	assert.Equal(t, "user@example.com", x["example"].Email)
}
