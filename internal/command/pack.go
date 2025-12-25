package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
)

func init() {
	packCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	packCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	packCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	packCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	packCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	packCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	rootCmd.AddCommand(packCmd)
}

var packCmd = &cobra.Command{
	Use:   "pack <image> <source>",
	Short: "Pack files/directories into an OCI/Docker image",
	Long: `Pack creates an OCI/Docker image from local files or directories.
The source can be a single file or an entire directory structure.

The created image is pushed to a registry.`,
	Example: `  # Pack a directory into an image
  artship pack myapp:latest ./myapp

  # Pack a file into an image and push to registry
  artship pack myregistry.com/myfile:v1.0 ./myfile

  # Pack with authentication
  artship pack private.registry.com/app:latest ./app -u username -p password`,
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

		if err := cli.Pack(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to pack: %w", err)
		}

		return nil
	},
}
