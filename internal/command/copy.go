package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"artship/internal/client"
	"artship/internal/logs"
)

var (
	artifacts []string
	output    string
)

func init() {
	copyCmd.Flags().StringSliceVarP(&artifacts, "artifact", "a", []string{}, "Artifact names to extract (files or directories, required)")
	copyCmd.Flags().StringVarP(&output, "output", "o", ".", "Target path for the extracted artifacts (required)")
	copyCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	copyCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	copyCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	copyCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	copyCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	copyCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	_ = copyCmd.MarkFlagRequired("artifact")

	rootCmd.AddCommand(copyCmd)
}

var copyCmd = &cobra.Command{
	Use:   "cp <image> -a <artifact> -o <output>",
	Short: "cp artifacts from an OCI/Docker image",
	Long: `Copy downloads an OCI/Docker image from a registry and extracts 
specific files or directories to target paths on the local filesystem.

Multiple artifacts can be extracted in a single operation by specifying
multiple --artifact and --output flags. The number of artifacts and outputs
must match.`,
	Example: `  # Copy a single binary from nginx image
  artship cp nginx:latest --artifact nginx --output /usr/local/bin
  
  # Copy multiple artifacts with short flags
  artship cp alpine:latest -a sh -a ls -o ./bin
  
  # Copy directories and files
  artship cp myapp:latest --artifact /app/bin --artifact /app/config --output ./local
  
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

		if err := cli.Copy(cmd.Context(), args[0], artifacts, output); err != nil {
			return fmt.Errorf("failed to copy artifacts: %w", err)
		}

		return nil
	},
}
