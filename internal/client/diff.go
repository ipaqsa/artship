package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/ipaqsa/artship/internal/logs"
	"github.com/ipaqsa/artship/internal/tools"
)

// FileStatus represents the status of a file in diff comparison
type FileStatus string

const (
	// File status constants
	FileStatusAdded     FileStatus = "added"
	FileStatusRemoved   FileStatus = "removed"
	FileStatusModified  FileStatus = "modified"
	FileStatusUnchanged FileStatus = "unchanged"
)

// FileInfo represents information about a file in an image
type FileInfo struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Mode string `json:"mode"`
	Hash string `json:"hash,omitempty"` // SHA256 hash for content comparison
	Type string `json:"type"`
}

// DiffEntry represents a single difference between images
type DiffEntry struct {
	Path    string     `json:"path"`
	Status  FileStatus `json:"status"`
	OldSize int64      `json:"old_size,omitempty"`
	NewSize int64      `json:"new_size,omitempty"`
	OldMode string     `json:"old_mode,omitempty"`
	NewMode string     `json:"new_mode,omitempty"`
	Type    string     `json:"type"`
}

// DiffResult contains the comparison results
type DiffResult struct {
	SourceImage  string      `json:"source_image"`
	TargetImage  string      `json:"target_image"`
	Added        []DiffEntry `json:"added"`
	Removed      []DiffEntry `json:"removed"`
	Modified     []DiffEntry `json:"modified"`
	Unchanged    []DiffEntry `json:"unchanged,omitempty"`
	TotalAdded   int         `json:"total_added"`
	TotalRemoved int         `json:"total_removed"`
	TotalChanged int         `json:"total_changed"`
}

