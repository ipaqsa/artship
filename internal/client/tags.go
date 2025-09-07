package client

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Tags lists all available tags for a repository
func (c *Client) Tags(_ context.Context, repoName string) ([]string, error) {
	if repoName == "" {
		return nil, errors.New("no repository provided")
	}

	startTime := time.Now()

	c.logger.Debug("Parsing repository: %s", repoName)
	repo, err := name.NewRepository(repoName, c.nameOptions...)
	if err != nil {
		return nil, fmt.Errorf("parse the repository '%s': %w", repo, err)
	}

	c.logger.Debug("Listing tags...")
	tags, err := remote.List(repo, c.remoteOptions...)
	if err != nil {
		return nil, fmt.Errorf("list tags for the repository '%s': %w", repoName, err)
	}

	c.logger.Debug("Successfully listed %d tags in %s", len(tags), time.Since(startTime))

	// Sort tags alphabetically
	sort.Slice(tags, func(i, j int) bool {
		return tags[i] < tags[j]
	})

	return tags, nil
}
