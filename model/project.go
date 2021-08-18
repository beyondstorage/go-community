package model

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

type Projects map[string]Project

type Project struct {
	Repos []string
}

func LoadProjects(path string) (Projects, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %v", path, err)
	}

	var x Projects
	err = toml.Unmarshal(data, &x)
	if err != nil {
		return nil, fmt.Errorf("toml unmarshal: %v", err)
	}

	return x, nil
}
