package repository

import (
	"fmt"
	"path"

	"github.com/google/go-github/github"
	"gopkg.in/src-d/go-git.v4"
)

type Cloner struct {
	CloneDir string
	Depth    int
}

func (r *Cloner) Clone(repo *github.Repository) (*Repository, error) {
	opts := &git.CloneOptions{
		URL:      repo.GetCloneURL(),
		Depth:    r.Depth,
		Progress: nil,
	}

	dir := path.Join(r.CloneDir, repo.GetOwner().GetLogin(), repo.GetName())
	gitRepo, err := git.PlainClone(dir, false, opts)
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		return nil, err
	}

	if err == git.ErrRepositoryAlreadyExists {
		if gitRepo, err := git.PlainOpen(dir); err == nil {
			gitRepo.Fetch(&git.FetchOptions{Depth: 1})
		} else {
			return nil, fmt.Errorf("cannot fetch repository %s/%s: %v", repo.GetOwner().GetLogin(), repo.GetName(), err)
		}
	}

	return &Repository{
		git:            gitRepo,
		Repository:     repo,
		LocalDirectory: dir,
	}, nil
}
