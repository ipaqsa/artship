package tools

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrStopWalk = errors.New("stop walk")

// WalkTar walks over the tar archive
func WalkTar(rc io.ReadCloser, f func(r io.Reader, header *tar.Header) error) error {
	reader := tar.NewReader(rc)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar header: %w", err)
		}

		// Skip whiteout files (Docker layer deletion markers)
		if strings.Contains(header.Name, ".wh.") {
			continue
		}

		if err = f(reader, header); err != nil {
			if errors.Is(err, ErrStopWalk) {
				return nil
			}

			return err
		}
	}

	return nil
}

// GetArtifactType converts tar type flag to human-readable type
func GetArtifactType(typeFlag byte) string {
	switch typeFlag {
	case tar.TypeReg:
		return "file"
	case tar.TypeDir:
		return "dir"
	case tar.TypeSymlink:
		return "symlink"
	case tar.TypeLink:
		return "hardlink"
	case tar.TypeChar:
		return "chardev"
	case tar.TypeBlock:
		return "blockdev"
	case tar.TypeFifo:
		return "fifo"
	default:
		return "unknown"
	}
}
