package client

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	crv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// Pack creates an OCI image from local files or directories using modern streaming APIs
func (c *Client) Pack(ctx context.Context, imageRef, sourcePath string) error {
	startTime := time.Now()

	// Validate inputs
	if err := c.validatePackInputs(imageRef, sourcePath); err != nil {
		return err
	}

	c.logger.Debug("Creating OCI image from source: %s", sourcePath)

	// Create streaming layer from source using modern API
	layer, err := c.createStreamingLayer(sourcePath)
	if err != nil {
		return fmt.Errorf("create streaming layer: %w", err)
	}
	defer func() {
		if closer, ok := layer.(io.Closer); ok {
			if closeErr := closer.Close(); closeErr != nil {
				c.logger.Debug("Failed to close layer: %v", closeErr)
			}
		}
	}()

	// Create image with proper platform metadata
	img, err := c.createImageWithMetadata(layer, sourcePath)
	if err != nil {
		return fmt.Errorf("create image with metadata: %w", err)
	}

	c.logger.Debug("Successfully created image from %s in %s", sourcePath, time.Since(startTime))

	c.logger.Info("Pushing layer to the registry")
	if err = c.pushImage(ctx, imageRef, img); err != nil {
		return fmt.Errorf("push image: %w", err)
	}

	return nil
}

// validatePackInputs validates the input parameters
func (c *Client) validatePackInputs(imageRef, sourcePath string) error {
	if imageRef == "" {
		return errors.New("no image reference provided")
	}

	if sourcePath == "" {
		return errors.New("source path is required")
	}

	info, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source path '%s' does not exist", sourcePath)
		}

		return fmt.Errorf("stat the source path '%s': %w", sourcePath, err)
	}

	// Additional validation for special cases
	if info.Mode()&os.ModeSocket != 0 {
		return fmt.Errorf("source path '%s' is a socket, which cannot be packed", sourcePath)
	}

	return nil
}

// createStreamingLayer creates a layer using the modern streaming API
func (c *Client) createStreamingLayer(sourcePath string) (crv1.Layer, error) {
	// Create pipe for streaming tar data
	r, w := io.Pipe()

	// Start goroutine to write tar data
	go func() {
		defer func() {
			if err := w.Close(); err != nil {
				c.logger.Debug("Error closing pipe writer: %v", err)
			}
		}()

		tw := tar.NewWriter(w)
		defer func() {
			if err := tw.Close(); err != nil {
				c.logger.Debug("Error closing tar writer: %v", err)
			}
		}()

		if err := c.addToTarStream(tw, sourcePath, ""); err != nil {
			c.logger.Debug("Error creating tar stream: %v", err)
			if closeErr := w.CloseWithError(err); closeErr != nil {
				c.logger.Debug("Error closing pipe with error: %v", closeErr)
			}
		}
	}()

	// Use modern streaming API with optimized compression
	layer := stream.NewLayer(r,
		stream.WithCompressionLevel(gzip.BestSpeed),
		stream.WithMediaType(types.OCILayer),
	)

	return layer, nil
}

// addToTarStream adds files to tar stream with improved error handling
func (c *Client) addToTarStream(tw *tar.Writer, sourcePath, basePath string) error {
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk error for '%s': %w", path, walkErr)
		}

		// Skip problematic file types early
		mode := info.Mode()
		if mode&(os.ModeSocket|os.ModeDevice|os.ModeNamedPipe) != 0 {
			c.logger.Debug("Skipping special file: %s (mode: %s)", path, mode)
			return nil
		}

		// Create tar header with proper error handling
		header, err := c.createTarHeader(info, path, sourcePath, basePath)
		if err != nil {
			return err
		}

		// Handle different file types with proper cleanup
		return c.writeToTar(tw, header, path, info)
	})
}

// createTarHeader creates a tar header with proper path handling
func (c *Client) createTarHeader(info os.FileInfo, path, sourcePath, basePath string) (*tar.Header, error) {
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return nil, fmt.Errorf("create tar header for '%s': %w", path, err)
	}

	// Calculate relative path
	relPath, err := filepath.Rel(sourcePath, path)
	if err != nil {
		return nil, fmt.Errorf("calculate relative path for '%s': %w", path, err)
	}

	// Normalize path for tar (always use forward slashes)
	tarPath := filepath.ToSlash(relPath)
	if basePath != "" {
		tarPath = strings.TrimSuffix(basePath, "/") + "/" + tarPath
	}

	header.Name = tarPath

	// Handle symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(path)
		if err != nil {
			return nil, fmt.Errorf("read the symlink '%s': %w", path, err)
		}
		header.Linkname = link
	}

	return header, nil
}

