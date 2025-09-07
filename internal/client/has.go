package client

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"

	"artship/internal/tools"
)

var ErrNotFound = errors.New("artifact not found")

func (c *Client) Has(ctx context.Context, imageRef, artifact string) error {
	if imageRef == "" {
		return fmt.Errorf("no image ref provided")
	}

	if len(artifact) == 0 {
		return fmt.Errorf("no artifact provided")
	}

	img, err := c.extractImage(ctx, imageRef)
	if err != nil {
		return err
	}
	defer img.Close()

	var found bool
	c.logger.Debug("Searching for artifact...")
	err = tools.WalkTar(img, func(r io.Reader, header *tar.Header) error {
		if tools.MatchName(header.Name, artifact) {
			c.logger.Debug("Found matching artifact: %s", header.Name)
			found = true
			return tools.ErrStopWalk
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk image: %w", err)
	}

	if !found {
		return ErrNotFound
	}

	return nil
}
