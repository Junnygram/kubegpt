package cmd

import (
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// Version is the current version of KubeGPT
	Version = "0.1.0"
	
	// BuildDate is the date when the binary was built
	BuildDate = "unknown"
	
	// GitCommit is the git commit hash
	GitCommit = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version, build date, git commit, and runtime information of KubeGPT.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Print logo
		printLogo()
		
		color.New(color.FgCyan, color.Bold).Println("KubeGPT Version Information")
		color.New(color.FgCyan).Println("---------------------------")
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	// No specific flags for version command
}