// writeToTar writes the tar header and content with proper resource management
func (c *Client) writeToTar(tw *tar.Writer, header *tar.Header, path string, info os.FileInfo) error {
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("write tar header for '%s': %w", path, err)
	}

	mode := info.Mode()
	switch {
	case mode.IsRegular():
		return c.writeFileContent(tw, path, header.Name)
	case mode.IsDir():
		c.logger.Debug("Added directory: %s", header.Name)
		return nil
	case mode&os.ModeSymlink != 0:
		c.logger.Debug("Added symlink: %s -> %s", header.Name, header.Linkname)
		return nil
	default:
		c.logger.Debug("Added special file: %s", header.Name)
		return nil
	}
}

// writeFileContent writes file content to tar with proper resource management
func (c *Client) writeFileContent(tw *tar.Writer, filePath, tarPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file '%s': %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			c.logger.Debug("Error closing file '%s': %v", filePath, closeErr)
		}
	}()

	if _, err = io.Copy(tw, file); err != nil {
		return fmt.Errorf("write file content for '%s': %w", filePath, err)
	}

	c.logger.Debug("Added file: %s", tarPath)
	return nil
}

// createImageWithMetadata creates image with proper platform metadata and OCI compliance
func (c *Client) createImageWithMetadata(layer crv1.Layer, sourcePath string) (crv1.Image, error) {
	// Create comprehensive metadata first
	labels := c.createImageLabels(sourcePath)

	// Start with empty image and configure it first
	baseImg := empty.Image

	// Create the config with metadata
	config := crv1.Config{
		Labels: labels,
	}

	// Apply config to base image
	img, err := mutate.Config(baseImg, config)
	if err != nil {
		return nil, fmt.Errorf("apply initial config: %w", err)
	}

	// Set platform information
	configFile, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("get initial config: %w", err)
	}

	// Update platform info
	newConfigFile := *configFile
	newConfigFile.Architecture = runtime.GOARCH
	newConfigFile.OS = runtime.GOOS

	img, err = mutate.ConfigFile(img, &newConfigFile)
	if err != nil {
		return nil, fmt.Errorf("apply platform config: %w", err)
	}

	// Now append the layer
	img, err = mutate.AppendLayers(img, layer)
	if err != nil {
		return nil, fmt.Errorf("append layer to image: %w", err)
	}

	return img, nil
}

// createImageLabels creates comprehensive OCI-compliant labels
func (c *Client) createImageLabels(sourcePath string) map[string]string {
	now := time.Now().UTC()
	labels := map[string]string{
		"org.opencontainers.image.created":     now.Format(time.RFC3339),
		"org.opencontainers.image.source":      "artship",
		"org.opencontainers.image.title":       "Packed by artship",
		"org.opencontainers.image.description": fmt.Sprintf("OCI image created from %s", sourcePath),
		"org.opencontainers.image.vendor":      "artship",
		"org.opencontainers.image.version":     "latest",
		"artship.source.path":                  sourcePath,
		"artship.created.timestamp":            fmt.Sprintf("%d", now.Unix()),
	}

	// Add platform-specific labels
	labels["artship.platform.os"] = runtime.GOOS
	labels["artship.platform.arch"] = runtime.GOARCH

	return labels
}

// pushImage handles image pushing with proper error handling
func (c *Client) pushImage(ctx context.Context, imageRef string, img crv1.Image) error {
	pushStart := time.Now()
	c.logger.Debug("Pushing image to registry...")

	ref, err := name.ParseReference(imageRef, c.nameOptions...)
	if err != nil {
		return fmt.Errorf("parse image reference '%s': %w", imageRef, err)
	}

	// Add context to remote options for better cancellation support
	remoteOpts := append(c.remoteOptions, remote.WithContext(ctx))

	if err = remote.Write(ref, img, remoteOpts...); err != nil {
		return fmt.Errorf("push image to registry: %w", err)
	}

	c.logger.Info("Successfully pushed image to %s in %s", imageRef, time.Since(pushStart))
	return nil
}
