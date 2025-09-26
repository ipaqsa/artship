package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
)

func init() {
	packageCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	packageCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	packageCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	packageCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	packageCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	packageCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	rootCmd.AddCommand(packageCmd)
}

var packageCmd = &cobra.Command{
	Use:   "package <image> <source>",
	Short: "Package files/directories into an OCI/Docker image",
	Long: `Package creates an OCI/Docker image from local files or directories.
The source can be a single file or an entire directory structure.

The created image can optionally be pushed to a registry using the --push flag.`,
	Example: `  # Package a directory into an image (local only)
  artship package myapp:latest ./myapp

  # Package a file into an image and push to registry
  artship package myregistry.com/myfile:v1.0 ./myfile

  # Package with authentication
  artship package private.registry.com/app:latest ./app -u username -p password`,
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

		if err := cli.Package(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to package: %w", err)
		}

		return nil
	},
}
