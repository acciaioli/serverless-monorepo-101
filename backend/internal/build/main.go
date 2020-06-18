package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"backend/internal"
)

const binariesDir = ".bin"

type Variables struct {
	Service string
	*internal.Secrets
	*internal.GitHubEnv
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

	return &Variables{Service: *service, Secrets: secrets, GitHubEnv: githubEnv}, nil
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

		isBin, err := filepath.Match(filepath.Join(root, binariesDir, "*"), fPath)
		if err != nil {
			return err
		}
		if isBin {
			return nil
		}

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

func latestDirHashCheckSum(bucket, region, service string) (string, error) {
	buf := aws.NewWriteAtBuffer([]byte{})
	session, err := aws_session.NewSession(&aws.Config{Region: aws.String(region)})
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

func main() {
	vars, err := loadVariables()
	if err != nil {
		log.Fatal(err)
	}

	hashSum, err := dirHashCheckSum(vars.Service)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(fmt.Sprintf("service hash checksum: '%s'", hashSum))

	//latestHasSum, err := latestDirHashCheckSum(vars.DeploymentBucket, vars.DeploymentRegion, vars.Service)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//log.Print(fmt.Sprintf("latest service hash checksum: '%s'", latestHasSum))
	//
	//if hashSum == latestHasSum {
	//	log.Print("service was not updated")
	//	return
	//}
	//
	//githubClient, err := internal.NewGitHubClient(vars.GitHubRepository, vars.PersonalAccessToken)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Print("service was updated - triggering deployment...")
	//payload := map[string]string{
	//	"service": vars.Service,
	//}
	//err = githubClient.RepositoryDispatch(context.Background(), fmt.Sprintf("deploy %s", vars.Service), payload)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Print("service deployment triggered!")
}
