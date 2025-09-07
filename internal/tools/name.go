package tools

import (
	"path/filepath"
	"strings"
)

// MatchName checks if a filepath matches the requested
func MatchName(filePath, artifact string) bool {
	// Direct match
	if filePath == artifact {
		return true
	}

	// Match by filename only (for simple binary names)
	if filepath.Base(filePath) == artifact {
		return true
	}

	// Match if file is inside the requested directory
	if strings.HasPrefix(filePath, strings.TrimSuffix(artifact, "/")+"/") {
		return true
	}

	return false
}
