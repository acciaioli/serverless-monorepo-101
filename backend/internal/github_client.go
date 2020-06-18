package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	owner  string
	repo   string
	client github.Client
}

func NewGitHubClient(githubOwnerRepo, accessToken string) (*GitHubClient, error) {
	ownerRepoArr := strings.Split(githubOwnerRepo, "/")
	if len(ownerRepoArr) != 2 {
		return nil, errors.New(fmt.Sprintf("invalid githubOwnerRepo: %s, couldn't be split", githubOwnerRepo))
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(context.Background(), ts)

	return &GitHubClient{
		owner:  ownerRepoArr[0],
		repo:   ownerRepoArr[1],
		client: *github.NewClient(tc),
	}, nil
}

func (c *GitHubClient) RepositoryDispatch(ctx context.Context, eventType string, eventPayload interface{}) error {
	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal event payload")
	}
	jsonRawMessage := json.RawMessage(payloadBytes)

	req := github.DispatchRequestOptions{
		EventType:     eventType,
		ClientPayload: &jsonRawMessage,
	}

	_, _, err = c.client.Repositories.Dispatch(ctx, c.owner, c.repo, req)
	if err != nil {
		return errors.Wrap(err, "repository dispatch req/resp error")
	}

	return nil
}
