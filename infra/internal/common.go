package internal

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/pkg/errors"
)

const (
	binariesDir         = ".bin"
	serverlessYML       = "serverless.yml"
	distZip             = "dist.zip"
	lastCodeCheckSumKey = "last-checksum"
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

func (bu *BuildUtils) lastCodeChecksumKey() string {
	return filepath.Join(bu.service, lastCodeCheckSumKey)
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

func (bu *BuildUtils) GetLastCodeChecksum() (string, error) {
	data, err := bu.download(bu.lastCodeChecksumKey())
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

func (bu *BuildUtils) SetLastCodeChecksum(checksum string) error {
	return bu.upload(bu.lastCodeChecksumKey(), []byte(checksum))
}

func (bu *BuildUtils) GenerateDistZip() ([]byte, error) {
	fPaths, err := filepath.Glob(bu.binariesPattern())
	if err != nil {
		return nil, err
	}
	if len(fPaths) < 1 {
		return nil, errors.New("no binaries files found")
	}
	fPaths = append(fPaths, filepath.Join(bu.service, serverlessYML))

	return zipFiles(fPaths, func(fPath string) (string, error) {
		return filepath.Rel(bu.service, fPath)
	})
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

func (bu *BuildUtils) Deploy(env string, distZipPath string) error {
	distPath := "dist"
	err := unzipFiles(distZipPath, distPath)
	if err != nil {
		return err
	}

	cmd := exec.Command("serverless", "deploy", "--stage", env)
	cmd.Dir = distPath
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	fmt.Println("----out----")
	fmt.Println(stdout.String())
	fmt.Println("--------")
	fmt.Println("----err----")
	fmt.Println(stdout.String())
	fmt.Println("--------")
	return nil
}
