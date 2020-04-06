package repository

import (
	"github.com/google/go-github/github"
	"gopkg.in/src-d/go-git.v4"
)

type Repository struct {
	git *git.Repository
	*github.Repository
	LocalDirectory string
}
