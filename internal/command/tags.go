package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
)

func init() {
	tagsCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	tagsCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	tagsCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	tagsCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	tagsCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	tagsCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	rootCmd.AddCommand(tagsCmd)
}

var tagsCmd = &cobra.Command{
	Use:   "tags <repository>",
	Short: "List available tags for an OCI/Docker repository",
	Long: `List all available tags for a specific OCI/Docker repository.
This command queries the registry and displays all tags in alphabetical order.`,
	Example: `  # List all tags for nginx repository
  artship tags nginx
  
  # List tags for a specific registry
  artship tags gcr.io/my-project/my-app
  
  # List tags with authentication
  artship tags private-registry.com/app -u user -p pass`,
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

		tags, err := cli.Tags(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list tags: %w", err)
		}

		logger.Info("Available tags:")
		for _, tag := range tags {
			logger.Info(" " + tag)
		}

		return nil
	},
}
