package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/junioroyewunmi/kubegpt/pkg/ai"
	"github.com/junioroyewunmi/kubegpt/pkg/utils"
)

var (
	inputFormat  string
	transformOutputFormat string
	inputFile    string
	outputFile   string
	targetLang   string
)

// transformCmd represents the transform command
var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transform Kubernetes resources between different formats and languages",
	Long: `Transform Kubernetes resources between different formats and languages using Amazon Q Developer.

This command can:
- Convert between YAML and JSON formats
- Translate Kubernetes resources to different IaC languages (Terraform, Pulumi, CDK)
- Convert Helm charts to plain Kubernetes manifests
- Generate Kustomize overlays from standard manifests

Examples:
  # Convert YAML to JSON
  kubegpt transform --input-format yaml --output-format json -f deployment.yaml -o deployment.json

  # Convert Kubernetes YAML to Terraform
  kubegpt transform --target-lang terraform -f deployment.yaml -o deployment.tf

  # Convert Kubernetes YAML to AWS CDK (TypeScript)
  kubegpt transform --target-lang cdk-ts -f deployment.yaml -o cdk-stack.ts

  # Convert Kubernetes YAML to Pulumi (Python)
  kubegpt transform --target-lang pulumi-py -f deployment.yaml -o pulumi.py
`,
	Run: func(cmd *cobra.Command, args []string) {
		runTransform()
	},
}

func init() {
	rootCmd.AddCommand(transformCmd)
	transformCmd.Flags().StringVar(&inputFormat, "input-format", "yaml", "input format (yaml, json)")
	transformCmd.Flags().StringVar(&transformOutputFormat, "output-format", "", "output format (yaml, json)")
	transformCmd.Flags().StringVarP(&inputFile, "file", "f", "", "input file to transform")
	transformCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file for transformed content")
	transformCmd.Flags().StringVar(&targetLang, "target-lang", "", "target language (terraform, pulumi-py, pulumi-ts, cdk-ts, cdk-py)")
}

func runTransform() {
	printLogo()
	color.New(color.FgGreen, color.Bold).Println("🔄 KUBERNETES RESOURCE TRANSFORMER 🔄")
	fmt.Println()

	// Validate input
	if inputFile == "" {
		color.New(color.FgRed).Println("Error: Input file is required")
		os.Exit(1)
	}

	// Read input file
	content, err := os.ReadFile(inputFile)
	if err != nil {
		color.New(color.FgRed).Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Determine transformation type
	var transformedContent string
	if targetLang != "" {
		// IaC transformation
		transformedContent = transformToIaC(string(content), targetLang)
	} else if transformOutputFormat != "" {
		// Format conversion
		transformedContent = transformFormat(string(content), inputFormat, transformOutputFormat)
	} else {
		color.New(color.FgRed).Println("Error: Either --output-format or --target-lang must be specified")
		os.Exit(1)
	}

	// Output the transformed content
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(transformedContent), 0644)
		if err != nil {
			color.New(color.FgRed).Printf("Error writing output file: %v\n", err)
			os.Exit(1)
		}
		color.New(color.FgGreen).Printf("✅ Successfully transformed %s to %s\n", inputFile, outputFile)
	} else {
		fmt.Println(transformedContent)
	}
}

func transformFormat(content, inputFormat, outputFormat string) string {
	color.New(color.FgYellow).Printf("Converting from %s to %s...\n", inputFormat, outputFormat)
	fmt.Println()

	// Simple format conversion between YAML and JSON
	if (inputFormat == "yaml" && outputFormat == "json") || 
	   (inputFormat == "json" && outputFormat == "yaml") {
		result, err := utils.ConvertFormat(content, inputFormat, outputFormat)
		if err != nil {
			color.New(color.FgRed).Printf("Error converting format: %v\n", err)
			os.Exit(1)
		}
		return result
	}

	color.New(color.FgRed).Printf("Unsupported format conversion: %s to %s\n", inputFormat, outputFormat)
	os.Exit(1)
	return ""
}

func transformToIaC(content, targetLang string) string {
	color.New(color.FgYellow).Printf("Transforming Kubernetes resource to %s using Amazon Q Developer...\n", targetLang)
	fmt.Println()

	// Create Amazon Q client
	client := ai.NewAmazonQClient()

	// Build prompt for transformation
	prompt := buildTransformationPrompt(content, targetLang)

	// Call Amazon Q
	result, err := client.CallAmazonQ(prompt)
	if err != nil {
		color.New(color.FgRed).Printf("Error calling Amazon Q: %v\n", err)
		os.Exit(1)
	}

	// Extract code block from the response
	return extractCodeBlock(result, targetLang)
}

func buildTransformationPrompt(content, targetLang string) string {
	var sb strings.Builder

	sb.WriteString("As a Kubernetes and Infrastructure as Code expert, please transform the following Kubernetes resource into ")

	switch targetLang {
	case "terraform":
		sb.WriteString("Terraform HCL code. Include all necessary providers and resources.")
	case "pulumi-py":
		sb.WriteString("Pulumi Python code. Include all necessary imports and resource definitions.")
	case "pulumi-ts":
		sb.WriteString("Pulumi TypeScript code. Include all necessary imports and resource definitions.")
	case "cdk-ts":
		sb.WriteString("AWS CDK TypeScript code. Include all necessary imports and constructs.")
	case "cdk-py":
		sb.WriteString("AWS CDK Python code. Include all necessary imports and constructs.")
	default:
		sb.WriteString(fmt.Sprintf("%s code. Include all necessary imports and definitions.", targetLang))
	}

	sb.WriteString("\n\nKubernetes resource:\n```yaml\n")
	sb.WriteString(content)
	sb.WriteString("\n```\n\n")

	sb.WriteString("Please provide only the code with no explanations or markdown formatting.")

	return sb.String()
}

func extractCodeBlock(response, targetLang string) string {
	// Try to extract code block from markdown response
	if strings.Contains(response, "```") {
		parts := strings.Split(response, "```")
		if len(parts) >= 3 {
			// Find the right code block based on language
			for i := 1; i < len(parts); i += 2 {
				codeBlockLang := strings.TrimSpace(parts[i])
				if codeBlockLang == "" || 
				   strings.HasPrefix(codeBlockLang, targetLang) || 
				   (targetLang == "terraform" && strings.HasPrefix(codeBlockLang, "hcl")) ||
				   (strings.HasPrefix(targetLang, "pulumi") && strings.Contains(codeBlockLang, "python")) ||
				   (strings.HasPrefix(targetLang, "cdk") && strings.Contains(codeBlockLang, "typescript")) {
					return strings.TrimSpace(parts[i+1])
				}
			}
			// If we didn't find a language-specific block, return the first code block
			return strings.TrimSpace(parts[1])
		}
	}

	// If no code block found, return the whole response
	return response
}