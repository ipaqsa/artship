package client

import (
	"context"
	"fmt"
)

// MirrorOptions contains options for mirroring an image
type MirrorOptions struct {
	SourceUsername string
	SourcePassword string
	SourceToken    string
	SourceAuth     string
	SourceInsecure bool // Allow insecure connections to source registry
	DestUsername   string
	DestPassword   string
	DestToken      string
	DestAuth       string
	DestInsecure   bool // Allow insecure connections to destination registry
}

// MirrorResult contains information about the mirroring operation
type MirrorResult struct {
	SourceImage string
	DestImage   string
	Digest      string
	Size        int64
	Success     bool
}

// Mirror copies an image from source to destination
func (c *Client) Mirror(ctx context.Context, sourceRef, destRef string, opts *MirrorOptions) (*MirrorResult, error) {
	if sourceRef == "" {
		return nil, fmt.Errorf("source image reference is required")
	}
	if destRef == "" {
		return nil, fmt.Errorf("destination image reference is required")
	}

	c.logger.Info("Fetching source image: %s", sourceRef)

	// Convert source credentials to ImageAuthOptions
	sourceAuthOpts := &ImageAuthOptions{
		Username: opts.SourceUsername,
		Password: opts.SourcePassword,
		Token:    opts.SourceToken,
		Auth:     opts.SourceAuth,
		Insecure: opts.SourceInsecure,
	}

	// Fetch the image from source
	img, err := c.fetchImageWithOptions(ctx, sourceRef, sourceAuthOpts)
	if err != nil {
		return nil, fmt.Errorf("fetch source image: %w", err)
	}

	// Get image digest and size for reporting
	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("get image digest: %w", err)
	}

	size, err := img.Size()
	if err != nil {
		c.logger.Warn("Could not determine image size: %v", err)
		size = 0
	}

	c.logger.Info("Pushing image to destination: %s", destRef)

	// Convert destination credentials to ImageAuthOptions
	destAuthOpts := &ImageAuthOptions{
		Username: opts.DestUsername,
		Password: opts.DestPassword,
		Token:    opts.DestToken,
		Auth:     opts.DestAuth,
		Insecure: opts.DestInsecure,
	}

	// Write image to destination
	if err := c.writeImageWithOptions(ctx, destRef, img, destAuthOpts); err != nil {
		return nil, fmt.Errorf("write image to destination: %w", err)
	}

	c.logger.Info("Successfully mirrored image")
	c.logger.Debug("Digest: %s", digest.String())

	return &MirrorResult{
		SourceImage: sourceRef,
		DestImage:   destRef,
		Digest:      digest.String(),
		Size:        size,
		Success:     true,
	}, nil
}
