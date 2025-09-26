package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
)

var (
	artifacts  []string
	output     string
	extractTar bool
)

func init() {
	copyCmd.Flags().StringSliceVarP(&artifacts, "artifact", "a", []string{}, "Artifact names to extract (files or directories, required)")
	copyCmd.Flags().StringVarP(&output, "output", "o", ".", "Target path for the extracted artifacts (required)")
	copyCmd.Flags().BoolVar(&extractTar, "tar", false, "Extract the entire image as a tar archive")
	copyCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	copyCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	copyCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	copyCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	copyCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	copyCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	copyCmd.MarkFlagsMutuallyExclusive("artifact", "tar")

	rootCmd.AddCommand(copyCmd)
}

var copyCmd = &cobra.Command{
	Use:   "cp <image> [flags]",
	Short: "cp artifacts from an OCI/Docker image",
	Long: `Copy downloads an OCI/Docker image from a registry and extracts
specific files or directories to target paths on the local filesystem.

Multiple artifacts can be extracted in a single operation by specifying
multiple --artifact flags. Alternatively, use --tar to extract the entire
image as a tar archive.`,
	Example: `  # Copy a single binary from nginx image
  artship cp nginx:latest --artifact nginx --output /usr/local/bin

  # Copy multiple artifacts with short flags
  artship cp alpine:latest -a sh -a ls -o ./bin

  # Copy directories and files
  artship cp myapp:latest --artifact /app/bin --artifact /app/config --output ./local

  # Extract entire image as tar archive
  artship cp nginx:latest --tar --output ./nginx.tar

  # Copy from a private registry
  artship cp my-registry.com/myapp:v1.0 --artifact myapp --output ./bin/myapp`,
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

		if extractTar {
			if err := cli.ExtractTar(cmd.Context(), args[0], output); err != nil {
				return fmt.Errorf("failed to extract tar: %w", err)
			}
		} else {
			if len(artifacts) == 0 {
				return fmt.Errorf("no artifacts specified (use --artifact or --tar)")
			}
			if err := cli.Copy(cmd.Context(), args[0], artifacts, output); err != nil {
				return fmt.Errorf("failed to copy artifacts: %w", err)
			}
		}

		return nil
	},
}
