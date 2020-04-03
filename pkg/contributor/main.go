package main

import (
	"context"
	"fmt"
	"github/segflow/contributehub/discovery"

	"github.com/sirupsen/logrus"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	githubToken = "feeffc0f3ffbf771c60ae7e9c5d2416afd5b3dc6"
)

var (
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

func StartRepositoriesDiscoverer() chan *github.Repository {
	client := createGitHubClient()
	eventDiscoverer := discovery.NewEventRepoDiscoverer(client)

	ctx := context.Background()
	return eventDiscoverer.Discover(ctx)

}

func StartRepositoriesFilterer(in chan *github.Repository) chan *github.Repository {
	filter := &RepositoriesFilter{
		Languages: map[string]bool{
			"Go": true,
		},
		Ignore: IgnoreRepos,
	}

	return filter.FilterChan(in)
}

func StartRepositoriesCloner(in chan *github.Repository) chan *Repository {
	cloner := &RepositoryCloner{
		Depth:    1,
		CloneDir: "/tmp/contributehub",
	}

	ch := make(chan *Repository)
	go func() {
		for repo := range in {
			gitRepo, err := cloner.Clone(repo)
			if err != nil {
				logrus.Warnf("Error cloning repository %q: %s", repo.GetURL(), err)
				continue
			}

			ch <- gitRepo
		}
	}()

	return ch
}

func main() {
	allRepos := StartRepositoriesDiscoverer()
	repos := StartRepositoriesFilterer(allRepos)
	clonedRepos := StartRepositoriesCloner(repos)

	for repo := range clonedRepos {
		fmt.Printf("%s/%s cloned\n", repo.GetOwner().GetLogin(), repo.GetName())
	}
}
