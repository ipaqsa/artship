package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"artship/internal/client"
	"artship/internal/logs"
)

func init() {
	catCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	catCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	catCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	catCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	catCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	catCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	rootCmd.AddCommand(catCmd)
}

var catCmd = &cobra.Command{
	Use:   "cat <image> <artifact>",
	Short: "Show the content of an artifact from an OCI/Docker image",
	Long: `Cat prints the content of a specific file artifact from an OCI/Docker image
to stdout. This is useful for examining configuration files, scripts, or other
text-based artifacts without extracting them to the filesystem.`,
	Example: `  # Show content of a configuration file
  artship cat nginx:latest /etc/nginx/nginx.conf
  
  # Show content of passwd file
  artship cat alpine:latest /etc/passwd`,
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

		content, err := cli.Cat(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to cat artifact: %w", err)
		}

		logger.Info(string(content))

		return nil
	},
}
