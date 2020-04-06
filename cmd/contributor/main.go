package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/segflow/contribuehub/pkg/repository"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	cloneDir = "/tmp/contributehub"
)

var (
	githubToken = os.Getenv("GITHUB_TOKEN")
	// IgnoreRepos contains list of ignored repositories.
	IgnoreRepos = map[string]bool{
		"kubernetes/kubernetes": true,
	}
)

func createGitHubClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)

	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func startRepositoriesDiscoverer() chan *github.Repository {
	client := createGitHubClient()
	eventDiscoverer := repository.NewEventDiscoverer(client)

	ctx := context.Background()
	return eventDiscoverer.Discover(ctx)

}

func startRepositoriesFilterer(in chan *github.Repository) chan *github.Repository {
	filter := &repository.Filter{
		Languages: map[string]bool{
			"Go": true,
		},
		Ignore: IgnoreRepos,
	}

	return filter.FilterChan(in)
}

func startRepositoriesCloner(in chan *github.Repository) chan *repository.Repository {
	cloner := &repository.Cloner{
		Depth:    1,
		CloneDir: cloneDir,
	}

	ch := make(chan *repository.Repository)
	go func() {
		for repo := range in {
			gitRepo, err := cloner.Clone(repo)
			if err != nil {
				logrus.Warnf("Error cloning repository %q: %s", repo.GetURL(), err)
				continue
			}

			fmt.Printf("%s/%s cloned\n", repo.GetOwner().GetLogin(), repo.GetName())
			ch <- gitRepo
		}
	}()

	return ch
}

type processResult struct {
	*repository.Repository
	changeCount int
}

func startRepoProcessor(in chan *repository.Repository) chan *processResult {
	ch := make(chan *processResult)
	go func() {
		for repo := range in {
			count, err := repoProcessChanDirection(repo)
			if err != nil {
				logrus.Warnf("Error checking repository %q: %s", repo.GetURL(), err)
				continue
			}

			ch <- &processResult{
				Repository:  repo,
				changeCount: count,
			}
		}
	}()

	return ch
}

func main() {
	allRepos := startRepositoriesDiscoverer()
	repos := startRepositoriesFilterer(allRepos)
	clonedRepos := startRepositoriesCloner(repos)
	processedRepos := startRepoProcessor(clonedRepos)

	for repo := range processedRepos {
		if repo.changeCount == 0 {
			continue
		}
		fmt.Printf("Repo %s processed. %d changes.\n", repo.LocalDirectory, repo.changeCount)
	}
}
