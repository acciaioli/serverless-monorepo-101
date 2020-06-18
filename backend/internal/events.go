package internal

type BackendDeployEvent struct {
	Env     string `envconfig:"ENV" required:"true"`
	Service string `envconfig:"SERVICE" required:"true"`
}
