package client

import (
	"archive/tar"
	"context"
	"fmt"
	"io"

	"github.com/ipaqsa/artship/internal/tools"
)

// ArtifactList represents a collection of artifacts
type ArtifactList []Artifact

// String returns a formatted string representation of the artifact list
func (l ArtifactList) String(detailed bool) string {
	if len(l) == 0 {
		return "No artifacts found"
	}

	if detailed {
		result := fmt.Sprintf("%-8s %-10s %-8s %s\n", "TYPE", "SIZE", "MODE", "PATH")
		result += "-------- ---------- -------- --------\n"
		for _, artifact := range l {
			result += artifact.String(false)
		}

		return result
	}

	var res string
	for _, artifact := range l {
		res += fmt.Sprintf("%s\n", artifact.Path)
	}

	return res
}

// List lists all available artifacts in image
func (c *Client) List(ctx context.Context, imageRef, filter string) (ArtifactList, error) {
	if imageRef == "" {
		return nil, fmt.Errorf("no image ref provided")
	}

	c.logger.Debug("Walking the image...")

	img, err := c.extractImage(ctx, imageRef)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	var artifacts []Artifact
	c.logger.Debug("Scanning image artifacts...")
	err = tools.WalkTar(img, func(_ io.Reader, header *tar.Header) error {
		artType := tools.GetArtifactType(header.Typeflag)

		// Apply type filter
		if filter != "" && filter != "all" && filter != artType {
			return nil
		}

		artifacts = append(artifacts, Artifact{
			Path: header.Name,
			Size: header.Size,
			Type: artType,
			Mode: fmt.Sprintf("%04o", header.Mode),
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk the image: %w", err)
	}

	c.logger.Debug("Found %d artifacts", len(artifacts))
	return artifacts, nil
}
