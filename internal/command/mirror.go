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
	srcInsecure bool
)

func init() {
	// Standard auth flags apply to destination registry (most common use case)
	mirrorCmd.Flags().StringVarP(&username, "username", "u", "", "Username for destination registry authentication")
	mirrorCmd.Flags().StringVarP(&password, "password", "p", "", "Password for destination registry authentication")
	mirrorCmd.Flags().StringVarP(&token, "token", "t", "", "Token for destination registry authentication")
	mirrorCmd.Flags().StringVarP(&auth, "auth", "", "", "Auth string for destination registry authentication")
	mirrorCmd.Flags().BoolVar(&insecure, "insecure", false, "Allow insecure connections to destination registry")

	// Source registry auth (only needed if different from destination or if pulling from private registry)
	mirrorCmd.Flags().StringVar(&srcUsername, "src-username", "", "Username for source registry (if different from destination)")
	mirrorCmd.Flags().StringVar(&srcPassword, "src-password", "", "Password for source registry (if different from destination)")
	mirrorCmd.Flags().StringVar(&srcToken, "src-token", "", "Token for source registry (if different from destination)")
	mirrorCmd.Flags().StringVar(&srcAuth, "src-auth", "", "Auth string for source registry (if different from destination)")
	mirrorCmd.Flags().BoolVar(&srcInsecure, "src-insecure", false, "Allow insecure connections to source registry")

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
	Example: `  # Copy from Docker Hub to a private registry (public source, authenticated destination)
  artship mirror nginx:latest myregistry.com/nginx:latest -u admin -p secret

  # Copy between registries (auth applies to destination by default)
  artship mirror docker.io/nginx:latest myregistry.com/nginx:latest -u admin -p secret

  # Rename an image tag in the same registry
  artship mirror myregistry.com/app:v1.0 myregistry.com/app:latest -u admin -p secret

  # Copy from private to private with different credentials
  artship mirror gcr.io/private/app:v1 registry.company.com/app:v1 \
    --src-username user --src-password "$(cat key.json)" \
    -u admin -p secret

  # Copy from insecure source to secure destination
  artship mirror insecure.io/app:latest registry.company.com/app:latest \
    --src-insecure -u admin -p secret

  # Copy to insecure destination registry
  artship mirror docker.io/nginx:latest myregistry.local:5000/nginx:latest \
    --dest-insecure

  # Copy with verbose output
  artship mirror alpine:3.18 myregistry.com/alpine:3.18 -u user -p pass -v`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logs.New(verbose)

		cli := client.New(&client.Options{
			Username: username,
			Password: password,
			Token:    token,
			Auth:     auth,
			Insecure: insecure, // Use destination insecure as default for client
			Logger:   logger,
		})

		// Prepare mirror options
		mirrorOpts := &client.MirrorOptions{
			SourceUsername: srcUsername,
			SourcePassword: srcPassword,
			SourceToken:    srcToken,
			SourceAuth:     srcAuth,
			SourceInsecure: srcInsecure,

			DestUsername: username,
			DestPassword: password,
			DestToken:    token,
			DestAuth:     auth,
			DestInsecure: insecure,
		}

		// Perform mirror operation
		result, err := cli.Mirror(cmd.Context(), args[0], args[1], mirrorOpts)
		if err != nil {
			return fmt.Errorf("failed to mirror image: %w", err)
		}

		// Print success message
		logger.Info("")
		logger.Info(logs.BoldGreen("✓ Image successfully mirrored!"))
		logger.Info(logs.Gray("─────────────────────────────────────────────────────────────"))
		logger.Info("Source:      %s", logs.Blue(result.SourceImage))
		logger.Info("Destination: %s", logs.Green(result.DestImage))
		logger.Info("Digest:      %s", logs.Gray(result.Digest))
		if result.Size > 0 {
			logger.Info("Size:        %s", logs.Gray(tools.FormatSize(result.Size)))
		}

		return nil
	},
}
