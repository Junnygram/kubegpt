package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	v1 "k8s.io/api/core/v1"
	"github.com/yourusername/kubegpt/pkg/k8s"
)

// DiagnosticResults represents the results of a diagnostic run
type DiagnosticResults struct {
	Namespace                string
	Timestamp                time.Time
	UnhealthyPods            []k8s.PodIssue
	FailedEvents             []k8s.EventIssue
	MisconfiguredDeployments []k8s.DeploymentIssue
	ServiceIssues            []k8s.ServiceIssue
}

// ClusterReport represents a comprehensive cluster health report
type ClusterReport struct {
	Timestamp  time.Time
	Namespaces []NamespaceReport
	Stats      ClusterStats
}

// NamespaceReport represents a health report for a namespace
type NamespaceReport struct {
	Name                     string
	HealthyPods              []v1.Pod
	UnhealthyPods            []k8s.PodIssue
	HealthyDeployments       []interface{}
	MisconfiguredDeployments []k8s.DeploymentIssue
	FailedEvents             []k8s.EventIssue
	ServiceIssues            []k8s.ServiceIssue
}

// ClusterStats represents statistics for the entire cluster
type ClusterStats struct {
	NamespaceCount          int
	TotalPods               int
	HealthyPods             int
	UnhealthyPods           int
	TotalDeployments        int
	HealthyDeployments      int
	MisconfiguredDeployments int
	FailedEventCount        int
	ServiceIssueCount       int
}

// CalculateStats calculates statistics for the cluster report
func (r *ClusterReport) CalculateStats() {
	stats := ClusterStats{
		NamespaceCount: len(r.Namespaces),
	}

	for _, ns := range r.Namespaces {
		stats.TotalPods += len(ns.HealthyPods) + len(ns.UnhealthyPods)
		stats.HealthyPods += len(ns.HealthyPods)
		stats.UnhealthyPods += len(ns.UnhealthyPods)
		stats.TotalDeployments += len(ns.HealthyDeployments) + len(ns.MisconfiguredDeployments)
		stats.HealthyDeployments += len(ns.HealthyDeployments)
		stats.MisconfiguredDeployments += len(ns.MisconfiguredDeployments)
		stats.FailedEventCount += len(ns.FailedEvents)
		stats.ServiceIssueCount += len(ns.ServiceIssues)
	}

	r.Stats = stats
}

