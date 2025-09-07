package command

import (
	"github.com/spf13/cobra"
)

var (
	username string
	password string
	token    string
	auth     string
	insecure bool
	verbose  bool
)

var rootCmd = &cobra.Command{
	Use:   "artship",
	Short: "Extract/examine artifacts from OCI/Docker images",
	Long:  `A CLI tool to extract/analyze artifacts from OCI/Docker images.`,
	Example: `  # Ship a binary from an image
  artship cp nginx:latest -a nginx -o /usr/local/bin
  
  # List artifacts with detailed info
  artship ls nginx:latest -d
  
  # Show file content from image
  artship cat nginx:latest nginx.conf
  
  # Show help for cp command
  artship cp --help`,
}

func Run() error {
	return rootCmd.Execute()
}
