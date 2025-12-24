package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
)

var (
	jsonOutput    bool
	showUnchanged bool
	noColor       bool
	diffFilter    string
)

func init() {
	diffCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	diffCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	diffCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	diffCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	diffCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	diffCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")
	diffCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output diff results in JSON format")
	diffCmd.Flags().BoolVar(&showUnchanged, "show-unchanged", false, "Show unchanged files in the output")
	diffCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	diffCmd.Flags().StringVarP(&diffFilter, "filter", "f", "", "Filter diff results (added, removed, modified, all)")

	rootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff <image1> <image2>",
	Short: "Show file differences between two OCI/Docker images",
	Long: `Diff compares two OCI/Docker images and shows the differences in their filesystems.

The output includes:
- Added files (files that exist only in image2)
- Removed files (files that exist only in image1)
- Modified files (files that exist in both but have different size or permissions)

Results are displayed with color-coded output by default:
- Green (+) for added files
- Red (-) for removed files
- Yellow (~) for modified files

Use --json flag for machine-readable JSON output.`,
	Example: `  # Compare two versions of the same image
  artship diff nginx:1.24 nginx:1.25

  # Compare with verbose output
  artship diff alpine:3.17 alpine:3.18 -v

  # Output as JSON
  artship diff nginx:latest nginx:alpine --json

  # Show unchanged files
  artship diff redis:7.0 redis:7.2 --show-unchanged

  # Compare images from private registry
  artship diff registry.io/app:v1 registry.io/app:v2 -u user -p pass

  # Filter to show only added files
  artship diff node:18 node:20 --filter added`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logs.New(verbose)

		cli := client.New(&client.Options{
			Username: username,
			Password: password,
			Token:    token,
			Auth:     auth,
			Insecure: insecure,
			Logger:   logger,
		})

		// Perform diff
		result, err := cli.Diff(cmd.Context(), args[0], args[1], showUnchanged)
		if err != nil {
			return fmt.Errorf("failed to compare images: %w", err)
		}

		// Apply filter if specified
		if diffFilter != "" {
			result = filterDiffResult(result, diffFilter)
		}

		// Output results
		if jsonOutput {
			jsonStr, err := result.ToJSON()
			if err != nil {
				return fmt.Errorf("failed to generate JSON output: %w", err)
			}
			fmt.Println(jsonStr)
		} else {
			// TODO: Handle --no-color flag if needed
			// For now, colors are always enabled in non-JSON mode
			fmt.Print(result.String(showUnchanged))
		}

		return nil
	},
}

// filterDiffResult filters the diff result based on the specified filter
func filterDiffResult(result *client.DiffResult, filter string) *client.DiffResult {
	filtered := &client.DiffResult{
		Image1: result.Image1,
		Image2: result.Image2,
	}

	switch filter {
	case "added":
		filtered.Added = result.Added
		filtered.TotalAdded = result.TotalAdded
	case "removed":
		filtered.Removed = result.Removed
		filtered.TotalRemoved = result.TotalRemoved
	case "modified":
		filtered.Modified = result.Modified
		filtered.TotalChanged = result.TotalChanged
	case "all", "":
		return result
	default:
		// Invalid filter, return original result
		return result
	}

	return filtered
}
