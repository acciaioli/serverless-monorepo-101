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
	*internal.UserEnv
	*internal.Secrets
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

	userEnv, err := internal.LoadUserEnv()
	if err != nil {
		return nil, err
	}

	secrets, err := internal.LoadSecrets()
	if err != nil {
		return nil, err
	}

	return &Variables{Env: *env, Service: *service, Checksum: checksum, UserEnv: userEnv, Secrets: secrets}, nil
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
	githubClient, err := internal.NewGitHubClient(vars.Repository, vars.PersonalAccessToken)
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
