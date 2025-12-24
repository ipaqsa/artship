package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
	"github.com/ipaqsa/artship/internal/tools"
)

var (
	filter   string
	detailed bool
	layer    string
)

func init() {
	listCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	listCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	listCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	listCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	listCmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter by type: file, dir, symlink, hardlink, all")
	listCmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed info (size, type, permissions)")
	listCmd.Flags().StringVarP(&layer, "layer", "l", "", "Show files from specific layer (layer digest)")
	listCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	listCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "ls <image>",
	Short: "Show available artifacts in an OCI/Docker image",
	Long: `Show all files and directories available in an OCI/Docker image.
This command downloads the image layers and lists all available artifacts`,
	Example: `  # List all artifacts in an image
  artship ls nginx:latest

  # List artifacts with detailed info
  artship ls nginx:latest --detailed

  # List only files
  artship ls nginx:latest --filter file

  # List files from specific layer
  artship ls nginx:latest --layer sha256:abc123...

  # List directories with info
  artship ls nginx:latest -f dir -d`,
	Args: cobra.ExactArgs(1),
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

		artifacts, err := cli.List(cmd.Context(), args[0], filter, layer)
		if err != nil {
			return fmt.Errorf("failed to list artifacts: %w", err)
		}

		logger.Info("")
		logger.Info(tools.BoldBlue("Image artifacts:"))
		logger.Info(tools.Gray("─────────────────────────────────────────────────────────────"))
		logger.Info(artifacts.String(detailed))

		return nil
	},
}
