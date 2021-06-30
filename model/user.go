package model

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Users map[string]User

type User struct {
	Email string `toml:"email"`
}

func LoadUsers(path string) (Users, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %v", path, err)
	}

	var x Users
	err = toml.Unmarshal(data, &x)
	if err != nil {
		return nil, fmt.Errorf("toml unmarshal: %v", err)
	}

	return x, nil
}
