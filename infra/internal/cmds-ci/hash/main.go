package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/pkg/errors"

	"infra/internal"
)

type Variables struct {
	CommitSHA string
	Service   string
	*internal.GitHubEnv
	*internal.GitHubSecrets
}

func loadVariables() (*Variables, error) {
	commitSHA := flag.String("commit-sha", "", "commit sha")
	service := flag.String("service", "", "service id")
	flag.Parse()

	if *commitSHA == "" {
		return nil, errors.New("`--service` not provided")
	}

	if *service == "" {
		return nil, errors.New("`--service` not provided")
	}

	githubEnv, err := internal.LoadGitHubEnv()
	if err != nil {
		return nil, err
	}

	secrets, err := internal.LoadSecrets()
	if err != nil {
		return nil, err
	}

	return &Variables{CommitSHA: *commitSHA, Service: *service, GitHubSecrets: secrets, GitHubEnv: githubEnv}, nil
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

	log.Print("computing checksum")
	checksum, err := bu.ComputeCodeChecksum()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(fmt.Sprintf("code checksum: %s", checksum))

	log.Print("getting last checksum")
	lastChecksum, err := bu.GetLastCodeChecksum()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(fmt.Sprintf("last code checksum: %s", lastChecksum))

	if checksum != lastChecksum {
		log.Print("new checksum! triggering build event")
		githubClient, err := internal.NewGitHubClient(vars.GitHubRepository, vars.PersonalAccessToken)
		if err != nil {
			log.Fatal(err)
		}

		eventType := internal.BackendBuildEventType(vars.Service)
		eventPayload := internal.BackendBuildEventPayload{
			CommitSHA: vars.CommitSHA,
			Service:   vars.Service,
		}
		err = githubClient.RepositoryDispatch(context.Background(), eventType, eventPayload)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("build event triggered")
	} else {
		log.Print("nothing to do")
	}
	log.Print("done")
}
