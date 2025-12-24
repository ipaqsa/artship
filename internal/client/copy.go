package client

import (
	"archive/tar"
	"context"
	"fmt"
	"io"

	"github.com/ipaqsa/artship/internal/tools"
)

func (c *Client) Copy(ctx context.Context, imageRef string, artifacts []string, output string) error {
	if imageRef == "" {
		return fmt.Errorf("no image ref provided")
	}

	if len(artifacts) == 0 {
		return fmt.Errorf("no artifacts provided")
	}

	if output == "" {
		return fmt.Errorf("no output provided")
	}

	img, err := c.extractImage(ctx, imageRef)
	if err != nil {
		return err
	}
	defer img.Close()

	var found int
	c.logger.Debug("Searching for artifacts...")
	err = tools.WalkTar(img, func(r io.Reader, header *tar.Header) error {
		for _, artifact := range artifacts {
			if tools.MatchName(header.Name, artifact) {
				c.logger.Debug("Found artifact: %s (matches %s)", header.Name, artifact)
				found++
				if err = tools.CopyArtifact(r, header, output); err != nil {
					return fmt.Errorf("copy the artifact '%s': %w", artifact, err)
				}
				c.logger.Info(tools.Green("✓")+" Copied: %s", tools.Blue(header.Name))
			}
		}

		if found == len(artifacts) {
			c.logger.Debug("All artifacts found, stopping search")
			return tools.ErrStopWalk
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk image: %w", err)
	}

	if found < len(artifacts) {
		c.logger.Info(tools.Yellow("⚠")+" Warning: Only found %d of %d requested artifacts", found, len(artifacts))
	} else {
		c.logger.Info(tools.BoldGreen(fmt.Sprintf("✓ Successfully copied %d artifacts", found)))
	}

	return nil
}
