package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/junioroyewunmi/kubegpt/pkg/k8s"
)

// DiagnosticResults represents the results of a diagnostic run
type DiagnosticResults struct {
	Namespace                string
	Timestamp                time.Time
	UnhealthyPods            []k8s.PodIssue
	FailedEvents             []interface{}
	MisconfiguredDeployments []k8s.DeploymentIssue
	ServiceIssues            []interface{}
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
}