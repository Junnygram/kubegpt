package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yourusername/kubegpt/pkg/k8s"
	"github.com/yourusername/kubegpt/pkg/ai"
	"github.com/yourusername/kubegpt/pkg/output"
)

var (
	includeEvents    bool
	includePods      bool
	includeDeployments bool
	includeServices  bool
	maxItems         int
)

// diagnoseCmd represents the diagnose command
var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose issues in your Kubernetes cluster",
	Long: `Diagnose issues in your Kubernetes cluster by analyzing unhealthy resources
and providing AI-powered explanations and suggestions for fixes.

Examples:
  # Diagnose all issues in the current namespace
  kubegpt diagnose

  # Diagnose issues in a specific namespace
  kubegpt diagnose --namespace monitoring

  # Diagnose only pod-related issues
  kubegpt diagnose --pods-only

  # Generate YAML patches to fix issues
  kubegpt diagnose --fix
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Print logo
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
		color.New(color.FgCyan).Printf("Diagnosing issues in namespace: %s\n\n", currentNamespace)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Initialize results
		var results output.DiagnosticResults
		results.Namespace = currentNamespace
		results.Timestamp = time.Now()

		// Collect unhealthy pods
		if includePods {
			fmt.Println("Checking pods...")
			pods, err := client.GetUnhealthyPods(ctx)
			if err != nil {
				color.Red("Error getting unhealthy pods: %v", err)
			} else {
				results.UnhealthyPods = pods
				color.Yellow("Found %d unhealthy pods", len(pods))
			}
		}

		// Collect failed events
		if includeEvents {
			fmt.Println("Checking events...")
			events, err := client.GetFailedEvents(ctx)
			if err != nil {
				color.Red("Error getting failed events: %v", err)
			} else {
				results.FailedEvents = events
				color.Yellow("Found %d failed events", len(events))
			}
		}

		// Collect misconfigured deployments
		if includeDeployments {
			fmt.Println("Checking deployments...")
			deployments, err := client.GetMisconfiguredDeployments(ctx)
			if err != nil {
				color.Red("Error getting misconfigured deployments: %v", err)
			} else {
				results.MisconfiguredDeployments = deployments
				color.Yellow("Found %d misconfigured deployments", len(deployments))
			}
		}

		// Collect service issues
		if includeServices {
			fmt.Println("Checking services...")
			services, err := client.GetServiceIssues(ctx)
			if err != nil {
				color.Red("Error getting service issues: %v", err)
			} else {
				results.ServiceIssues = services
				color.Yellow("Found %d service issues", len(services))
			}
		}

		// If no issues found
		if len(results.UnhealthyPods) == 0 && 
		   len(results.FailedEvents) == 0 && 
		   len(results.MisconfiguredDeployments) == 0 &&
		   len(results.ServiceIssues) == 0 {
			color.Green("\n✓ No issues found in namespace %s", currentNamespace)
			return
		}

		// Analyze issues with Amazon Q
		fmt.Println("\nAnalyzing issues with Amazon Q...")
		amazonQ := ai.NewAmazonQClient()

		// Analyze unhealthy pods
		for i, pod := range results.UnhealthyPods {
			if i >= maxItems {
				break
			}
			fmt.Printf("Analyzing pod %s...\n", pod.Name)
			analysis, err := amazonQ.AnalyzePodIssue(pod)
			if err != nil {
				color.Red("Error analyzing pod %s: %v", pod.Name, err)
				continue
			}
			results.UnhealthyPods[i].Analysis = analysis

			// Generate fix if requested
			if fix {
				fixYAML, err := amazonQ.GeneratePodFix(pod)
				if err != nil {
					color.Red("Error generating fix for pod %s: %v", pod.Name, err)
				} else {
					results.UnhealthyPods[i].Fix = fixYAML
				}
			}
		}

		// Analyze deployment issues
		for i, deployment := range results.MisconfiguredDeployments {
			if i >= maxItems {
				break
			}
			fmt.Printf("Analyzing deployment %s...\n", deployment.Name)
			analysis, err := amazonQ.AnalyzeDeploymentIssue(deployment)
			if err != nil {
				color.Red("Error analyzing deployment %s: %v", deployment.Name, err)
				continue
			}
			results.MisconfiguredDeployments[i].Analysis = analysis

			// Generate fix if requested
			if fix {
				fixYAML, err := amazonQ.GenerateDeploymentFix(deployment)
				if err != nil {
					color.Red("Error generating fix for deployment %s: %v", deployment.Name, err)
				} else {
					results.MisconfiguredDeployments[i].Fix = fixYAML
				}
			}
		}

		// Output results based on format
		switch outputFormat {
		case "terminal":
			output.PrintTerminalOutput(results)
		case "markdown":
			markdownContent := output.GenerateMarkdownReport(results)
			if reportFile != "" {
				err := output.WriteToFile(reportFile, markdownContent)
				if err != nil {
					color.Red("Error writing to file: %v", err)
				} else {
					color.Green("Report written to %s", reportFile)
				}
			} else {
				fmt.Println(markdownContent)
			}
		case "slack":
			if slackWebhook != "" {
				err := output.SendToSlack(slackWebhook, results)
				if err != nil {
					color.Red("Error sending to Slack: %v", err)
				} else {
					color.Green("Report sent to Slack")
				}
			} else {
				color.Red("Slack webhook URL not provided")
			}
		default:
			color.Red("Unknown output format: %s", outputFormat)
		}
	},
}

func init() {
	// Flags specific to the diagnose command
	diagnoseCmd.Flags().BoolVar(&includeEvents, "events", true, "include failed events in diagnosis")
	diagnoseCmd.Flags().BoolVar(&includePods, "pods", true, "include unhealthy pods in diagnosis")
	diagnoseCmd.Flags().BoolVar(&includeDeployments, "deployments", true, "include misconfigured deployments in diagnosis")
	diagnoseCmd.Flags().BoolVar(&includeServices, "services", true, "include service issues in diagnosis")
	diagnoseCmd.Flags().BoolVar(&includePods, "pods-only", false, "only check pods")
	diagnoseCmd.Flags().IntVar(&maxItems, "max-items", 5, "maximum number of items to analyze per resource type")
}