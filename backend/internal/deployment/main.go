package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/go-github/v32/github"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	AWSRegion = "eu-west-1"
)

func validateArgs(args map[string]*string) error {
	for k, v := range args {
		if *v == "" {
			return fmt.Errorf("`%s` is required", k)
		}
	}
	return nil
}

func dirHashCheckSum(root string) (string, error) {
	hash := sha1.New()

	if err := filepath.Walk(root, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fInfo.IsDir() {
			return nil
		}

		log.Print(fmt.Sprintf("adding '%s' to the hash", fPath))
		fReader, err := os.Open(fPath)
		if err != nil {
			return err
		}

		_, err = hash.Write([]byte(fPath))
		if err != nil {
			return err
		}

		content, err := ioutil.ReadAll(fReader)
		if err != nil {
			return err
		}

		_, err = hash.Write(content)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func latestdirHashCheckSum(bucket, service string) (string, error) {
	buf := aws.NewWriteAtBuffer([]byte{})
	session, err := session.NewSession(&aws.Config{Region: aws.String(AWSRegion)})
	if err != nil {
		return "", err
	}

	downloader := s3manager.NewDownloader(session)
	_, err = downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(service),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchKey:
				return "", nil // service was never deployed
			}
		}
		return "", err
	}

	return string(buf.Bytes()), nil
}

func scheduleServiceDeployment(service, githubOwnerRepo, githubToken string) error {
	githubOwnerRepoArr := strings.SplitN(githubOwnerRepo, "/", 2)
	githubOwner := githubOwnerRepoArr[0]
	githubRepo := githubOwnerRepoArr[1]

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	dispatchPayload := map[string]string{
		"service": service,
	}
	b, err := json.Marshal(dispatchPayload)
	if err != nil {
		return err
	}
	jsonRawMessage := json.RawMessage(b)
	// list all repositories for the authenticated user
	dispatchRequest := github.DispatchRequestOptions{
		EventType:     "deploy",
		ClientPayload: &jsonRawMessage,
	}
	_, _, err = client.Repositories.Dispatch(ctx, githubOwner, githubRepo, dispatchRequest)
	if err != nil {
		return err
	}
	return nil
}

type Environment struct {
	Service          string `envconfig:"SERVICE" required:"true"`
	DeploymentBucket string `envconfig:"DEPLOYMENT_BUCKET" required:"true"`
	GithubRepo       string `envconfig:"GITHUB_REPO" required:"true"`
	GithubToken      string `envconfig:"GITHUB_TOKEN" required:"true"`
}

func main() {
	env := Environment{}
	err := envconfig.Process("", &env)
	if err != nil {
		log.Fatal(err)
	}

	hashSum, err := dirHashCheckSum(env.Service)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(fmt.Sprintf("service hash checksum: <%s>", hashSum))

	latestHasSum, err := latestdirHashCheckSum(env.DeploymentBucket, env.Service)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(fmt.Sprintf("latest service hash checksum: <%s>", latestHasSum))
	err = scheduleServiceDeployment(env.Service, env.GithubRepo, env.GithubToken)
	if err != nil {
		log.Fatal(err)
	}
}
