package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/yourusername/kubegpt/pkg/k8s"
	"github.com/yourusername/kubegpt/pkg/output"
)

var (
	allNamespaces bool
	includeHealthy bool
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a comprehensive cluster health report",
	Long: `Generate a comprehensive health report for your Kubernetes cluster.
The report includes information about unhealthy resources, failed events,
and misconfigured deployments across all or selected namespaces.

Examples:
  # Generate a report for the current namespace
  kubegpt report

  # Generate a report for all namespaces
  kubegpt report --all-namespaces

  # Generate a markdown report and save it to a file
  kubegpt report --output markdown --file cluster-health.md

  # Send the report to Slack
  kubegpt report --output slack --slack-webhook https://hooks.slack.com/services/...
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

		// Get namespaces to check
		var namespaces []string
		if allNamespaces {
			ns, err := client.GetAllNamespaces(context.Background())
			if err != nil {
				color.Red("Error getting namespaces: %v", err)
				return
			}
			namespaces = ns
			color.Cyan("Generating report for all %d namespaces...", len(namespaces))
		} else {
			// Set namespace if provided
			if namespace != "" {
				client.SetNamespace(namespace)
			}
			namespaces = []string{client.GetCurrentNamespace()}
			color.Cyan("Generating report for namespace: %s", namespaces[0])
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Initialize cluster report
		clusterReport := output.ClusterReport{
			Timestamp:  time.Now(),
			Namespaces: make([]output.NamespaceReport, 0, len(namespaces)),
		}

		// Check each namespace
		for _, ns := range namespaces {
			client.SetNamespace(ns)
			
			fmt.Printf("\nChecking namespace: %s\n", ns)
			
			// Initialize namespace report
			namespaceReport := output.NamespaceReport{
				Name: ns,
			}

			// Get pod stats
			healthyPods, unhealthyPods, err := client.GetPodStats(ctx)
			if err != nil {
				color.Red("Error getting pod stats: %v", err)
			} else {
				namespaceReport.HealthyPods = healthyPods
				namespaceReport.UnhealthyPods = unhealthyPods
				fmt.Printf("Pods: %d healthy, %d unhealthy\n", len(healthyPods), len(unhealthyPods))
			}

			// Get deployment stats
			healthyDeployments, misconfiguredDeployments, err := client.GetDeploymentStats(ctx)
			if err != nil {
				color.Red("Error getting deployment stats: %v", err)
			} else {
				namespaceReport.HealthyDeployments = healthyDeployments
				namespaceReport.MisconfiguredDeployments = misconfiguredDeployments
				fmt.Printf("Deployments: %d healthy, %d misconfigured\n", len(healthyDeployments), len(misconfiguredDeployments))
			}

			// Get failed events
			failedEvents, err := client.GetFailedEvents(ctx)
			if err != nil {
				color.Red("Error getting failed events: %v", err)
			} else {
				namespaceReport.FailedEvents = failedEvents
				fmt.Printf("Events: %d failed\n", len(failedEvents))
			}

			// Get service issues
			serviceIssues, err := client.GetServiceIssues(ctx)
			if err != nil {
				color.Red("Error getting service issues: %v", err)
			} else {
				namespaceReport.ServiceIssues = serviceIssues
				fmt.Printf("Services: %d with issues\n", len(serviceIssues))
			}

			// Add namespace report to cluster report
			clusterReport.Namespaces = append(clusterReport.Namespaces, namespaceReport)
		}

		// Calculate cluster stats
		clusterReport.CalculateStats()

		// Output report based on format
		switch outputFormat {
		case "terminal":
			output.PrintClusterReport(clusterReport, includeHealthy)
		case "markdown":
			markdownContent := output.GenerateMarkdownClusterReport(clusterReport, includeHealthy)
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
				err := output.SendClusterReportToSlack(slackWebhook, clusterReport)
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
	// Flags specific to the report command
	reportCmd.Flags().BoolVar(&allNamespaces, "all-namespaces", false, "include all namespaces in the report")
	reportCmd.Flags().BoolVar(&includeHealthy, "include-healthy", false, "include healthy resources in the report")
}