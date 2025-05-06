package output

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/junioroyewunmi/kubegpt/pkg/k8s"
)

// DiagnosticResults contains the results of a diagnostic run
type DiagnosticResults struct {
	Namespace               string
	Timestamp               time.Time
	UnhealthyPods           []k8s.PodIssue
	FailedEvents            []interface{}
	MisconfiguredDeployments []k8s.DeploymentIssue
	ServiceIssues           []interface{}
}

// PrintTerminalOutput prints the diagnostic results to the terminal
func PrintTerminalOutput(results DiagnosticResults) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("Diagnostic Results for Namespace: %s\n", results.Namespace)
	color.New(color.FgCyan).Printf("Time: %s\n\n", results.Timestamp.Format("Mon, 02 Jan 2006 15:04:05 MST"))

	// Print summary
	color.New(color.FgWhite, color.Bold).Println("Summary:")
	fmt.Printf("- Unhealthy Pods: %d\n", len(results.UnhealthyPods))
	fmt.Printf("- Failed Events: %d\n", len(results.FailedEvents))
	fmt.Printf("- Misconfigured Deployments: %d\n", len(results.MisconfiguredDeployments))
	fmt.Printf("- Service Issues: %d\n", len(results.ServiceIssues))
	fmt.Println()

	// Print unhealthy pods
	if len(results.UnhealthyPods) > 0 {
		color.New(color.FgWhite, color.Bold).Println("Unhealthy Pods:")
		fmt.Println()

		for i, pod := range results.UnhealthyPods {
			color.New(color.FgYellow, color.Bold).Printf("[%d] Pod: %s\n", i+1, pod.Name)
			fmt.Printf("    Status: %s\n", pod.Status)

			if len(pod.Containers) > 0 {
				fmt.Println("    Container Issues:")
				for _, container := range pod.Containers {
					fmt.Printf("    - %s: %s (Restarts: %d)\n", container.Name, container.Status, container.Restarts)
					if container.Reason != "" {
						fmt.Printf("      Reason: %s\n", container.Reason)
					}
					if container.Message != "" {
						fmt.Printf("      Message: %s\n", container.Message)
					}
				}
			}

			if pod.Analysis != "" {
				fmt.Println()
				fmt.Println("    Analysis:")
				for _, line := range strings.Split(pod.Analysis, "\n") {
					fmt.Printf("    %s\n", line)
				}
			}

			if pod.Fix != "" {
				fmt.Println()
				fmt.Println("    Suggested Fix:")
				for _, line := range strings.Split(pod.Fix, "\n") {
					fmt.Printf("    %s\n", line)
				}
			}

			fmt.Println()
		}
	}

	// Print misconfigured deployments
	if len(results.MisconfiguredDeployments) > 0 {
		color.New(color.FgWhite, color.Bold).Println("Misconfigured Deployments:")
		fmt.Println()

		for i, deployment := range results.MisconfiguredDeployments {
			color.New(color.FgYellow, color.Bold).Printf("[%d] Deployment: %s\n", i+1, deployment.Name)
			fmt.Printf("    Replicas: %d/%d ready\n", deployment.ReadyReplicas, deployment.Replicas)
			
			if deployment.Reason != "" {
				fmt.Printf("    Reason: %s\n", deployment.Reason)
			}
			if deployment.Message != "" {
				fmt.Printf("    Message: %s\n", deployment.Message)
			}

			if deployment.Analysis != "" {
				fmt.Println()
				fmt.Println("    Analysis:")
				for _, line := range strings.Split(deployment.Analysis, "\n") {
					fmt.Printf("    %s\n", line)
				}
			}

			if deployment.Fix != "" {
				fmt.Println()
				fmt.Println("    Suggested Fix:")
				for _, line := range strings.Split(deployment.Fix, "\n") {
					fmt.Printf("    %s\n", line)
				}
			}

			fmt.Println()
		}
	}
}

