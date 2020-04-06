package repository

import (
	"github.com/google/go-github/github"
)

// Discoverer is the interface all repo discoverer should implement
type Discoverer interface {
	Discover() chan *github.Repository
}
