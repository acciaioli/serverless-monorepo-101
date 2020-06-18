package internal

type BackendDeployEvent struct {
	Env      string `json:"env" envconfig:"ENV" required:"true"`
	Service  string `json:"service" envconfig:"SERVICE" required:"true"`
	Checksum string `json:"checksum" envconfig:"CHECKSUM" required:"true"`
}
