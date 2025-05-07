package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/junioroyewunmi/kubegpt/pkg/k8s"
	"github.com/junioroyewunmi/kubegpt/pkg/output"

	"github.com/spf13/cobra"
)

var (
	allNamespacesFlag bool
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a report of your Kubernetes cluster health",
	Long: `Generate a comprehensive report of your Kubernetes cluster health.

This command will:
1. Connect to your Kubernetes cluster
2. Collect information about all resources
3. Identify issues and potential improvements
4. Generate a report in the specified format

Examples:
  # Generate a report for the current namespace
  kubegpt report

  # Generate a report for all namespaces
  kubegpt report --all-namespaces

  # Generate a report in markdown format
  kubegpt report --output markdown --file cluster-health.md

  # Send the report to Slack
  kubegpt report --output slack --slack-webhook https://hooks.slack.com/services/...
`,
	Run: func(cmd *cobra.Command, args []string) {
		runReport()
	},
}

func init() {
	reportCmd.Flags().BoolVar(&allNamespacesFlag, "all-namespaces", false, "report on all namespaces")
}

func runReport() {
	printLogo()

	// Create Kubernetes client
	client, err := k8s.NewClient(kubeconfig)
	if err != nil {
		color.Red("Error creating Kubernetes client: %v", err)
		return
	}

	// Set namespace if provided
	if namespace != "" {
		client.SetNamespace(namespace)
	}

	// Get current namespace
	currentNamespace := client.GetCurrentNamespace()

	// Initialize diagnostic results
	results := output.DiagnosticResults{
		Namespace: currentNamespace,
		Timestamp: time.Now(),
	}

	// Generate report header
	color.New(color.FgGreen, color.Bold).Println("Kubernetes Cluster Health Report")
	color.New(color.FgWhite).Printf("Time: %s\n\n", time.Now().Format(time.RFC1123))

	fmt.Printf("Namespace: %s\n\n", currentNamespace)

	ctx := context.Background()

	// Get pods
	fmt.Println("Checking pods...")
	pods, err := client.GetUnhealthyPods(ctx)
	if err != nil {
		color.Red("Error checking pods: %v", err)
	} else {
		results.UnhealthyPods = pods
		if len(pods) > 0 {
			color.Red("Found %d unhealthy pods\n", len(pods))
			for _, pod := range pods {
				color.White("- %s: %s\n", pod.Name, pod.Status)
				if pod.Reason != "" {
					color.White("  Reason: %s\n", pod.Reason)
				}
			}
		} else {
			color.Green("All pods are healthy")
		}
	}
	fmt.Println()

	// Get deployments
	fmt.Println("Checking deployments...")
	deployments, err := client.GetMisconfiguredDeployments(ctx)
	if err != nil {
		color.Red("Error checking deployments: %v", err)
	} else {
		results.MisconfiguredDeployments = deployments
		if len(deployments) > 0 {
			color.Red("Found %d unhealthy deployments\n", len(deployments))
			for _, deployment := range deployments {
				color.White("- %s: %d/%d replicas ready\n", deployment.Name, deployment.ReadyReplicas, deployment.Replicas)
				if deployment.Reason != "" {
					color.White("  Reason: %s\n", deployment.Reason)
				}
			}
		} else {
			color.Green("All deployments are healthy")
		}
	}
	fmt.Println()

	// Get events
	fmt.Println("Checking events...")
	events, err := client.GetFailedEvents(ctx)
	if err != nil {
		color.Red("Error checking events: %v", err)
	} else {
		results.FailedEvents = events
		if len(events) > 0 {
			color.Red("Found %d failed events\n", len(events))
			for _, event := range events {
				if eventMap, ok := event.(map[string]interface{}); ok {
					reason := eventMap["reason"]
					message := eventMap["message"]
					involvedObject := eventMap["involvedObject"]

					if involvedObjectMap, ok := involvedObject.(map[string]interface{}); ok {
						kind := involvedObjectMap["kind"]
						name := involvedObjectMap["name"]

						color.White("- %s %s: %s - %s\n", kind, name, reason, message)
					}
				}
			}
		} else {
			color.Green("No failed events found")
		}
	}
	fmt.Println()

	// Save report to file if requested
	if reportFile != "" && outputFormat == "markdown" {
		// Generate the markdown report
		markdownContent := output.GenerateMarkdownReport(results)

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			color.Red("Error getting current directory: %v", err)
			return
		}

		// Create the full path to save the file
		fullPath := filepath.Join(cwd, reportFile)

		// Create the directory if it doesn't exist
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			color.Red("Error creating directory: %v", err)
			return
		}

		// Open the file with write permissions, create if it doesn't exist
		file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			color.Red("Error opening file for writing: %v", err)
			return
		}
		defer file.Close()

		// Write the content to the file
		_, err = file.WriteString(markdownContent)
		if err != nil {
			color.Red("Error writing to file: %v", err)
		} else {
			color.Green("Report saved to %s", fullPath)
			fmt.Printf("Report content length: %d bytes\n", len(markdownContent))
		}
	}

}
