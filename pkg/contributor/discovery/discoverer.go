package discovery

import (
	"github.com/google/go-github/github"
)

type RepoDiscoverer interface {
	Discover() chan *github.Repository
}
