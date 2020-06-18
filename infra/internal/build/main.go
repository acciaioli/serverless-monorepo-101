package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/pkg/errors"

	"infra/internal"
)

const binariesDir = ".bin"

type Variables struct {
	Service string
	*internal.GitHubEnv
	*internal.GitHubSecrets
}

func loadVariables() (*Variables, error) {
	service := flag.String("service", "", "service id")
	flag.Parse()

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

	return &Variables{Service: *service, GitHubSecrets: secrets, GitHubEnv: githubEnv}, nil
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

	liveChecksum, err := bu.GetLiveCodeChecksum()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(fmt.Sprintf("live code checksum: %s", liveChecksum))

	if checksum == liveChecksum {
		log.Print("service was not updated")
		return
	}

	zData, err := bu.GenerateDistZip()
	if err != nil {
		log.Fatal(err)
	}

	err = bu.UploadDistZip(checksum, zData)
	if err != nil {
		log.Fatal(err)
	}

	githubClient, err := internal.NewGitHubClient(vars.GitHubRepository, vars.PersonalAccessToken)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("service was updated - triggering deployment...")
	env := "dev" // deploy on dev automatically
	payload := internal.BackendDeployEvent{
		Env:      env,
		Service:  vars.Service,
		Checksum: checksum,
	}
	err = githubClient.RepositoryDispatch(context.Background(), fmt.Sprintf("[deploy] %s @ %s", vars.Service, env), payload)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("service deployment triggered!")
}
