package client

import (
	"context"
	"errors"
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

func (c *Client) extract(ctx context.Context, imageRef, layer string) (io.ReadCloser, error) {
	if len(layer) == 0 {
		return c.extractImage(ctx, imageRef)
	}

	return c.extractLayer(ctx, imageRef, layer)
}

func (c *Client) extractImage(ctx context.Context, imageRef string) (io.ReadCloser, error) {
	img, err := c.image(ctx, imageRef)
	if err != nil {
		return nil, err
	}

	return mutate.Extract(img), nil
}

func (c *Client) extractLayer(ctx context.Context, imageRef, layerDigest string) (io.ReadCloser, error) {
	img, err := c.image(ctx, imageRef)
	if err != nil {
		return nil, err
	}

	digest, err := crv1.NewHash(layerDigest)
	if err != nil {
		return nil, fmt.Errorf("parse layer digest: %w", err)
	}

	// Get layers
	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("get image layers: %w", err)
	}

	// Find the layer with matching digest
	for _, layer := range layers {
		layerHash, err := layer.Digest()
		if err != nil {
			continue
		}

		if layerHash == digest {
			c.logger.Debug("Found matching layer: %s", layerDigest)
			return layer.Uncompressed()
		}
	}

	return nil, errors.New("layer not found")
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

// ImageAuthOptions contains authentication and connection options for registry operations
type ImageAuthOptions struct {
	Username string
	Password string
	Token    string
	Auth     string
	Insecure bool
}

// fetchImageWithOptions fetches an image with custom authentication options
func (c *Client) fetchImageWithOptions(ctx context.Context, imageRef string, opts *ImageAuthOptions) (crv1.Image, error) {
	// Build name options (for insecure registry support)
	var nameOpts []name.Option
	if opts != nil && opts.Insecure {
		nameOpts = append(nameOpts, name.Insecure)
	} else {
		nameOpts = c.nameOptions
	}

	ref, err := name.ParseReference(imageRef, nameOpts...)
	if err != nil {
		return nil, fmt.Errorf("parse image reference '%s': %w", imageRef, err)
	}

	// Determine which credentials to use
	var remoteOpts []remote.Option
	if opts != nil && (opts.Username != "" || opts.Token != "" || opts.Auth != "") {
		// Use custom credentials
		remoteOpts = setupRemoteOptions(opts.Username, opts.Password, opts.Auth, opts.Token)
	} else {
		// Use default client credentials
		remoteOpts = c.remoteOptions
	}

	c.logger.Debug("Fetching image from registry...")
	img, err := remote.Image(ref, remoteOpts...)
	if err != nil {
		return nil, fmt.Errorf("fetch image '%s': %w", imageRef, err)
	}

	return img, nil
}

// writeImageWithOptions writes an image to a registry with custom authentication options
func (c *Client) writeImageWithOptions(ctx context.Context, imageRef string, img crv1.Image, opts *ImageAuthOptions) error {
	// Build name options (for insecure registry support)
	var nameOpts []name.Option
	if opts != nil && opts.Insecure {
		nameOpts = append(nameOpts, name.Insecure)
	} else {
		nameOpts = c.nameOptions
	}

	ref, err := name.ParseReference(imageRef, nameOpts...)
	if err != nil {
		return fmt.Errorf("parse image reference '%s': %w", imageRef, err)
	}

	// Determine which credentials to use
	var remoteOpts []remote.Option
	if opts != nil && (opts.Username != "" || opts.Token != "" || opts.Auth != "") {
		// Use custom credentials
		remoteOpts = setupRemoteOptions(opts.Username, opts.Password, opts.Auth, opts.Token)
	} else {
		// Use default client credentials
		remoteOpts = c.remoteOptions
	}

	c.logger.Debug("Uploading image to registry...")
	if err := remote.Write(ref, img, remoteOpts...); err != nil {
		return fmt.Errorf("write image '%s': %w", imageRef, err)
	}

	return nil
}
