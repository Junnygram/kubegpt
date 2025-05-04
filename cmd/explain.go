package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yourusername/kubegpt/pkg/ai"
)

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:   "explain [error message or log snippet]",
	Short: "Explain Kubernetes errors or log snippets",
	Long: `Explain Kubernetes errors or log snippets using Amazon Q Developer.
Provide the error message or log snippet as an argument or pipe it from stdin.

Examples:
  # Explain a specific error message
  kubegpt explain "CrashLoopBackOff: container exited with code 1"

  # Explain a log snippet
  kubectl logs my-pod | kubegpt explain

  # Explain a YAML configuration
  kubectl get deployment my-deployment -o yaml | kubegpt explain
`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// Print logo
		printLogo()

		var input string

		// Get input from arguments or stdin
		if len(args) > 0 {
			input = strings.Join(args, " ")
		} else {
			// Check if there's input from stdin
			stdinInput, err := readStdin()
			if err != nil {
				color.Red("Error reading from stdin: %v", err)
				return
			}
			if stdinInput == "" {
				color.Red("No input provided. Please provide an error message or log snippet.")
				cmd.Help()
				return
			}
			input = stdinInput
		}

		// Detect input type
		inputType := detectInputType(input)
		color.Cyan("Detected input type: %s\n", inputType)

		// Create Amazon Q client
		amazonQ := ai.NewAmazonQClient()

		// Get explanation based on input type
		var explanation string
		var err error

		switch inputType {
		case "error":
			explanation, err = amazonQ.ExplainError(input)
		case "logs":
			explanation, err = amazonQ.ExplainLogs(input)
		case "yaml":
			explanation, err = amazonQ.ExplainYAML(input)
		default:
			explanation, err = amazonQ.ExplainGeneric(input)
		}

		if err != nil {
			color.Red("Error getting explanation: %v", err)
			return
		}

		// Output explanation
		fmt.Println("\n" + explanation)

		// Generate fix if requested
		if fix {
			color.Cyan("\nGenerating fix...")
			fixYAML, err := amazonQ.GenerateFix(input, inputType)
			if err != nil {
				color.Red("Error generating fix: %v", err)
				return
			}

			color.Green("\nSuggested fix:")
			fmt.Println(fixYAML)
		}
	},
}

// readStdin reads input from stdin
func readStdin() (string, error) {
	var input strings.Builder
	var b [1024]byte
	for {
		n, err := fmt.Stdin.Read(b[:])
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}
		if n == 0 {
			break
		}
		input.Write(b[:n])
	}
	return input.String(), nil
}

// detectInputType tries to determine the type of input
func detectInputType(input string) string {
	// Check if it's YAML
	if strings.Contains(input, "apiVersion:") && strings.Contains(input, "kind:") {
		return "yaml"
	}

	// Check if it's an error message
	errorPatterns := []string{
		"Error:", "error:", "Exception:", "exception:",
		"failed", "Failed", "CrashLoopBackOff", "OOMKilled",
		"ImagePullBackOff", "ErrImagePull",
	}
	for _, pattern := range errorPatterns {
		if strings.Contains(input, pattern) {
			return "error"
		}
	}

	// Check if it looks like logs
	if strings.Contains(input, "\n") && 
	   (strings.Contains(input, "INFO") || 
	    strings.Contains(input, "ERROR") || 
	    strings.Contains(input, "WARNING") ||
		strings.Contains(input, "DEBUG")) {
		return "logs"
	}

	// Default to generic
	return "generic"
}

func init() {
	// No specific flags for explain command
}