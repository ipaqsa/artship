package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"artship/internal/client"
	"artship/internal/logs"
)

func init() {
	infoCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	infoCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	infoCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	infoCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	infoCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	infoCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info <image> <artifact>",
	Short: "Show detailed information about an artifact from an OCI/Docker image",
	Long: `Info displays detailed metadata information about a specific artifact in an OCI/Docker image,
including type, size, permissions, and link targets. This is useful for understanding
artifact properties before extracting them.`,
	Example: `  # Show detailed info about nginx binary
  artship info nginx:latest nginx
  
  # Show info about configuration file
  artship info nginx:latest /etc/nginx/nginx.conf
  
  # Show info with authentication
  artship info private-registry.com/app:latest myapp -u user -p pass`,
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

		// Get detailed artifact information
		info, err := cli.GetArtifact(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to get artifact info: %w", err)
		}

		logger.Info("\nArtifact Info: \n%s", info.String(true))

		return nil
	},
}
