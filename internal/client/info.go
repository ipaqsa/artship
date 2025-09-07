package client

import (
	"archive/tar"
	"context"
	"fmt"
	"io"

	"artship/internal/tools"
)

// Artifact contains information about an artifact
type Artifact struct {
	Path string
	Size int64
	Type string
	Mode string
}

// String returns a formatted string representation of the artifact
func (a Artifact) String(header bool) string {
	sizeStr := tools.FormatSize(a.Size)
	if a.Type == "dir" || a.Type == "symlink" || a.Type == "hardlink" {
		sizeStr = "-"
	}

	var result string
	if header {
		result = fmt.Sprintf("%-8s %-10s %-8s %s\n", "TYPE", "SIZE", "MODE", "PATH")
		result += "-------- ---------- -------- --------\n"
	}

	result += fmt.Sprintf("%-8s %-10s %-8s %s\n", a.Type, sizeStr, a.Mode, a.Path)

	return result
}

// GetArtifact retrieves detailed information about a specific artifact
func (c *Client) GetArtifact(ctx context.Context, imageRef, artifact string) (*Artifact, error) {
	if imageRef == "" {
		return nil, fmt.Errorf("no image ref provided")
	}

	if artifact == "" {
		return nil, fmt.Errorf("no artifact provided")
	}

	img, err := c.extractImage(ctx, imageRef)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	var info *Artifact
	c.logger.Debug("Searching for artifact...")
	err = tools.WalkTar(img, func(_ io.Reader, header *tar.Header) error {
		if tools.MatchName(header.Name, artifact) {
			info = &Artifact{
				Path: header.Name,
				Size: header.Size,
				Type: tools.GetArtifactType(header.Typeflag),
				Mode: fmt.Sprintf("%04o", header.Mode),
			}

			return tools.ErrStopWalk
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk the image: %w", err)
	}

	if info == nil {
		return nil, fmt.Errorf("artifact '%s' not found", artifact)
	}

	c.logger.Debug("Successfully retrieved artifact info")
	return info, nil
}
