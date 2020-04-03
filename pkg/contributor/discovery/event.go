package discovery

import (
	"context"
	"errors"
	"time"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

const (
	eventPerPage   = 300
	discoverPeriod = 30 * time.Second
	goroutines     = 8
)

type EventRepoDiscoverer struct {
	client    *github.Client
	seenRepos map[int64]bool
	lastETAG  string
}

func NewEventRepoDiscoverer(c *github.Client) *EventRepoDiscoverer {
	return &EventRepoDiscoverer{
		client:    c,
		seenRepos: make(map[int64]bool),
	}
}

func (e *EventRepoDiscoverer) getNewEvents(ctx context.Context) ([]*github.Event, error) {
	var events []*github.Event
	opt := &github.ListOptions{PerPage: eventPerPage}

	for {
		evs, resp, err := e.client.Activity.ListEvents(ctx, opt)
		if err != nil {
			return nil, err
		}
		events = append(events, evs...)
		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return events, nil
}

func (e *EventRepoDiscoverer) Discover(ctx context.Context) chan *github.Repository {
	ch := make(chan *github.Repository)
	go func() {
		e.discoverLoop(ctx, ch)
	}()
	return ch
}

func (e *EventRepoDiscoverer) discoverLoop(ctx context.Context, ch chan<- *github.Repository) {
	for {
		logrus.Info("Discovering new repositories.")
		err := e.discover(ctx, ch)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			logrus.Infof("Stopping event based discoverer: %s", err)
			return
		}

		time.Sleep(discoverPeriod)
	}
}

func (e *EventRepoDiscoverer) discover(ctx context.Context, ch chan<- *github.Repository) error {
	events, err := e.getNewEvents(ctx)
	if err != nil {
		return err
	}

	eventsCh := make(chan *github.Event)
	for i := 0; i < goroutines; i++ {
		go func() {
			for event := range eventsCh {
				repo, _, err := e.client.Repositories.GetByID(ctx, event.GetRepo().GetID())
				if err != nil {
					// logrus.Warnf("Error getting repository %q: %s", event.GetRepo().GetName(), err)
					continue
				}

				select {
				case <-ctx.Done():
					return
				case ch <- repo:
				}
			}
		}()
	}

	for _, event := range events {

		if e.seenRepos[event.GetRepo().GetID()] {
			continue
		}
		e.seenRepos[event.GetRepo().GetID()] = true // mark repository as seen

		select {
		case eventsCh <- event:
		case <-ctx.Done():
			close(ch)
			return ctx.Err()
		}
	}

	return ctx.Err()
}
