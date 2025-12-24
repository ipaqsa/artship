package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
	"github.com/ipaqsa/artship/internal/tools"
)

func init() {
	metaCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	metaCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	metaCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	metaCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	metaCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	metaCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	rootCmd.AddCommand(metaCmd)
}

var metaCmd = &cobra.Command{
	Use:   "meta <image>",
	Short: "Show OCI/Docker image meta info about",
	Long: `Meta displays detailed metadata information about an OCI/Docker image,
including manifest data, configuration, layer information, and other
image properties. This is useful for understanding image structure
before extracting artifacts.`,
	Example: `  # Show image metadata
  artship meta nginx:latest`,
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

		meta, err := cli.GetImageMeta(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get the image metadata: %w", err)
		}

		logger.Info("")
		logger.Info(tools.BoldBlue("Image Metadata:"))
		logger.Info(tools.Gray("─────────────────────────────────────────────────────────────"))
		logger.Info("%s", meta.String())

		return nil
	},
}
