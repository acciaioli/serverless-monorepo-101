package internal

import (
	"github.com/pkg/errors"

	"github.com/kelseyhightower/envconfig"
)

type GitHubEnv struct {
	// Always set to true.
	Ci string `envconfig:"CI" required:"true"`
	// The path to the GitHub home directory used to store user data. For example, /github/home.
	Home string `envconfig:"HOME" required:"true"`
	// The name of the workflow.
	GitHubWorkflow string `envconfig:"GITHUB_WORKFLOW" required:"true"`
	// A unique number for each run within a repository. This number does not change if you re-run the workflow run.
	GitHubRunID string `envconfig:"GITHUB_RUN_ID" required:"true"`
	// A unique number for each run of a particular workflow in a repository. This number begins at 1 for the workflow's first run, and increments with each new run. This number does not change if you re-run the workflow run.
	GitHubRunNumber string `envconfig:"GITHUB_RUN_NUMBER" required:"true"`
	// The unique identifier (id) of the action.
	GitHubAction string `envconfig:"GITHUB_ACTION" required:"true"`
	// Always set to true when GitHub Actions is running the workflow. You can use this variable to differentiate when tests are being run locally or by GitHub Actions.
	GitHubActions string `envconfig:"GITHUB_ACTIONS" required:"true"`
	// The name of the person or app that initiated the workflow. For example, octocat.
	GitHubActor string `envconfig:"GITHUB_ACTOR" required:"true"`
	// The owner and repository name. For example, octocat/Hello-World.
	GitHubRepository string `envconfig:"GITHUB_REPOSITORY" required:"true"`
	// The name of the webhook event that triggered the workflow.
	GitHubEventName string `envconfig:"GITHUB_EVENT_NAME" required:"true"`
	// The path of the file with the complete webhook event payload. For example, /github/workflow/event.json.
	GitHubEventPath string `envconfig:"GITHUB_EVENT_PATH" required:"true"`
	// The GitHub workspace directory path. The workspace directory contains a subdirectory with a copy of your repository if your workflow uses the actions/checkout action. If you don't use the actions/checkout action, the directory will be empty. For example, /home/runner/work/my-repo-name/my-repo-name.
	GitHubWorkspace string `envconfig:"GITHUB_WORKSPACE" required:"true"`
	// The commit SHA that triggered the workflow. For example, ffac537e6cbbf934b08745a378932722df287a53.
	GitHubSha string `envconfig:"GITHUB_SHA" required:"true"`
	// The branch or tag ref that triggered the workflow. For example, refs/heads/feature-branch-1. If neither a branch or tag is available for the event type, the variable will not exist.
	GitHubRef string `envconfig:"GITHUB_REF" required:"true"`
	// Only set for forked repositories. The branch of the head repository.
	GitHubHeadRef string `envconfig:"GITHUB_HEAD_REF" required:"false"`
	// Only set for forked repositories. The branch of the base repository.
	GitHubBaseRef string `envconfig:"GITHUB_BASE_REF" required:"false"`
}

func LoadGitHubEnv() (*GitHubEnv, error) {
	env := GitHubEnv{}
	err := envconfig.Process("", &env)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load github actions env variables")
	}
	return &env, nil
}