// PrintTerminalOutput prints diagnostic results to the terminal
func PrintTerminalOutput(results DiagnosticResults) {
	// Print header
	color.New(color.FgCyan, color.Bold).Printf("\nDiagnostic Results for Namespace: %s\n", results.Namespace)
	color.New(color.FgCyan).Printf("Time: %s\n\n", results.Timestamp.Format(time.RFC1123))

	// Print summary
	color.New(color.Bold).Println("Summary:")
	fmt.Printf("- Unhealthy Pods: %d\n", len(results.UnhealthyPods))
	fmt.Printf("- Failed Events: %d\n", len(results.FailedEvents))
	fmt.Printf("- Misconfigured Deployments: %d\n", len(results.MisconfiguredDeployments))
	fmt.Printf("- Service Issues: %d\n", len(results.ServiceIssues))
	fmt.Println()

	// Print unhealthy pods
	if len(results.UnhealthyPods) > 0 {
		color.New(color.FgYellow, color.Bold).Println("Unhealthy Pods:")
		for i, pod := range results.UnhealthyPods {
			color.New(color.FgYellow).Printf("\n[%d] Pod: %s\n", i+1, pod.Name)
			fmt.Printf("    Status: %s\n", pod.Status)
			if pod.Message != "" {
				fmt.Printf("    Message: %s\n", pod.Message)
			}
			if pod.Reason != "" {
				fmt.Printf("    Reason: %s\n", pod.Reason)
			}

			// Print container issues
			if len(pod.Containers) > 0 {
				fmt.Println("    Container Issues:")
				for _, container := range pod.Containers {
					fmt.Printf("    - %s: %s", container.Name, container.Status)
					if container.Restarts > 0 {
						fmt.Printf(" (Restarts: %d)", container.Restarts)
					}
					fmt.Println()
					if container.Reason != "" {
						fmt.Printf("      Reason: %s\n", container.Reason)
					}
					if container.Message != "" {
						fmt.Printf("      Message: %s\n", container.Message)
					}
				}
			}

			// Print analysis if available
			if pod.Analysis != "" {
				color.New(color.FgGreen, color.Bold).Println("\n    Analysis:")
				fmt.Printf("    %s\n", strings.ReplaceAll(pod.Analysis, "\n", "\n    "))
			}

			// Print fix if available
			if pod.Fix != "" {
				color.New(color.FgCyan, color.Bold).Println("\n    Suggested Fix:")
				fmt.Printf("    %s\n", strings.ReplaceAll(pod.Fix, "\n", "\n    "))
			}
		}
		fmt.Println()
	}

	// Print misconfigured deployments
	if len(results.MisconfiguredDeployments) > 0 {
		color.New(color.FgYellow, color.Bold).Println("Misconfigured Deployments:")
		for i, deployment := range results.MisconfiguredDeployments {
			color.New(color.FgYellow).Printf("\n[%d] Deployment: %s\n", i+1, deployment.Name)
			fmt.Printf("    Replicas: %d/%d ready\n", deployment.ReadyReplicas, deployment.Replicas)
			if deployment.Message != "" {
				fmt.Printf("    Message: %s\n", deployment.Message)
			}
			if deployment.Reason != "" {
				fmt.Printf("    Reason: %s\n", deployment.Reason)
			}

			// Print analysis if available
			if deployment.Analysis != "" {
				color.New(color.FgGreen, color.Bold).Println("\n    Analysis:")
				fmt.Printf("    %s\n", strings.ReplaceAll(deployment.Analysis, "\n", "\n    "))
			}

			// Print fix if available
			if deployment.Fix != "" {
				color.New(color.FgCyan, color.Bold).Println("\n    Suggested Fix:")
				fmt.Printf("    %s\n", strings.ReplaceAll(deployment.Fix, "\n", "\n    "))
			}
		}
		fmt.Println()
	}

	// Print service issues
	if len(results.ServiceIssues) > 0 {
		color.New(color.FgYellow, color.Bold).Println("Service Issues:")
		for i, service := range results.ServiceIssues {
			color.New(color.FgYellow).Printf("\n[%d] Service: %s\n", i+1, service.Name)
			fmt.Printf("    Type: %s\n", service.Type)
			fmt.Printf("    Endpoints: %d\n", service.EndpointCount)
			if service.Message != "" {
				fmt.Printf("    Message: %s\n", service.Message)
			}
			if service.Reason != "" {
				fmt.Printf("    Reason: %s\n", service.Reason)
			}

			// Print analysis if available
			if service.Analysis != "" {
				color.New(color.FgGreen, color.Bold).Println("\n    Analysis:")
				fmt.Printf("    %s\n", strings.ReplaceAll(service.Analysis, "\n", "\n    "))
			}

			// Print fix if available
			if service.Fix != "" {
				color.New(color.FgCyan, color.Bold).Println("\n    Suggested Fix:")
				fmt.Printf("    %s\n", strings.ReplaceAll(service.Fix, "\n", "\n    "))
			}
		}
		fmt.Println()
	}

	// Print failed events
	if len(results.FailedEvents) > 0 {
		color.New(color.FgYellow, color.Bold).Println("Failed Events:")
		for i, event := range results.FailedEvents {
			color.New(color.FgYellow).Printf("\n[%d] Event: %s\n", i+1, event.Name)
			fmt.Printf("    Type: %s\n", event.Type)
			fmt.Printf("    Reason: %s\n", event.Reason)
			fmt.Printf("    Message: %s\n", event.Message)
			fmt.Printf("    Count: %d\n", event.Count)
			fmt.Printf("    Object: %s/%s\n", event.InvolvedObject.Kind, event.InvolvedObject.Name)
			fmt.Printf("    Last Seen: %s\n", event.LastTimestamp.Format(time.RFC1123))

			// Print analysis if available
			if event.Analysis != "" {
				color.New(color.FgGreen, color.Bold).Println("\n    Analysis:")
				fmt.Printf("    %s\n", strings.ReplaceAll(event.Analysis, "\n", "\n    "))
			}

			// Print fix if available
			if event.Fix != "" {
				color.New(color.FgCyan, color.Bold).Println("\n    Suggested Fix:")
				fmt.Printf("    %s\n", strings.ReplaceAll(event.Fix, "\n", "\n    "))
			}
		}
		fmt.Println()
	}
}

