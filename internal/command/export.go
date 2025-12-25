package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
)

func init() {
	exportCmd.Flags().StringVarP(&output, "output", "o", "", "Target file path for the tar archive")
	exportCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registry authentication")
	exportCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registry authentication")
	exportCmd.Flags().StringVarP(&token, "token", "t", "", "Token for registry authentication")
	exportCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth for registry authentication")
	exportCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	exportCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	_ = exportCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export <image>",
	Short: "Export an OCI/Docker image as a tar archive",
	Long: `Export downloads an OCI/Docker image from a registry and saves it as a tar archive file.

This command downloads the image and saves it as a raw tar file that contains all layers
and metadata. The tar file can be imported into Docker or other container runtimes.`,
	Example: `  # Export nginx image as a tar archive
  artship export nginx:latest -o ./nginx.tar

  # Export alpine image to a specific file
  artship export alpine:latest -o ./alpine-image.tar

  # Export from a private registry
  artship export my-registry.com/myapp:v1.0 -o ./myapp.tar -u username -p password`,
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

		if err := cli.ExtractTar(cmd.Context(), args[0], output); err != nil {
			return fmt.Errorf("failed to export image: %w", err)
		}

		return nil
	},
}
