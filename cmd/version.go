package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the version of the CLI
	Version = "0.1.0"
	// Commit is the git commit hash
	Commit = "development"
	// BuildDate is the date the binary was built
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of KubeGPT",
	Long:  `Print the version, commit hash, and build date of KubeGPT.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("KubeGPT v%s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Built: %s\n", BuildDate)
	},
}