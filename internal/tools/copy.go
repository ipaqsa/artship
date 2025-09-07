package tools

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CopyResult struct {
	FilesExtracted int64
	DirsCreated    int64
	LinksCreated   int64
	TotalSize      int64
	ExecutionTime  time.Duration
}

// String returns a formatted string representation of the copy result
func (r CopyResult) String() string {
	return fmt.Sprintf("Successfully extracted image:\n"+
		"  Files extracted: %d\n"+
		"  Directories created: %d\n"+
		"  Links created: %d\n"+
		"  Total size: %s\n"+
		"  Execution time: %v",
		r.FilesExtracted, r.DirsCreated, r.LinksCreated,
		FormatSize(r.TotalSize), r.ExecutionTime)
}

// Print outputs the result (deprecated: use String() method)
func (r CopyResult) Print() {
	fmt.Print(r.String() + "\n")
}

func CopyTar(_ context.Context, rc io.ReadCloser, output string) (CopyResult, error) {
	startTime := time.Now()
	res := CopyResult{}

	reader := tar.NewReader(rc)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, fmt.Errorf("read tar header: %w", err)
		}

		// Skip whiteout files (Docker layer deletion markers)
		if strings.Contains(header.Name, ".wh.") {
			continue
		}

		target := filepath.Join(output, header.Name)

		// Ensure the path is within our output directory (security check)
		if !strings.HasPrefix(target, filepath.Clean(output)+string(os.PathSeparator)) {
			fmt.Printf("Warning: Skipping path outside target directory: '%s'\n", header.Name)
			continue
		}

		switch header.Typeflag {
		case tar.TypeReg:
			if err = extractFile(reader, header, target); err != nil {
				return res, fmt.Errorf("extract the file '%s': %w", header.Name, err)
			}

			res.FilesExtracted++
			res.TotalSize += header.Size

		case tar.TypeDir:
			if err = extractDir(header, target); err != nil {
				return res, fmt.Errorf("extract the directory '%s': %w", header.Name, err)
			}

			res.DirsCreated++

		case tar.TypeSymlink, tar.TypeLink:
			if err = extractLink(header, target); err != nil {
				return res, fmt.Errorf("extract the link '%s': %w", header.Name, err)
			}

			res.LinksCreated++

		default:
			// Other types (character devices, block devices, FIFOs, etc.)
			fmt.Printf("Warning: Skipping unsupported file type %d for '%s'\n", header.Typeflag, header.Name)
		}
	}

	res.ExecutionTime = time.Since(startTime)

	return res, nil
}

func CopyArtifact(r io.Reader, header *tar.Header, output string) error {
	// Determine the target path
	targetPath := output

	// If output is a directory, use the artifact's name within that directory
	if stat, err := os.Stat(output); err == nil && stat.IsDir() {
		// Use just the base name of the artifact, not the full path
		targetPath = filepath.Join(output, filepath.Base(header.Name))
	} else if strings.HasSuffix(output, "/") {
		// If output ends with /, treat it as a directory even if it doesn't exist yet
		targetPath = filepath.Join(output, filepath.Base(header.Name))
	}

	switch header.Typeflag {
	case tar.TypeReg:
		if err := extractFile(r, header, targetPath); err != nil {
			return fmt.Errorf("extract the file '%s': %w", header.Name, err)
		}

	case tar.TypeDir:
		if err := extractDir(header, targetPath); err != nil {
			return fmt.Errorf("extract directory '%s': %w", header.Name, err)
		}
	case tar.TypeSymlink, tar.TypeLink:
		if err := extractLink(header, targetPath); err != nil {
			return fmt.Errorf("extract the link '%s': %w", header.Name, err)
		}
	}

	return nil
}

// extractFile extracts a single file from the tar reader to the filesystem
func extractFile(reader io.Reader, header *tar.Header, targetPath string) error {
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create the target dir '%s': %w", targetDir, err)
	}

	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
	if err != nil {
		return fmt.Errorf("create the target file '%s': %w", targetPath, err)
	}
	defer target.Close()

	if _, err = io.Copy(target, reader); err != nil {
		return fmt.Errorf("copy file content: %w", err)
	}

	return nil
}

// extractDir creates the target directory structure
func extractDir(header *tar.Header, targetPath string) error {
	if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
		return fmt.Errorf("create the target dir '%s': %w", targetPath, err)
	}

	return nil
}

// extractLink creates symbolic or hard links
func extractLink(header *tar.Header, targetPath string) error {
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create the target dir '%s': %w", targetDir, err)
	}

	// Remove existing file/link if it exists
	if _, err := os.Lstat(targetPath); err == nil {
		if err = os.Remove(targetPath); err != nil {
			return fmt.Errorf("remove existing link '%s': %w", targetPath, err)
		}
	}

	if header.Typeflag == tar.TypeSymlink {
		if err := os.Symlink(header.Linkname, targetPath); err != nil {
			return fmt.Errorf("create symlink '%s' -> '%s': %w", targetPath, header.Linkname, err)
		}

		return nil
	}

	if header.Typeflag == tar.TypeLink {
		// For hard links, we need the target to be relative to our extraction directory
		linkTarget := filepath.Join(filepath.Dir(targetPath), header.Linkname)
		if err := os.Link(linkTarget, targetPath); err != nil {
			// If hard link fails, create a copy instead
			fmt.Printf("Warning: Could not create hard link, creating copy instead: %s -> %s\n", targetPath, header.Linkname)
			return nil
		}
	}

	return nil
}
