package client

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/ipaqsa/artship/internal/tools"
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
	Path    string `json:"path"`
	Status  string `json:"status"` // added, removed, modified, unchanged
	OldSize int64  `json:"old_size,omitempty"`
	NewSize int64  `json:"new_size,omitempty"`
	OldMode string `json:"old_mode,omitempty"`
	NewMode string `json:"new_mode,omitempty"`
	Type    string `json:"type"`
}

// DiffResult contains the comparison results
type DiffResult struct {
	Image1       string      `json:"image1"`
	Image2       string      `json:"image2"`
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
	result := fmt.Sprintf("\n%s\n", tools.BoldBlue(fmt.Sprintf("Comparing %s → %s", r.Image1, r.Image2)))
	result += fmt.Sprintf("%s\n\n", tools.Gray("─────────────────────────────────────────────────────────────"))

	// Summary
	result += tools.BoldGreen(fmt.Sprintf("+ Added:    %d files\n", r.TotalAdded))
	result += tools.BoldRed(fmt.Sprintf("- Removed:  %d files\n", r.TotalRemoved))
	result += tools.BoldYellow(fmt.Sprintf("~ Modified: %d files\n", r.TotalChanged))
	result += fmt.Sprintf("\n%s\n\n", tools.Gray("─────────────────────────────────────────────────────────────"))

	// Added files
	if len(r.Added) > 0 {
		result += tools.BoldGreen("Added files:\n")
		for _, entry := range r.Added {
			details := fmt.Sprintf("(%s, %s)", entry.Type, tools.FormatSize(entry.NewSize))
			result += tools.FormatDiffLine("added", entry.Path, details) + "\n"
		}
		result += "\n"
	}

	// Removed files
	if len(r.Removed) > 0 {
		result += tools.BoldRed("Removed files:\n")
		for _, entry := range r.Removed {
			details := fmt.Sprintf("(%s, %s)", entry.Type, tools.FormatSize(entry.OldSize))
			result += tools.FormatDiffLine("removed", entry.Path, details) + "\n"
		}
		result += "\n"
	}

	// Modified files
	if len(r.Modified) > 0 {
		result += tools.BoldYellow("Modified files:\n")
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
			result += tools.FormatDiffLine("modified", entry.Path, details) + "\n"
		}
		result += "\n"
	}

	// Unchanged files (optional)
	if showUnchanged && len(r.Unchanged) > 0 {
		result += tools.Gray(fmt.Sprintf("Unchanged files: %d\n", len(r.Unchanged)))
		for _, entry := range r.Unchanged {
			details := fmt.Sprintf("(%s)", tools.FormatSize(entry.NewSize))
			result += tools.FormatDiffLine("unchanged", entry.Path, details) + "\n"
		}
	}

	return result
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
	files1, err := c.getFileMap(ctx, image1Ref)
	if err != nil {
		return nil, fmt.Errorf("get files from %s: %w", image1Ref, err)
	}

	c.logger.Info("Analyzing %s...", image2Ref)
	files2, err := c.getFileMap(ctx, image2Ref)
	if err != nil {
		return nil, fmt.Errorf("get files from %s: %w", image2Ref, err)
	}

	c.logger.Debug("Comparing %d files from image1 with %d files from image2", len(files1), len(files2))

	result := &DiffResult{
		Image1:   image1Ref,
		Image2:   image2Ref,
		Added:    []DiffEntry{},
		Removed:  []DiffEntry{},
		Modified: []DiffEntry{},
	}

	if includeUnchanged {
		result.Unchanged = []DiffEntry{}
	}

	// Find removed and modified files
	for path, info1 := range files1 {
		if info2, exists := files2[path]; exists {
			// File exists in both images
			if info1.Size != info2.Size || info1.Mode != info2.Mode {
				// File was modified
				result.Modified = append(result.Modified, DiffEntry{
					Path:    path,
					Status:  "modified",
					OldSize: info1.Size,
					NewSize: info2.Size,
					OldMode: info1.Mode,
					NewMode: info2.Mode,
					Type:    info2.Type,
				})
				result.TotalChanged++
			} else if includeUnchanged {
				// File is unchanged
				result.Unchanged = append(result.Unchanged, DiffEntry{
					Path:    path,
					Status:  "unchanged",
					NewSize: info2.Size,
					NewMode: info2.Mode,
					Type:    info2.Type,
				})
			}
		} else {
			// File was removed
			result.Removed = append(result.Removed, DiffEntry{
				Path:    path,
				Status:  "removed",
				OldSize: info1.Size,
				OldMode: info1.Mode,
				Type:    info1.Type,
			})
			result.TotalRemoved++
		}
	}

	// Find added files
	for path, info2 := range files2 {
		if _, exists := files1[path]; !exists {
			// File was added
			result.Added = append(result.Added, DiffEntry{
				Path:    path,
				Status:  "added",
				NewSize: info2.Size,
				NewMode: info2.Mode,
				Type:    info2.Type,
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

// computeFileHash computes SHA256 hash of a file (for deep content comparison)
// This is expensive and optional - can be added later if needed
func computeFileHash(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
