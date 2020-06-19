package internal

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type UserEnv struct {
	// "owner/repo"
	Repository string `envconfig:"REPOSITORY" required:"true"`
}

func LoadUserEnv() (*UserEnv, error) {
	env := UserEnv{}
	err := envconfig.Process("", &env)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load user env")
	}
	return &env, nil
}
