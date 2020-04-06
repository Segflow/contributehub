package repository

import (
	"fmt"

	"github.com/google/go-github/github"
)

type Filter struct {
	IncludeFork bool
	Languages   map[string]bool
	Ignore      map[string]bool
}

func (f *Filter) Check(repo *github.Repository) bool {
	name := fmt.Sprintf("%s/%s", repo.GetOwner().GetLogin(), repo.GetName())
	if f.Ignore[name] {
		return true
	}

	if !f.IncludeFork && repo.GetFork() {
		return false
	}

	if len(f.Languages) != 0 && !f.Languages[repo.GetLanguage()] {
		return false
	}

	return true
}

func (f *Filter) FilterChan(in <-chan *github.Repository) chan *github.Repository {
	out := make(chan *github.Repository)
	go func() {
		for repo := range in {
			if f.Check(repo) {
				out <- repo
			}
		}
		close(out)
	}()
	return out
}
