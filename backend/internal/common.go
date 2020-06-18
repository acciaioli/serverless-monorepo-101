package internal

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type Secrets struct {
	// The bucket use to store deployment related state.
	DeploymentBucket string `envconfig:"DEPLOYMENT_AWS_S3_BUCKET" required:"true"`
	// The region for everything deployment related
	DeploymentRegion string `envconfig:"DEPLOYMENT_AWS_REGION" default:"eu-west-1"`
	// Github Personal Access Token
	// (https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line)
	PersonalAccessToken string `envconfig:"PERSONAL_ACCESS_TOKEN" required:"true"`
}

func LoadSecrets() (*Secrets, error) {
	env := Secrets{}
	err := envconfig.Process("", &env)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load secrets")
	}
	return &env, nil
}

const (
	binariesDir         = ".bin"
	serverlessYML       = "serverless.yml"
	distZip             = "dist.zip"
	liveCodeCheckSumKey = "live-code-checksum"
)

type BuildUtils struct {
	region string
	bucket string

	service string
}

func NewBuildUtils(region, bucket, service string) (*BuildUtils, error) {
	return &BuildUtils{region: region, bucket: bucket, service: service}, nil
}

func (bu *BuildUtils) binariesPattern() string {
	return filepath.Join(bu.service, binariesDir, "*")
}

func (bu *BuildUtils) liveCodeChecksumKey() string {
	return filepath.Join(bu.service, liveCodeCheckSumKey)
}

func (bu *BuildUtils) checksumDistZipKey(checksum string) string {
	return filepath.Join(bu.service, checksum, distZip)
}

func (bu *BuildUtils) download(key string) ([]byte, error) {
	session, err := aws_session.NewSession(&aws.Config{Region: aws.String(bu.region)})
	if err != nil {
		return nil, err
	}

	buf := aws.NewWriteAtBuffer([]byte{})
	downloader := s3manager.NewDownloader(session)
	_, err = downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(bu.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (bu *BuildUtils) upload(key string, data []byte) error {
	session, err := aws_session.NewSession(&aws.Config{Region: aws.String(bu.region)})
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(session)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bu.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	return err
}

func (bu *BuildUtils) ComputeCodeChecksum() (string, error) {
	hash := sha1.New()
	err := filepath.Walk(bu.service, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fInfo.IsDir() {
			return nil
		}

		isBin, err := filepath.Match(bu.binariesPattern(), fPath)
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
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (bu *BuildUtils) GetLiveCodeChecksum() (string, error) {
	data, err := bu.download(bu.liveCodeChecksumKey())
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchKey:
				return "", nil // service was never deployed
			}
		}
		return "", err
	}

	return string(data), nil
}

func (bu *BuildUtils) SetLiveCodeChecksum(checksum string) error {
	return bu.upload(bu.liveCodeChecksumKey(), []byte(checksum))
}

func (bu *BuildUtils) GenerateDistZip() ([]byte, error) {
	fPaths, err := filepath.Glob(bu.binariesPattern())
	if err != nil {
		return nil, err
	}
	fPaths = append(fPaths, filepath.Join(bu.service, serverlessYML))

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for _, fPath := range fPaths {
		if err := func() error {
			r, err := os.Open(fPath)
			if err != nil {
				return err
			}
			defer r.Close()

			zPath, err := filepath.Rel(bu.service, fPath)
			if err != nil {
				return err
			}
			w, err := zipWriter.Create(zPath)
			if err != nil {
				return err
			}

			_, err = io.Copy(w, r)
			if err != nil {
				return err
			}

			return nil
		}(); err != nil {
			return nil, err
		}
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (bu *BuildUtils) UploadDistZip(checksum string, zipData []byte) error {
	return bu.upload(bu.checksumDistZipKey(checksum), zipData)
}

func (bu *BuildUtils) DownloadDistZip(checksum string) (string, error) {
	data, err := bu.download(bu.checksumDistZipKey(checksum))
	if err != nil {
		return "", err
	}

	f, err := os.Create(distZip)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return "", err
	}

	return distZip, nil
}
