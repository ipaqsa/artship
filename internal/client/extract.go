package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ipaqsa/artship/internal/logs"
	"github.com/ipaqsa/artship/internal/tools"
)

// progressTracker interface for tracking bytes
type progressTracker interface {
	Written() int64
}

// spinner provides a simple loading animation
type spinner struct {
	frames  []string
	message string
	stop    chan bool
	wg      sync.WaitGroup
	tracker progressTracker
}

func newSpinner(message string) *spinner {
	return &spinner{
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
		stop:    make(chan bool),
	}
}

func (s *spinner) setTracker(tracker progressTracker) {
	s.tracker = tracker
}

func (s *spinner) start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K") // Clear the line
				return
			default:
				msg := s.message
				if s.tracker != nil {
					written := s.tracker.Written()
					if written > 0 {
						msg = fmt.Sprintf("%s (%s)", s.message, logs.Gray(tools.FormatSize(written)))
					}
				}
				fmt.Printf("\r%s %s", s.frames[i%len(s.frames)], msg)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *spinner) stopSpinner() {
	close(s.stop)
	s.wg.Wait()
	fmt.Print("\r\033[K") // Clear the line
}

// bytesWriter wraps an io.Writer and tracks bytes written
type bytesWriter struct {
	writer  io.Writer
	written int64
	mu      sync.Mutex
}

func newBytesWriter(w io.Writer) *bytesWriter {
	return &bytesWriter{writer: w}
}

func (bw *bytesWriter) Write(p []byte) (int, error) {
	n, err := bw.writer.Write(p)
	bw.mu.Lock()
	bw.written += int64(n)
	bw.mu.Unlock()
	return n, err
}

func (bw *bytesWriter) Written() int64 {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	return bw.written
}

// bytesReader wraps an io.Reader and tracks bytes read
type bytesReader struct {
	reader io.Reader
	read   int64
	mu     sync.Mutex
}

func newBytesReader(r io.Reader) *bytesReader {
	return &bytesReader{reader: r}
}

func (br *bytesReader) Read(p []byte) (int, error) {
	n, err := br.reader.Read(p)
	br.mu.Lock()
	br.read += int64(n)
	br.mu.Unlock()
	return n, err
}

func (br *bytesReader) Close() error {
	if closer, ok := br.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (br *bytesReader) Written() int64 {
	br.mu.Lock()
	defer br.mu.Unlock()
	return br.read
}

// Extract extracts all files from an OCI image
func (c *Client) Extract(ctx context.Context, imageRef, output string) error {
	startTime := time.Now()

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

	// Wrap image reader to track progress
	br := newBytesReader(img)
	defer br.Close()

	// Start spinner with progress tracking
	spin := newSpinner(logs.Green("Extracting image files..."))
	spin.setTracker(br)
	spin.start()

	c.logger.Debug("Creating output directory: %s", output)
	if err = os.MkdirAll(output, 0755); err != nil {
		spin.stopSpinner()
		return fmt.Errorf("create the output path '%s': %w", output, err)
	}

	c.logger.Debug("Extracting...")
	res, err := tools.CopyTar(ctx, br, output)
	if err != nil {
		spin.stopSpinner()
		return fmt.Errorf("copy the image '%s' to the target path '%s': %w", imageRef, output, err)
	}

	spin.stopSpinner()

	executionTime := time.Since(startTime)
	c.logger.Info(logs.BoldGreen("✓")+" Successfully extracted image: %s", logs.Blue(output))
	c.logger.Info(logs.Green("  📁 Files extracted: ")+"%d", res.FilesExtracted)
	c.logger.Info(logs.Green("  📂 Directories created: ")+"%d", res.DirsCreated)
	c.logger.Info(logs.Green("  🔗 Links created: ")+"%d", res.LinksCreated)
	c.logger.Info(logs.Green("  💾 Total size: ")+"%s", tools.FormatSize(res.TotalSize))
	c.logger.Info(logs.Green("  ⏱  Time: ")+"%s", executionTime.Round(time.Millisecond).String())

	return nil
}

// ExtractTar extracts raw tar archive from an OCI image
func (c *Client) ExtractTar(ctx context.Context, imageRef string, output string) error {
	startTime := time.Now()

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

	// Create bytes writer to track progress
	bw := newBytesWriter(out)

	// Start spinner with progress tracking
	spin := newSpinner(logs.Green("Exporting image to tar archive..."))
	spin.setTracker(bw)
	spin.start()

	// Copy the tar stream directly to the file with progress tracking
	copied, err := io.Copy(bw, img)
	if err != nil {
		spin.stopSpinner()
		return fmt.Errorf("copy tar stream: %w", err)
	}

	spin.stopSpinner()

	executionTime := time.Since(startTime)
	c.logger.Info(logs.BoldGreen("✓")+" Successfully exported tar archive: %s", logs.Blue(output))
	c.logger.Info(logs.Green("  📦 Archive size: ")+"%s", tools.FormatSize(copied))
	c.logger.Info(logs.Green("  ⏱  Time: ")+"%s", executionTime.Round(time.Millisecond).String())
	return nil
}
