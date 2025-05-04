package output

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// GenerateMarkdownReport generates a markdown report from diagnostic results
func GenerateMarkdownReport(results DiagnosticResults) string {
	var sb strings.Builder

	// Add header
	sb.WriteString("# Kubernetes Diagnostic Report\n\n")
	sb.WriteString(fmt.Sprintf("**Namespace:** %s  \n", results.Namespace))
	sb.WriteString(fmt.Sprintf("**Time:** %s  \n\n", results.Timestamp.Format(time.RFC1123)))

	// Add summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Unhealthy Pods:** %d\n", len(results.UnhealthyPods)))
	sb.WriteString(fmt.Sprintf("- **Failed Events:** %d\n", len(results.FailedEvents)))
	sb.WriteString(fmt.Sprintf("- **Misconfigured Deployments:** %d\n", len(results.MisconfiguredDeployments)))
	sb.WriteString(fmt.Sprintf("- **Service Issues:** %d\n\n", len(results.ServiceIssues)))

	// Add unhealthy pods
	if len(results.UnhealthyPods) > 0 {
		sb.WriteString("## Unhealthy Pods\n\n")
		for i, pod := range results.UnhealthyPods {
			sb.WriteString(fmt.Sprintf("### %d. Pod: %s\n\n", i+1, pod.Name))
			sb.WriteString(fmt.Sprintf("**Status:** %s  \n", pod.Status))
			if pod.Message != "" {
				sb.WriteString(fmt.Sprintf("**Message:** %s  \n", pod.Message))
			}
			if pod.Reason != "" {
				sb.WriteString(fmt.Sprintf("**Reason:** %s  \n", pod.Reason))
			}

			// Add container issues
			if len(pod.Containers) > 0 {
				sb.WriteString("\n**Container Issues:**\n\n")
				for _, container := range pod.Containers {
					sb.WriteString(fmt.Sprintf("- **%s:** %s", container.Name, container.Status))
					if container.Restarts > 0 {
						sb.WriteString(fmt.Sprintf(" (Restarts: %d)", container.Restarts))
					}
					sb.WriteString("  \n")
					if container.Reason != "" {
						sb.WriteString(fmt.Sprintf("  - Reason: %s  \n", container.Reason))
					}
					if container.Message != "" {
						sb.WriteString(fmt.Sprintf("  - Message: %s  \n", container.Message))
					}
				}
			}

			// Add analysis if available
			if pod.Analysis != "" {
				sb.WriteString("\n**Analysis:**\n\n")
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", pod.Analysis))
			}

			// Add fix if available
			if pod.Fix != "" {
				sb.WriteString("**Suggested Fix:**\n\n")
				sb.WriteString(fmt.Sprintf("```yaml\n%s\n```\n\n", pod.Fix))
			}
		}
	}

	// Add misconfigured deployments
	if len(results.MisconfiguredDeployments) > 0 {
		sb.WriteString("## Misconfigured Deployments\n\n")
		for i, deployment := range results.MisconfiguredDeployments {
			sb.WriteString(fmt.Sprintf("### %d. Deployment: %s\n\n", i+1, deployment.Name))
			sb.WriteString(fmt.Sprintf("**Replicas:** %d/%d ready  \n", deployment.ReadyReplicas, deployment.Replicas))
			if deployment.Message != "" {
				sb.WriteString(fmt.Sprintf("**Message:** %s  \n", deployment.Message))
			}
			if deployment.Reason != "" {
				sb.WriteString(fmt.Sprintf("**Reason:** %s  \n", deployment.Reason))
			}

			// Add analysis if available
			if deployment.Analysis != "" {
				sb.WriteString("\n**Analysis:**\n\n")
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", deployment.Analysis))
			}

			// Add fix if available
			if deployment.Fix != "" {
				sb.WriteString("**Suggested Fix:**\n\n")
				sb.WriteString(fmt.Sprintf("```yaml\n%s\n```\n\n", deployment.Fix))
			}
		}
	}

	// Add service issues
	if len(results.ServiceIssues) > 0 {
		sb.WriteString("## Service Issues\n\n")
		for i, service := range results.ServiceIssues {
			sb.WriteString(fmt.Sprintf("### %d. Service: %s\n\n", i+1, service.Name))
			sb.WriteString(fmt.Sprintf("**Type:** %s  \n", service.Type))
			sb.WriteString(fmt.Sprintf("**Endpoints:** %d  \n", service.EndpointCount))
			if service.Message != "" {
				sb.WriteString(fmt.Sprintf("**Message:** %s  \n", service.Message))
			}
			if service.Reason != "" {
				sb.WriteString(fmt.Sprintf("**Reason:** %s  \n", service.Reason))
			}

			// Add analysis if available
			if service.Analysis != "" {
				sb.WriteString("\n**Analysis:**\n\n")
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", service.Analysis))
			}

			// Add fix if available
			if service.Fix != "" {
				sb.WriteString("**Suggested Fix:**\n\n")
				sb.WriteString(fmt.Sprintf("```yaml\n%s\n```\n\n", service.Fix))
			}
		}
	}

	// Add failed events
	if len(results.FailedEvents) > 0 {
		sb.WriteString("## Failed Events\n\n")
		for i, event := range results.FailedEvents {
			sb.WriteString(fmt.Sprintf("### %d. Event: %s\n\n", i+1, event.Name))
			sb.WriteString(fmt.Sprintf("**Type:** %s  \n", event.Type))
			sb.WriteString(fmt.Sprintf("**Reason:** %s  \n", event.Reason))
			sb.WriteString(fmt.Sprintf("**Message:** %s  \n", event.Message))
			sb.WriteString(fmt.Sprintf("**Count:** %d  \n", event.Count))
			sb.WriteString(fmt.Sprintf("**Object:** %s/%s  \n", event.InvolvedObject.Kind, event.InvolvedObject.Name))
			sb.WriteString(fmt.Sprintf("**Last Seen:** %s  \n", event.LastTimestamp.Format(time.RFC1123)))

			// Add analysis if available
			if event.Analysis != "" {
				sb.WriteString("\n**Analysis:**\n\n")
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", event.Analysis))
			}

			// Add fix if available
			if event.Fix != "" {
				sb.WriteString("**Suggested Fix:**\n\n")
				sb.WriteString(fmt.Sprintf("```yaml\n%s\n```\n\n", event.Fix))
			}
		}
	}

	return sb.String()
}