// String returns formatted diff output with colors
func (r *DiffResult) String(showUnchanged bool) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(logs.BoldBlue(fmt.Sprintf("Comparing %s → %s", r.SourceImage, r.TargetImage)))
	sb.WriteString("\n")
	sb.WriteString(logs.Gray("─────────────────────────────────────────────────────────────"))
	sb.WriteString("\n\n")

	// Summary
	sb.WriteString(logs.BoldGreen(fmt.Sprintf("+ Added:    %d files\n", r.TotalAdded)))
	sb.WriteString(logs.BoldRed(fmt.Sprintf("- Removed:  %d files\n", r.TotalRemoved)))
	sb.WriteString(logs.BoldYellow(fmt.Sprintf("~ Modified: %d files\n", r.TotalChanged)))
	sb.WriteString("\n")
	sb.WriteString(logs.Gray("─────────────────────────────────────────────────────────────"))
	sb.WriteString("\n\n")

	// Added files
	if len(r.Added) > 0 {
		sb.WriteString(logs.BoldGreen("Added files:\n"))
		for _, entry := range r.Added {
			details := fmt.Sprintf("(%s, %s)", entry.Type, tools.FormatSize(entry.NewSize))
			sb.WriteString(logs.FormatDiffLine(string(FileStatusAdded), entry.Path, details))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Removed files
	if len(r.Removed) > 0 {
		sb.WriteString(logs.BoldRed("Removed files:\n"))
		for _, entry := range r.Removed {
			details := fmt.Sprintf("(%s, %s)", entry.Type, tools.FormatSize(entry.OldSize))
			sb.WriteString(logs.FormatDiffLine(string(FileStatusRemoved), entry.Path, details))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Modified files
	if len(r.Modified) > 0 {
		sb.WriteString(logs.BoldYellow("Modified files:\n"))
		for _, entry := range r.Modified {
			var details string
			if entry.OldSize != entry.NewSize {
				details = fmt.Sprintf("(%s → %s)", tools.FormatSize(entry.OldSize), tools.FormatSize(entry.NewSize))
			}
			if entry.OldMode != entry.NewMode {
				if details != "" {
					details += fmt.Sprintf(", mode: %s → %s", entry.OldMode, entry.NewMode)
				} else {
					details = fmt.Sprintf("(mode: %s → %s)", entry.OldMode, entry.NewMode)
				}
			}
			sb.WriteString(logs.FormatDiffLine(string(FileStatusModified), entry.Path, details))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Unchanged files (optional)
	if showUnchanged && len(r.Unchanged) > 0 {
		sb.WriteString(logs.Gray(fmt.Sprintf("Unchanged files: %d\n", len(r.Unchanged))))
		for _, entry := range r.Unchanged {
			details := fmt.Sprintf("(%s)", tools.FormatSize(entry.NewSize))
			sb.WriteString(logs.FormatDiffLine(string(FileStatusUnchanged), entry.Path, details))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ToJSON returns JSON representation of the diff result
func (r *DiffResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal diff result to JSON: %w", err)
	}
	return string(data), nil
}

// Diff compares two images and returns the differences
func (c *Client) Diff(ctx context.Context, image1Ref, image2Ref string, includeUnchanged bool) (*DiffResult, error) {
	if image1Ref == "" || image2Ref == "" {
		return nil, fmt.Errorf("both image references must be provided")
	}

	c.logger.Info("Analyzing %s...", image1Ref)
	sourceFiles, err := c.getFileMap(ctx, image1Ref)
	if err != nil {
		return nil, fmt.Errorf("get files from %s: %w", image1Ref, err)
	}

	c.logger.Info("Analyzing %s...", image2Ref)
	targetFiles, err := c.getFileMap(ctx, image2Ref)
	if err != nil {
		return nil, fmt.Errorf("get files from %s: %w", image2Ref, err)
	}

	c.logger.Debug("Comparing %d files from source with %d files from target", len(sourceFiles), len(targetFiles))

	result := &DiffResult{
		SourceImage: image1Ref,
		TargetImage: image2Ref,
		Added:       []DiffEntry{},
		Removed:     []DiffEntry{},
		Modified:    []DiffEntry{},
	}

	if includeUnchanged {
		result.Unchanged = []DiffEntry{}
	}

	// Find removed and modified files
	for path, sourceInfo := range sourceFiles {
		if targetInfo, exists := targetFiles[path]; exists {
			// File exists in both images
			if sourceInfo.Size != targetInfo.Size || sourceInfo.Mode != targetInfo.Mode {
				// File was modified
				result.Modified = append(result.Modified, DiffEntry{
					Path:    path,
					Status:  FileStatusModified,
					OldSize: sourceInfo.Size,
					NewSize: targetInfo.Size,
					OldMode: sourceInfo.Mode,
					NewMode: targetInfo.Mode,
					Type:    targetInfo.Type,
				})
				result.TotalChanged++
			} else if includeUnchanged {
				// File is unchanged
				result.Unchanged = append(result.Unchanged, DiffEntry{
					Path:    path,
					Status:  FileStatusUnchanged,
					NewSize: targetInfo.Size,
					NewMode: targetInfo.Mode,
					Type:    targetInfo.Type,
				})
			}
		} else {
			// File was removed
			result.Removed = append(result.Removed, DiffEntry{
				Path:    path,
				Status:  FileStatusRemoved,
				OldSize: sourceInfo.Size,
				OldMode: sourceInfo.Mode,
				Type:    sourceInfo.Type,
			})
			result.TotalRemoved++
		}
	}

	// Find added files
	for path, targetInfo := range targetFiles {
		if _, exists := sourceFiles[path]; !exists {
			// File was added
			result.Added = append(result.Added, DiffEntry{
				Path:    path,
				Status:  FileStatusAdded,
				NewSize: targetInfo.Size,
				NewMode: targetInfo.Mode,
				Type:    targetInfo.Type,
			})
			result.TotalAdded++
		}
	}

	// Sort results for consistent output
	sort.Slice(result.Added, func(i, j int) bool { return result.Added[i].Path < result.Added[j].Path })
	sort.Slice(result.Removed, func(i, j int) bool { return result.Removed[i].Path < result.Removed[j].Path })
	sort.Slice(result.Modified, func(i, j int) bool { return result.Modified[i].Path < result.Modified[j].Path })
	if includeUnchanged {
		sort.Slice(result.Unchanged, func(i, j int) bool { return result.Unchanged[i].Path < result.Unchanged[j].Path })
	}

	return result, nil
}

// getFileMap returns a map of file paths to file info for an image
func (c *Client) getFileMap(ctx context.Context, imageRef string) (map[string]*FileInfo, error) {
	artifacts, err := c.List(ctx, imageRef, "all", "")
	if err != nil {
		return nil, err
	}

	fileMap := make(map[string]*FileInfo)
	for _, artifact := range artifacts {
		// Skip directories for diff (only compare files)
		if artifact.Type == "dir" {
			continue
		}

		fileMap[artifact.Path] = &FileInfo{
			Path: artifact.Path,
			Size: artifact.Size,
			Mode: artifact.Mode,
			Type: artifact.Type,
		}
	}

	return fileMap, nil
}
