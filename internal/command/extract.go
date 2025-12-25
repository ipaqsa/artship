package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
)

func init() {
	extractCmd.Flags().StringVarP(&output, "output", "o", "", "Target directory to extract all files")
	extractCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	extractCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	extractCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	extractCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	extractCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	extractCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	_ = extractCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(extractCmd)
}

var extractCmd = &cobra.Command{
	Use:   "extract <image>",
	Short: "Extract all files and directories from an OCI/Docker image",
	Long: `Extract downloads an OCI/Docker image from a registry and extracts
all files, directories, and links to the target directory on the local filesystem.

This command copies the entire filesystem from the image to your local machine,
preserving the directory structure, file permissions, and symbolic links.

To export the image as a tar archive instead, use the 'export' command.`,
	Example: `  # Extract all files from nginx image to a directory
  artship extract nginx:latest -o ./extracted

  # Extract all files to a specific directory
  artship extract alpine:latest -o ./extracted-alpine

  # Extract from a private registry
  artship extract my-registry.com/myapp:v1.0 -o ./extracted-app`,
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

		if err := cli.Extract(cmd.Context(), args[0], output); err != nil {
			return fmt.Errorf("failed to extract image: %w", err)
		}

		return nil
	},
}
