package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"infra/internal"
)

type Variables struct {
	Env      string
	Service  string
	Checksum *string
	*internal.GitHubEnv
	*internal.GitHubSecrets
}

func loadVariables() (*Variables, error) {
	env := flag.String("env", "", "environment")
	service := flag.String("service", "", "service id")
	checksum := flag.String("checksum", "", "service checksum")
	flag.Parse()

	if *env == "" {
		return nil, errors.New("`--env` not provided")
	}

	if *service == "" {
		return nil, errors.New("`--service` not provided")
	}

	if *checksum == "" {
		checksum = nil
	}

	githubEnv, err := internal.LoadGitHubEnv()
	if err != nil {
		return nil, err
	}

	secrets, err := internal.LoadSecrets()
	if err != nil {
		return nil, err
	}

	return &Variables{Service: *service, Checksum: checksum, GitHubSecrets: secrets, GitHubEnv: githubEnv}, nil
}

func main() {
	vars, err := loadVariables()
	if err != nil {
		log.Fatal(err)
	}

	var checksum string
	if vars.Checksum != nil {
		checksum = *vars.Checksum
	} else {
		log.Print("getting last checksum")
		bu, err := internal.NewBuildUtils(vars.InfraRegion, vars.InfraBucket, vars.Service)
		if err != nil {
			log.Fatal(err)
		}

		checksum, err = bu.GetLastCodeChecksum()
		if err != nil {
			log.Fatal(err)
		}
		log.Print(fmt.Sprintf("last code checksum: %s", checksum))
	}

	log.Print("triggering deploy event")
	githubClient, err := internal.NewGitHubClient(vars.GitHubRepository, vars.PersonalAccessToken)
	if err != nil {
		log.Fatal(err)
	}
	eventType := internal.BackendDeployEventType(vars.Service, vars.Env)
	eventPayload := internal.BackendDeployEventPayload{
		Env:      vars.Env,
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