// PrintClusterReport prints a cluster health report to the terminal
func PrintClusterReport(report ClusterReport, includeHealthy bool) {
	// Print header
	color.New(color.FgCyan, color.Bold).Println("\nCluster Health Report")
	color.New(color.FgCyan).Printf("Time: %s\n\n", report.Timestamp.Format(time.RFC1123))

	// Print cluster stats
	color.New(color.Bold).Println("Cluster Statistics:")
	fmt.Printf("- Namespaces: %d\n", report.Stats.NamespaceCount)
	fmt.Printf("- Pods: %d total (%d healthy, %d unhealthy)\n", 
		report.Stats.TotalPods, report.Stats.HealthyPods, report.Stats.UnhealthyPods)
	fmt.Printf("- Deployments: %d total (%d healthy, %d misconfigured)\n", 
		report.Stats.TotalDeployments, report.Stats.HealthyDeployments, report.Stats.MisconfiguredDeployments)
	fmt.Printf("- Failed Events: %d\n", report.Stats.FailedEventCount)
	fmt.Printf("- Service Issues: %d\n", report.Stats.ServiceIssueCount)
	fmt.Println()

	// Print health status
	healthStatus := "Healthy"
	healthColor := color.FgGreen
	if report.Stats.UnhealthyPods > 0 || report.Stats.MisconfiguredDeployments > 0 || 
	   report.Stats.FailedEventCount > 0 || report.Stats.ServiceIssueCount > 0 {
		healthStatus = "Unhealthy"
		healthColor = color.FgRed
	}
	color.New(healthColor, color.Bold).Printf("Cluster Health: %s\n\n", healthStatus)

	// Print namespace reports
	for _, ns := range report.Namespaces {
		// Skip namespaces with no issues if not including healthy
		if !includeHealthy && len(ns.UnhealthyPods) == 0 && len(ns.MisconfiguredDeployments) == 0 && 
		   len(ns.FailedEvents) == 0 && len(ns.ServiceIssues) == 0 {
			continue
		}

		color.New(color.FgCyan, color.Bold).Printf("Namespace: %s\n", ns.Name)
		
		// Print namespace stats
		fmt.Printf("- Pods: %d total (%d healthy, %d unhealthy)\n", 
			len(ns.HealthyPods)+len(ns.UnhealthyPods), len(ns.HealthyPods), len(ns.UnhealthyPods))
		fmt.Printf("- Deployments: %d total (%d healthy, %d misconfigured)\n", 
			len(ns.HealthyDeployments)+len(ns.MisconfiguredDeployments), 
			len(ns.HealthyDeployments), len(ns.MisconfiguredDeployments))
		fmt.Printf("- Failed Events: %d\n", len(ns.FailedEvents))
		fmt.Printf("- Service Issues: %d\n", len(ns.ServiceIssues))
		
		// Print unhealthy pods
		if len(ns.UnhealthyPods) > 0 {
			color.New(color.FgYellow, color.Bold).Println("\n  Unhealthy Pods:")
			for i, pod := range ns.UnhealthyPods {
				color.New(color.FgYellow).Printf("  [%d] %s: %s", i+1, pod.Name, pod.Status)
				if pod.Reason != "" {
					fmt.Printf(" (%s)", pod.Reason)
				}
				fmt.Println()
			}
		}
		
		// Print misconfigured deployments
		if len(ns.MisconfiguredDeployments) > 0 {
			color.New(color.FgYellow, color.Bold).Println("\n  Misconfigured Deployments:")
			for i, deployment := range ns.MisconfiguredDeployments {
				color.New(color.FgYellow).Printf("  [%d] %s: %d/%d ready", 
					i+1, deployment.Name, deployment.ReadyReplicas, deployment.Replicas)
				if deployment.Reason != "" {
					fmt.Printf(" (%s)", deployment.Reason)
				}
				fmt.Println()
			}
		}
		
		// Print service issues
		if len(ns.ServiceIssues) > 0 {
			color.New(color.FgYellow, color.Bold).Println("\n  Service Issues:")
			for i, service := range ns.ServiceIssues {
				color.New(color.FgYellow).Printf("  [%d] %s: %d endpoints", 
					i+1, service.Name, service.EndpointCount)
				if service.Reason != "" {
					fmt.Printf(" (%s)", service.Reason)
				}
				fmt.Println()
			}
		}
		
		fmt.Println()
	}
}

// WriteToFile writes content to a file
func WriteToFile(filename, content string) error {
	return nil // Placeholder - implement in file.go
}

// SendToSlack sends results to Slack
func SendToSlack(webhookURL string, results DiagnosticResults) error {
	return nil // Placeholder - implement in slack.go
}

// SendClusterReportToSlack sends a cluster report to Slack
func SendClusterReportToSlack(webhookURL string, report ClusterReport) error {
	return nil // Placeholder - implement in slack.go
}