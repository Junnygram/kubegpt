package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/junioroyewunmi/kubegpt/pkg/ai"
	"github.com/junioroyewunmi/kubegpt/pkg/utils"
)

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:   "explain [error message or YAML]",
	Short: "Explain Kubernetes errors or YAML configurations",
	Long: `Explain Kubernetes errors or YAML configurations using Amazon Q Developer.

This command will:
1. Analyze the provided error message or YAML configuration
2. Explain what it means and potential issues
3. Suggest solutions or improvements

Examples:
  # Explain a specific error message
  kubegpt explain "CrashLoopBackOff: container exited with code 1"

  # Explain a YAML configuration from a file
  kubegpt explain -f deployment.yaml

  # Explain output from kubectl
  kubectl logs my-pod | kubegpt explain
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runExplain(args)
	},
}

var (
	explainFile string
)

func init() {
	explainCmd.Flags().StringVarP(&explainFile, "file", "f", "", "file containing error message or YAML to explain")
}

func runExplain(args []string) {
	printLogo()

	// Get the content to explain
	var content string
	var err error

	// Check if input is coming from stdin (pipe)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Input is from pipe
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			color.New(color.FgRed).Printf("Error reading from stdin: %s\n", utils.FormatError(err))
			return
		}
		content = string(bytes)
	} else if explainFile != "" {
		// Input is from file
		bytes, err := os.ReadFile(explainFile)
		if err != nil {
			color.New(color.FgRed).Printf("Error reading file: %s\n", utils.FormatError(err))
			return
		}
		content = string(bytes)
	} else if len(args) > 0 {
		// Input is from command line argument
		content = args[0]
	} else {
		color.New(color.FgRed).Println("Error: No input provided. Please provide an error message, YAML, or use --file flag.")
		return
	}

	// Create Amazon Q client
	amazonQClient := ai.NewAmazonQClient()

	// Explain the content
	color.New(color.FgCyan).Println("Analyzing with Amazon Q Developer...")
	fmt.Println()

	explanation, err := amazonQClient.ExplainError(content)
	if err != nil {
		color.New(color.FgRed).Printf("Error getting explanation: %s\n", utils.FormatError(err))
		return
	}

	// Print the explanation
	color.New(color.FgGreen, color.Bold).Println("Explanation:")
	fmt.Println()
	fmt.Println(explanation)
}