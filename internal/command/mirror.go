package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ipaqsa/artship/internal/client"
	"github.com/ipaqsa/artship/internal/logs"
	"github.com/ipaqsa/artship/internal/tools"
)

var (
	srcUsername string
	srcPassword string
	srcToken    string
	srcAuth     string
	dstUsername string
	dstPassword string
	dstToken    string
	dstAuth     string
)

func init() {
	mirrorCmd.Flags().StringVar(&srcUsername, "src-username", "", "Username for source registry authentication")
	mirrorCmd.Flags().StringVar(&srcPassword, "src-password", "", "Password for source registry authentication")
	mirrorCmd.Flags().StringVar(&srcToken, "src-token", "", "Token for source registry authentication")
	mirrorCmd.Flags().StringVar(&srcAuth, "src-auth", "", "Auth string for source registry authentication")

	mirrorCmd.Flags().StringVar(&dstUsername, "dst-username", "", "Username for destination registry authentication")
	mirrorCmd.Flags().StringVar(&dstPassword, "dst-password", "", "Password for destination registry authentication")
	mirrorCmd.Flags().StringVar(&dstToken, "dst-token", "", "Token for destination registry authentication")
	mirrorCmd.Flags().StringVar(&dstAuth, "dst-auth", "", "Auth string for destination registry authentication")

	mirrorCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure registry connections")
	mirrorCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose debug output")

	rootCmd.AddCommand(mirrorCmd)
}

var mirrorCmd = &cobra.Command{
	Use:   "mirror <source-image> <destination-image>",
	Short: "Copy/mirror an OCI/Docker image from source to destination",
	Long: `Mirror copies an OCI/Docker image from one location to another.

This is useful for:
- Copying images between registries (e.g., Docker Hub → private registry)
- Creating backups of images
- Migrating images to different registries
- Renaming images or changing tags

The command downloads the image from the source registry and uploads it to the
destination registry. Both source and destination can have different authentication
credentials.

Examples of valid image references:
- nginx:latest
- docker.io/library/nginx:1.25
- gcr.io/my-project/my-app:v1.0.0
- myregistry.com:5000/app:latest`,
	Example: `  # Copy from Docker Hub to a private registry
  artship mirror nginx:latest myregistry.com/nginx:latest

  # Copy between different registries with authentication
  artship mirror docker.io/nginx:latest \
    myregistry.com/nginx:latest \
    --dst-username admin --dst-password secret

  # Rename an image tag in the same registry
  artship mirror myregistry.com/app:v1.0 myregistry.com/app:latest

  # Copy from public to private with different credentials
  artship mirror gcr.io/public/app:v1 \
    registry.company.com/app:v1 \
    --src-username _json_key --src-password "$(cat key.json)" \
    --dst-username admin --dst-password secret

  # Copy with verbose output
  artship mirror alpine:3.18 myregistry.com/alpine:3.18 -v

  # Use Docker credentials for source, explicit for destination
  artship mirror nginx:latest \
    myregistry.com/nginx:latest \
    --dst-username user --dst-password pass`,
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

		// Prepare mirror options
		mirrorOpts := &client.MirrorOptions{
			SourceUsername: srcUsername,
			SourcePassword: srcPassword,
			SourceToken:    srcToken,
			SourceAuth:     srcAuth,
			DestUsername:   dstUsername,
			DestPassword:   dstPassword,
			DestToken:      dstToken,
			DestAuth:       dstAuth,
			Insecure:       insecure,
		}

		// Perform mirror operation
		result, err := cli.Mirror(cmd.Context(), args[0], args[1], mirrorOpts)
		if err != nil {
			return fmt.Errorf("failed to mirror image: %w", err)
		}

		// Print success message
		logger.Info("")
		logger.Info(tools.BoldGreen("✓ Image successfully mirrored!"))
		logger.Info(tools.Gray("─────────────────────────────────────────────────────────────"))
		logger.Info("Source:      %s", tools.Blue(result.SourceImage))
		logger.Info("Destination: %s", tools.Green(result.DestImage))
		logger.Info("Digest:      %s", tools.Gray(result.Digest))
		if result.Size > 0 {
			logger.Info("Size:        %s", tools.Gray(tools.FormatSize(result.Size)))
		}

		return nil
	},
}
