package client

import (
	"context"
	"fmt"
	"os"

	"artship/internal/tools"
)

func (c *Client) Extract(ctx context.Context, imageRef, output string) error {
	if imageRef == "" {
		return fmt.Errorf("no image ref provided")
	}

	img, err := c.extractImage(ctx, imageRef)
	if err != nil {
		return err
	}
	defer img.Close()

	c.logger.Debug("Creating output directory: %s", output)
	if err = os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("create the output path '%s': %w", output, err)
	}

	c.logger.Debug("Extracting...")
	res, err := tools.CopyTar(ctx, img, output)
	if err != nil {
		return fmt.Errorf("copy the image '%s' to the target path '%s': %w", imageRef, output, err)
	}

	c.logger.Info(res.String())

	return nil
}
