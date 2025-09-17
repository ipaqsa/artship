package client

import (
	"archive/tar"
	"context"
	"fmt"
	"io"

	"github.com/ipaqsa/artship/internal/tools"
)

func (c *Client) Cat(ctx context.Context, imageRef, artifact string) ([]byte, error) {
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

	var content []byte
	c.logger.Debug("Searching for artifact...")
	err = tools.WalkTar(img, func(r io.Reader, header *tar.Header) error {
		if tools.MatchName(header.Name, artifact) && header.Typeflag == tar.TypeReg {
			c.logger.Debug("Found artifact: %s (size: %d bytes)", header.Name, header.Size)
			content, err = io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("read artifact content: %w", err)
			}

			return tools.ErrStopWalk
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk the image: %w", err)
	}

	if content == nil {
		return nil, fmt.Errorf("artifact empty or not found")
	}

	c.logger.Debug("Successfully read artifact content (%d bytes)", len(content))
	return content, nil
}
