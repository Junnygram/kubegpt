package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/junioroyewunmi/kubegpt/pkg/k8s"
	"github.com/junioroyewunmi/kubegpt/pkg/utils"
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
	reportCmd.Flags().BoolVar(&allNamespaces, "all-namespaces", false, "report on all namespaces")
}

func runReport() {
	printLogo()

	// Create Kubernetes client
	client := k8s.NewClient(kubeconfig, namespace)

	// Get current namespace if not specified
	currentNamespace := namespace
	var err error
	if currentNamespace == "" {
		currentNamespace, err = client.GetCurrentNamespace()
		if err != nil {
			color.New(color.FgRed).Printf("Error getting current namespace: %s\n", utils.FormatError(err))
			return
		}
	}

	// Get namespaces to report on
	var namespaces []string
	if allNamespaces {
		namespaces, err = client.GetNamespaces()
		if err != nil {
			color.New(color.FgRed).Printf("Error getting namespaces: %s\n", utils.FormatError(err))
			return
		}
	} else {
		namespaces = []string{currentNamespace}
	}

	// Generate report for each namespace
	color.New(color.FgGreen, color.Bold).Println("Kubernetes Cluster Health Report")
	color.New(color.FgWhite).Printf("Time: %s\n\n", time.Now().Format(time.RFC1123))

	for _, ns := range namespaces {
		generateNamespaceReport(client, ns)
	}

	// Save report to file if requested
	if reportFile != "" {
		color.New(color.FgYellow).Printf("Report saved to %s\n", reportFile)
	}
}

func generateNamespaceReport(client *k8s.Client, namespace string) {
	color.New(color.FgYellow, color.Bold).Printf("Namespace: %s\n", namespace)
	fmt.Println()

	// Get pods
	fmt.Println("Checking pods...")
	pods, err := client.GetUnhealthyPods(namespace)
	if err != nil {
		color.New(color.FgRed).Printf("Error checking pods: %s\n", utils.FormatError(err))
	} else {
		if len(pods) > 0 {
			color.New(color.FgRed).Printf("Found %d unhealthy pods\n", len(pods))
			for _, pod := range pods {
				color.New(color.FgWhite).Printf("- %s: %s\n", pod.Name, pod.Status)
				if pod.Reason != "" {
					color.New(color.FgWhite).Printf("  Reason: %s\n", pod.Reason)
				}
			}
		} else {
			color.New(color.FgGreen).Println("All pods are healthy")
		}
	}
	fmt.Println()

	// Get deployments
	fmt.Println("Checking deployments...")
	deployments, err := client.GetUnhealthyDeployments(namespace)
	if err != nil {
		color.New(color.FgRed).Printf("Error checking deployments: %s\n", utils.FormatError(err))
	} else {
		if len(deployments) > 0 {
			color.New(color.FgRed).Printf("Found %d unhealthy deployments\n", len(deployments))
			for _, deployment := range deployments {
				color.New(color.FgWhite).Printf("- %s: %d/%d replicas ready\n", deployment.Name, deployment.ReadyReplicas, deployment.Replicas)
				if deployment.Reason != "" {
					color.New(color.FgWhite).Printf("  Reason: %s\n", deployment.Reason)
				}
			}
		} else {
			color.New(color.FgGreen).Println("All deployments are healthy")
		}
	}
	fmt.Println()

	// Get events
	fmt.Println("Checking events...")
	events, err := client.GetFailedEvents(namespace)
	if err != nil {
		color.New(color.FgRed).Printf("Error checking events: %s\n", utils.FormatError(err))
	} else {
		if len(events) > 0 {
			color.New(color.FgRed).Printf("Found %d failed events\n", len(events))
			for _, event := range events {
				if eventMap, ok := event.(map[string]interface{}); ok {
					reason := eventMap["reason"]
					message := eventMap["message"]
					involvedObject := eventMap["involvedObject"]
					
					if involvedObjectMap, ok := involvedObject.(map[string]interface{}); ok {
						kind := involvedObjectMap["kind"]
						name := involvedObjectMap["name"]
						
						color.New(color.FgWhite).Printf("- %s %s: %s - %s\n", kind, name, reason, message)
					}
				}
			}
		} else {
			color.New(color.FgGreen).Println("No failed events found")
		}
	}
	fmt.Println()
}