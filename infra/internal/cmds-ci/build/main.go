package main

import (
	"context"
	"fmt"
	"log"

	"infra/internal"
)

type Variables struct {
	*internal.GitHubEnv
	*internal.Secrets
	*internal.BackendBuildEventPayload
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

	eventPayload, err := internal.LoadBackendBuildEventPayloadFromEnv()
	if err != nil {
		return nil, err
	}

	return &Variables{Secrets: secrets, GitHubEnv: githubEnv, BackendBuildEventPayload: eventPayload}, nil
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

	checksum, err := bu.ComputeCodeChecksum()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(fmt.Sprintf("code checksum: %s", checksum))

	zData, err := bu.GenerateDistZip()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("dist zip generated")

	log.Print("uploading dist zip")
	err = bu.UploadDistZip(checksum, zData)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("dist zip uploaded")

	log.Print("updating last checksum")
	err = bu.SetLastCodeChecksum(checksum)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("last checksum updated")

	log.Print("triggering deploy event")
	env := "dev" // deploy on dev automatically
	githubClient, err := internal.NewGitHubClient(vars.GitHubRepository, vars.PersonalAccessToken)
	if err != nil {
		log.Fatal(err)
	}
	eventType := internal.BackendDeployEventType(vars.Service, env)
	eventPayload := internal.BackendDeployEventPayload{
		Env:      env,
		Service:  vars.Service,
		Checksum: checksum,
	}
	err = githubClient.RepositoryDispatch(context.Background(), eventType, eventPayload)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("deploy event triggered")

	log.Print("done")
}
