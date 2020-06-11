export ENV=dev
export DEPLOYMENT_BUCKET=serverless-monorepo-101-deployments
	Service     string `required:"true"`
	S3Bucket    string `required:"true"`
	GithubRepo  string `required:"true"`
	GithubToken string `required:"true"`