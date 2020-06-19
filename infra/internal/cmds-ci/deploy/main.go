package main

import (
	"log"

	"infra/internal"
)

type Variables struct {
	*internal.GitHubEnv
	*internal.Secrets
	*internal.BackendDeployEventPayload
}

func loadVariables() (*Variables, error) {
	githubEnv, err := internal.LoadGitHubEnv()
	if err != nil {
		return nil, err
	}

	secrets, err := internal.LoadSecrets()
	if err != nil {
		return nil, err
	}

	eventPayload, err := internal.LoadBackendDeployEventPayloadFromEnv()
	if err != nil {
		return nil, err
	}

	return &Variables{Secrets: secrets, GitHubEnv: githubEnv, BackendDeployEventPayload: eventPayload}, nil
}

func main() {
	vars, err := loadVariables()
	if err != nil {
		log.Fatal(err)
	}

	bu, err := internal.NewBuildUtils(vars.InfraRegion, vars.InfraBucket, vars.Service)
	if err != nil {
		log.Fatal(err)
	}

	zFPath, err := bu.DownloadDistZip(vars.Checksum)
	if err != nil {
		log.Fatal(err)
	}

	err = bu.Deploy(vars.Env, zFPath)
	if err != nil {
		log.Fatal(err)
	}
}
