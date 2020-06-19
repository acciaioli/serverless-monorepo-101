package internal

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type GitHubSecrets struct {
	// The region for everything deployment related
	InfraRegion string `envconfig:"INFRA_AWS_REGION" required:"true"`
	// The bucket use to store deployment related state.
	InfraBucket string `envconfig:"INFRA_AWS_S3_BUCKET" required:"true"`
	// Github Personal Access Token
	// (https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line)
	PersonalAccessToken string `envconfig:"PERSONAL_ACCESS_TOKEN" required:"true"`
}

func LoadSecrets() (*GitHubSecrets, error) {
	env := GitHubSecrets{}
	err := envconfig.Process("", &env)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load secrets")
	}
	return &env, nil
}