// GenerateMarkdownClusterReport generates a markdown report for the entire cluster
func GenerateMarkdownClusterReport(report ClusterReport, includeHealthy bool) string {
	var sb strings.Builder

	// Add header
	sb.WriteString("# Kubernetes Cluster Health Report\n\n")
	sb.WriteString(fmt.Sprintf("**Time:** %s  \n\n", report.Timestamp.Format(time.RFC1123)))

	// Add cluster stats
	sb.WriteString("## Cluster Statistics\n\n")
	sb.WriteString(fmt.Sprintf("- **Namespaces:** %d\n", report.Stats.NamespaceCount))
	sb.WriteString(fmt.Sprintf("- **Pods:** %d total (%d healthy, %d unhealthy)\n", 
		report.Stats.TotalPods, report.Stats.HealthyPods, report.Stats.UnhealthyPods))
	sb.WriteString(fmt.Sprintf("- **Deployments:** %d total (%d healthy, %d misconfigured)\n", 
		report.Stats.TotalDeployments, report.Stats.HealthyDeployments, report.Stats.MisconfiguredDeployments))
	sb.WriteString(fmt.Sprintf("- **Failed Events:** %d\n", report.Stats.FailedEventCount))
	sb.WriteString(fmt.Sprintf("- **Service Issues:** %d\n\n", report.Stats.ServiceIssueCount))

	// Add health status
	healthStatus := "Healthy"
	if report.Stats.UnhealthyPods > 0 || report.Stats.MisconfiguredDeployments > 0 || 
	   report.Stats.FailedEventCount > 0 || report.Stats.ServiceIssueCount > 0 {
		healthStatus = "Unhealthy"
	}
	sb.WriteString(fmt.Sprintf("**Cluster Health:** %s\n\n", healthStatus))

	// Add namespace reports
	sb.WriteString("## Namespace Reports\n\n")
	for _, ns := range report.Namespaces {
		// Skip namespaces with no issues if not including healthy
		if !includeHealthy && len(ns.UnhealthyPods) == 0 && len(ns.MisconfiguredDeployments) == 0 && 
		   len(ns.FailedEvents) == 0 && len(ns.ServiceIssues) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### Namespace: %s\n\n", ns.Name))
		
		// Add namespace stats
		sb.WriteString(fmt.Sprintf("- **Pods:** %d total (%d healthy, %d unhealthy)\n", 
			len(ns.HealthyPods)+len(ns.UnhealthyPods), len(ns.HealthyPods), len(ns.UnhealthyPods)))
		sb.WriteString(fmt.Sprintf("- **Deployments:** %d total (%d healthy, %d misconfigured)\n", 
			len(ns.HealthyDeployments)+len(ns.MisconfiguredDeployments), 
			len(ns.HealthyDeployments), len(ns.MisconfiguredDeployments)))
		sb.WriteString(fmt.Sprintf("- **Failed Events:** %d\n", len(ns.FailedEvents)))
		sb.WriteString(fmt.Sprintf("- **Service Issues:** %d\n\n", len(ns.ServiceIssues)))
		
		// Add unhealthy pods
		if len(ns.UnhealthyPods) > 0 {
			sb.WriteString("#### Unhealthy Pods\n\n")
			sb.WriteString("| Name | Status | Reason | Message |\n")
			sb.WriteString("|------|--------|--------|--------|\n")
			for _, pod := range ns.UnhealthyPods {
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", 
					pod.Name, pod.Status, pod.Reason, pod.Message))
			}
			sb.WriteString("\n")
		}
		
		// Add misconfigured deployments
		if len(ns.MisconfiguredDeployments) > 0 {
			sb.WriteString("#### Misconfigured Deployments\n\n")
			sb.WriteString("| Name | Replicas | Reason | Message |\n")
			sb.WriteString("|------|----------|--------|--------|\n")
			for _, deployment := range ns.MisconfiguredDeployments {
				sb.WriteString(fmt.Sprintf("| %s | %d/%d | %s | %s |\n", 
					deployment.Name, deployment.ReadyReplicas, deployment.Replicas, 
					deployment.Reason, deployment.Message))
			}
			sb.WriteString("\n")
		}
		
		// Add service issues
		if len(ns.ServiceIssues) > 0 {
			sb.WriteString("#### Service Issues\n\n")
			sb.WriteString("| Name | Type | Endpoints | Reason |\n")
			sb.WriteString("|------|------|-----------|--------|\n")
			for _, service := range ns.ServiceIssues {
				sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n", 
					service.Name, service.Type, service.EndpointCount, service.Reason))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// WriteToFile writes content to a file
func WriteToFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}