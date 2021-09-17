package model

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Teams map[string]Team

type Team struct {
	Project string
	Role    Role
	Members []string
}

func LoadTeams(path string) (Teams, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %v", path, err)
	}

	var x Teams
	err = toml.Unmarshal(data, &x)
	if err != nil {
		return nil, fmt.Errorf("toml unmarshal: %v", err)
	}

	return x, nil
}
