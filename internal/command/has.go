package command

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
)

func init() {
	hasCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	hasCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	hasCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	hasCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	hasCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	hasCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	rootCmd.AddCommand(hasCmd)
}

var hasCmd = &cobra.Command{
	Use:   "has <image> <artifact>",
	Short: "Check if an artifact exists in an OCI/Docker image",
	Long: `Has checks whether a specific artifact exists in an OCI/Docker image
without downloading or extracting it. This is useful for quickly verifying
the presence of files or directories before performing operations.`,
	Example: `  # Check if nginx binary exists
  artship has nginx:latest nginx
  
  # Check for configuration file
  artship has nginx:latest /etc/nginx/nginx.conf
  
  # Check with authentication
  artship has private-registry.com/app:latest myapp -u user -p pass`,
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

		if err := cli.Has(cmd.Context(), args[0], args[1]); err != nil {
			if !errors.Is(err, client.ErrNotFound) {
				return fmt.Errorf("failed to check artifact: %w", err)
			}

			logger.Info(logs.Red("✗")+" Artifact %s not found in %s", logs.Yellow(args[1]), logs.Blue(args[0]))
			return nil
		}

		logger.Info(logs.Green("✓")+" Artifact %s found in %s", logs.Yellow(args[1]), logs.Blue(args[0]))

		return nil
	},
}
