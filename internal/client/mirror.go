package client

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// MirrorOptions contains options for mirroring an image
type MirrorOptions struct {
	SourceUsername string
	SourcePassword string
	SourceToken    string
	SourceAuth     string
	DestUsername   string
	DestPassword   string
	DestToken      string
	DestAuth       string
	Insecure       bool
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

	// Parse source reference
	srcRef, err := name.ParseReference(sourceRef, c.nameOptions...)
	if err != nil {
		return nil, fmt.Errorf("parse source reference '%s': %w", sourceRef, err)
	}

	// Parse destination reference
	dstRef, err := name.ParseReference(destRef, c.nameOptions...)
	if err != nil {
		return nil, fmt.Errorf("parse destination reference '%s': %w", destRef, err)
	}

	// Setup remote options for source (may have different credentials)
	var srcRemoteOpts []remote.Option
	if opts != nil && (opts.SourceUsername != "" || opts.SourceToken != "" || opts.SourceAuth != "") {
		srcRemoteOpts = setupRemoteOptions(opts.SourceUsername, opts.SourcePassword, opts.SourceAuth, opts.SourceToken)
	} else {
		srcRemoteOpts = c.remoteOptions
	}

	// Fetch the image from source
	c.logger.Debug("Downloading image from source registry...")
	img, err := remote.Image(srcRef, srcRemoteOpts...)
	if err != nil {
		return nil, fmt.Errorf("fetch source image '%s': %w", sourceRef, err)
	}

	// Get image digest and size for reporting
	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("get image digest: %w", err)
	}

	size, err := img.Size()
	if err != nil {
		c.logger.Debug("Warning: could not determine image size")
		size = 0
	}

	c.logger.Info("Pushing image to destination: %s", destRef)

	// Setup remote options for destination (may have different credentials)
	var dstRemoteOpts []remote.Option
	if opts != nil && (opts.DestUsername != "" || opts.DestToken != "" || opts.DestAuth != "") {
		dstRemoteOpts = setupRemoteOptions(opts.DestUsername, opts.DestPassword, opts.DestAuth, opts.DestToken)
	} else {
		dstRemoteOpts = c.remoteOptions
	}

	// Write the image to destination
	c.logger.Debug("Uploading image to destination registry...")
	if err := remote.Write(dstRef, img, dstRemoteOpts...); err != nil {
		return nil, fmt.Errorf("write image to destination '%s': %w", destRef, err)
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
