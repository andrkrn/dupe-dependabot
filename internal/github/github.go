package github

import (
	"context"
	"net/http"
	"os"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

type Github struct {
	gc *github.Client
}

func NewGithub(token string) *Github {
	client := github.NewClient(authenticate(token))

	return &Github{gc: client}
}

func authenticate(token string) *http.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.TODO(), ts)

	return tc
}

var dependabot = map[string]bool{
	"dependabot[bot]":         true,
	"dependabot-preview[bot]": true,
}

func (g *Github) DependabotPullRequests() []*github.PullRequest {
	page := 1
	result := make([]*github.PullRequest, 0)

	for {
		prs, _, err := g.gc.PullRequests.List(context.TODO(), os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO"), &github.PullRequestListOptions{
			ListOptions: github.ListOptions{Page: page, PerPage: 100},
			Direction:   "desc",
			State:       "open",
		})

		if err != nil {
			panic(err)
		}

		if len(prs) == 0 {
			break
		}

		for _, pr := range prs {
			if dependabot[pr.User.GetLogin()] {
				result = append(result, pr)
			}
		}

		page++
	}

	return result
}

func (g *Github) ClosePullRequest(number int) error {
	_, _, err := g.gc.PullRequests.Edit(context.TODO(), os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO"), number, &github.PullRequest{
		State: github.String("closed"),
	})

	if err != nil {
		return err
	}

	return nil
}
