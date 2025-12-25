package client

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ipaqsa/artship/internal/logs"
	"github.com/ipaqsa/artship/internal/tools"
)

// Extract extracts all files from an OCI image
func (c *Client) Extract(ctx context.Context, imageRef, output string) error {
	if imageRef == "" {
		return fmt.Errorf("no image ref provided")
	}

	if len(output) == 0 {
		return fmt.Errorf("no output provided")
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

// ExtractTar extracts raw tar
func (c *Client) ExtractTar(ctx context.Context, imageRef string, output string) error {
	if imageRef == "" {
		return fmt.Errorf("no image ref provided")
	}

	if output == "" {
		return fmt.Errorf("no output provided")
	}

	img, err := c.extractImage(ctx, imageRef)
	if err != nil {
		return err
	}
	defer img.Close()

	// Create or open the output file
	out, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	// Copy the tar stream directly to the file
	copied, err := io.Copy(out, img)
	if err != nil {
		return fmt.Errorf("copy tar stream: %w", err)
	}

	c.logger.Info(logs.BoldGreen("✓")+" Successfully extracted tar archive: %s %s", logs.Blue(output), logs.Gray("("+tools.FormatSize(copied)+")"))
	return nil
}
