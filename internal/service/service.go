package service

import (
	"fmt"
	"regexp"

	"github.com/andrkrn/dupe-dependabot/internal/github"
	g "github.com/google/go-github/v39/github"
	"github.com/hashicorp/go-version"
)

type Github interface {
	DependabotPullRequests() []*g.PullRequest
	ClosePullRequest(number int) error
}

type DupeDependabot struct {
	github Github
}

func NewService(token string) *DupeDependabot {
	return &DupeDependabot{
		github: github.NewGithub(token),
	}
}

type Dependency struct {
	pr   *g.PullRequest
	from string
	to   string
}

func (s *DupeDependabot) Run() {
	prs := s.github.DependabotPullRequests()

	serviceDeps := map[string]map[string][]*Dependency{}
	r := regexp.MustCompile("Bump (.+) from (.+) to (.+) in /(.+)")

	for _, pr := range prs {
		match := r.FindStringSubmatch(pr.GetTitle())

		if len(match) == 0 {
			continue
		}

		lib := match[1]
		from := match[2]
		to := match[3]
		service := match[4]

		if serviceDeps[service] == nil {
			serviceDeps[service] = map[string][]*Dependency{}
		}
		if serviceDeps[service][lib] == nil {
			serviceDeps[service][lib] = make([]*Dependency, 0)
		}

		serviceDeps[service][lib] = append(serviceDeps[service][lib], &Dependency{pr: pr, from: from, to: to})
	}

	for service, libs := range serviceDeps {
		for lib, deps := range libs {
			selectedVersion, _ := version.NewVersion("0")
			var selectedPr *g.PullRequest

			for i, dep := range deps {
				version, _ := version.NewVersion(dep.to)

				if i == 0 {
					selectedVersion = version
					selectedPr = dep.pr

					continue
				}

				if selectedVersion.GreaterThan(version) {
					fmt.Println("Closing service: ", service, " lib: ", lib, " => ", dep.pr.GetTitle())
					s.github.ClosePullRequest(dep.pr.GetNumber())
				} else {
					fmt.Println("Closing service: ", service, " lib: ", lib, " => ", selectedPr.GetTitle())
					s.github.ClosePullRequest(selectedPr.GetNumber())

					selectedVersion = version
					selectedPr = dep.pr
				}
			}

			fmt.Println("Service: ", service, " lib: ", lib, " => ", selectedPr.GetTitle())
		}
	}
}
