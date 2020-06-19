package internal

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	backendBuildEventTypePrefix = "backend-build"
)

func BackendBuildEventType(service string) string {
	return fmt.Sprintf("%s %s", backendBuildEventTypePrefix, service)
}

type BackendBuildEventPayload struct {
	CommitSHA string `json:"commitSHA"`
	Service   string `json:"service" envconfig:"SERVICE" required:"true"`
}

func LoadBackendBuildEventPayloadFromEnv() (*BackendBuildEventPayload, error) {
	eventPayload := BackendBuildEventPayload{}
	err := envconfig.Process("", &eventPayload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load event payload")
	}
	return &eventPayload, nil
}

const (
	backendDeployEventTypePrefix = "backend-deploy"
)

func BackendDeployEventType(service string, env string) string {
	return fmt.Sprintf("%s %s @ %s", backendDeployEventTypePrefix, service, env)
}

type BackendDeployEventPayload struct {
	Env      string `json:"env" envconfig:"ENV" required:"true"`
	Service  string `json:"service" envconfig:"SERVICE" required:"true"`
	Checksum string `json:"checksum" envconfig:"CHECKSUM" required:"true"`
}

func LoadBackendDeployEventPayloadFromEnv() (*BackendDeployEventPayload, error) {
	eventPayload := BackendDeployEventPayload{}
	err := envconfig.Process("", &eventPayload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load event payload")
	}
	return &eventPayload, nil
}
