package client

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	crv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/ipaqsa/artship/internal/logs"
)

const (
	userAgent = "artship"
)

type Options struct {
	Username string
	Password string
	Token    string
	Auth     string
	Insecure bool
	Logger   *logs.Logger
}

type Client struct {
	nameOptions   []name.Option
	remoteOptions []remote.Option
	logger        *logs.Logger
}

func New(opts *Options) *Client {
	var nameOpts []name.Option
	if opts.Insecure {
		nameOpts = append(nameOpts, name.Insecure)
	}

	return &Client{
		nameOptions:   nameOpts,
		remoteOptions: setupRemoteOptions(opts.Username, opts.Password, opts.Auth, opts.Token),
		logger:        opts.Logger,
	}
}

// setupRemoteOptions sets up authentication options for remote registry access
func setupRemoteOptions(username, password, auth, token string) []remote.Option {
	var remoteOpts []remote.Option
	if (username != "" && password != "") || token != "" || auth != "" {
		remoteOpts = append(remoteOpts, remote.WithAuth(authn.FromConfig(authn.AuthConfig{
			Username:      username,
			Password:      password,
			IdentityToken: token,
			Auth:          auth,
		})))
	} else {
		// Use default keychain (Docker config, etc.)
		remoteOpts = append(remoteOpts, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	}

	remoteOpts = append(remoteOpts, remote.WithUserAgent(userAgent))

	return remoteOpts
}

func (c *Client) extractImage(ctx context.Context, imageRef string) (io.ReadCloser, error) {
	img, err := c.image(ctx, imageRef)
	if err != nil {
		return nil, err
	}

	return mutate.Extract(img), nil
}

func (c *Client) image(_ context.Context, imageRef string) (crv1.Image, error) {
	startTime := time.Now()

	ref, err := name.ParseReference(imageRef, c.nameOptions...)
	if err != nil {
		return nil, fmt.Errorf("parse the image reference '%s': %w", imageRef, err)
	}

	c.logger.Debug("Pulling the image...")
	img, err := remote.Image(ref, c.remoteOptions...)
	if err != nil {
		return nil, fmt.Errorf("fetch the image '%s': %w", imageRef, err)
	}

	defer c.logger.Debug("Successfully pulled image in %s", time.Since(startTime))

	return img, nil
}