// GenerateMarkdownReport generates a markdown report from the diagnostic results
func GenerateMarkdownReport(results DiagnosticResults) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Kubernetes Diagnostic Report\n\n"))
	sb.WriteString(fmt.Sprintf("**Namespace:** %s  \n", results.Namespace))
	sb.WriteString(fmt.Sprintf("**Time:** %s  \n\n", results.Timestamp.Format("Mon, 02 Jan 2006 15:04:05 MST")))

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- Unhealthy Pods: %d\n", len(results.UnhealthyPods)))
	sb.WriteString(fmt.Sprintf("- Failed Events: %d\n", len(results.FailedEvents)))
	sb.WriteString(fmt.Sprintf("- Misconfigured Deployments: %d\n", len(results.MisconfiguredDeployments)))
	sb.WriteString(fmt.Sprintf("- Service Issues: %d\n\n", len(results.ServiceIssues)))

	// Unhealthy pods
	if len(results.UnhealthyPods) > 0 {
		sb.WriteString("## Unhealthy Pods\n\n")

		for i, pod := range results.UnhealthyPods {
			sb.WriteString(fmt.Sprintf("### %d. Pod: %s\n\n", i+1, pod.Name))
			sb.WriteString(fmt.Sprintf("**Status:** %s  \n", pod.Status))

			if len(pod.Containers) > 0 {
				sb.WriteString("**Container Issues:**  \n")
				for _, container := range pod.Containers {
					sb.WriteString(fmt.Sprintf("- %s: %s (Restarts: %d)  \n", container.Name, container.Status, container.Restarts))
					if container.Reason != "" {
						sb.WriteString(fmt.Sprintf("  - Reason: %s  \n", container.Reason))
					}
					if container.Message != "" {
						sb.WriteString(fmt.Sprintf("  - Message: %s  \n", container.Message))
					}
				}
			}

			if pod.Analysis != "" {
				sb.WriteString("\n**Analysis:**  \n")
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", pod.Analysis))
			}

			if pod.Fix != "" {
				sb.WriteString("**Suggested Fix:**  \n")
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", pod.Fix))
			}
		}
	}

	// Misconfigured deployments
	if len(results.MisconfiguredDeployments) > 0 {
		sb.WriteString("## Misconfigured Deployments\n\n")

		for i, deployment := range results.MisconfiguredDeployments {
			sb.WriteString(fmt.Sprintf("### %d. Deployment: %s\n\n", i+1, deployment.Name))
			sb.WriteString(fmt.Sprintf("**Replicas:** %d/%d ready  \n", deployment.ReadyReplicas, deployment.Replicas))
			
			if deployment.Reason != "" {
				sb.WriteString(fmt.Sprintf("**Reason:** %s  \n", deployment.Reason))
			}
			if deployment.Message != "" {
				sb.WriteString(fmt.Sprintf("**Message:** %s  \n", deployment.Message))
			}

			if deployment.Analysis != "" {
				sb.WriteString("\n**Analysis:**  \n")
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", deployment.Analysis))
			}

			if deployment.Fix != "" {
				sb.WriteString("**Suggested Fix:**  \n")
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", deployment.Fix))
			}
		}
	}

	return sb.String()
}

// WriteToFile writes content to a file
func WriteToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// SendToSlack sends the report to a Slack webhook
func SendToSlack(webhookURL string, results DiagnosticResults) error {
	// Generate a simple markdown report
	report := GenerateMarkdownReport(results)
	
	// Create a simple payload
	payload := fmt.Sprintf(`{"text": "Kubernetes Diagnostic Report for namespace %s", "blocks": [{"type": "section", "text": {"type": "mrkdwn", "text": "%s"}}]}`, 
		results.Namespace, 
		strings.ReplaceAll(report, "\n", "\\n"),
	)
	
	// Send the request
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send to Slack: %s", resp.Status)
	}
	
	return nil
}