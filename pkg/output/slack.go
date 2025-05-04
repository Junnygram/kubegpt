package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackMessage represents a Slack message
type SlackMessage struct {
	Text        string        `json:"text,omitempty"`
	Blocks      []interface{} `json:"blocks,omitempty"`
	Attachments []interface{} `json:"attachments,omitempty"`
}

// SlackBlock represents a Slack block
type SlackBlock struct {
	Type string      `json:"type"`
	Text interface{} `json:"text,omitempty"`
}

// SlackTextBlock represents a Slack text block
type SlackTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SlackSectionBlock represents a Slack section block
type SlackSectionBlock struct {
	Type string      `json:"type"`
	Text interface{} `json:"text"`
}

// SlackDividerBlock represents a Slack divider block
type SlackDividerBlock struct {
	Type string `json:"type"`
}

// SlackAttachment represents a Slack attachment
type SlackAttachment struct {
	Color     string        `json:"color"`
	Blocks    []interface{} `json:"blocks"`
	Title     string        `json:"title,omitempty"`
	TitleLink string        `json:"title_link,omitempty"`
	Text      string        `json:"text,omitempty"`
	Footer    string        `json:"footer,omitempty"`
	Ts        int64         `json:"ts,omitempty"`
}

// SendToSlack sends diagnostic results to Slack
func SendToSlack(webhookURL string, results DiagnosticResults) error {
	// Create Slack message
	message := createSlackMessage(results)

	// Convert to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	// Send to Slack
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message to Slack: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message to Slack: %s", resp.Status)
	}

	return nil
}

// SendClusterReportToSlack sends a cluster report to Slack
func SendClusterReportToSlack(webhookURL string, report ClusterReport) error {
	// Create Slack message
	message := createSlackClusterReport(report)

	// Convert to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	// Send to Slack
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message to Slack: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message to Slack: %s", resp.Status)
	}

	return nil
}

// createSlackMessage creates a Slack message from diagnostic results
func createSlackMessage(results DiagnosticResults) SlackMessage {
	// Create header
	headerText := fmt.Sprintf("*Kubernetes Diagnostic Results for Namespace: %s*", results.Namespace)
	
	// Create blocks
	blocks := []interface{}{
		SlackSectionBlock{
			Type: "section",
			Text: SlackTextBlock{
				Type: "mrkdwn",
				Text: headerText,
			},
		},
		SlackSectionBlock{
			Type: "section",
			Text: SlackTextBlock{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Time:* %s", results.Timestamp.Format(time.RFC1123)),
			},
		},
		SlackDividerBlock{
			Type: "divider",
		},
	}

	// Add summary
	summaryText := "*Summary:*\n" +
		fmt.Sprintf("• Unhealthy Pods: %d\n", len(results.UnhealthyPods)) +
		fmt.Sprintf("• Failed Events: %d\n", len(results.FailedEvents)) +
		fmt.Sprintf("• Misconfigured Deployments: %d\n", len(results.MisconfiguredDeployments)) +
		fmt.Sprintf("• Service Issues: %d", len(results.ServiceIssues))
	
	blocks = append(blocks, SlackSectionBlock{
		Type: "section",
		Text: SlackTextBlock{
			Type: "mrkdwn",
			Text: summaryText,
		},
	})

	// Create attachments for issues
	var attachments []interface{}

	// Add unhealthy pods
	if len(results.UnhealthyPods) > 0 {
		podBlocks := []interface{}{
			SlackSectionBlock{
				Type: "section",
				Text: SlackTextBlock{
					Type: "mrkdwn",
					Text: "*Unhealthy Pods*",
				},
			},
		}

		for i, pod := range results.UnhealthyPods {
			if i >= 5 { // Limit to 5 pods to avoid message size limits
				podBlocks = append(podBlocks, SlackSectionBlock{
					Type: "section",
					Text: SlackTextBlock{
						Type: "mrkdwn",
						Text: fmt.Sprintf("... and %d more", len(results.UnhealthyPods)-5),
					},
				})
				break
			}

			podText := fmt.Sprintf("*%s*: %s", pod.Name, pod.Status)
			if pod.Reason != "" {
				podText += fmt.Sprintf(" (%s)", pod.Reason)
			}
			
			// Add container issues
			if len(pod.Containers) > 0 {
				podText += "\n*Container Issues:*"
				for _, container := range pod.Containers {
					podText += fmt.Sprintf("\n• %s: %s", container.Name, container.Status)
					if container.Restarts > 0 {
						podText += fmt.Sprintf(" (Restarts: %d)", container.Restarts)
					}
					if container.Reason != "" {
						podText += fmt.Sprintf(" - %s", container.Reason)
					}
				}
			}

			podBlocks = append(podBlocks, SlackSectionBlock{
				Type: "section",
				Text: SlackTextBlock{
					Type: "mrkdwn",
					Text: podText,
				},
			})
		}

		attachments = append(attachments, SlackAttachment{
			Color:  "danger",
			Blocks: podBlocks,
		})
	}

	// Add misconfigured deployments
	if len(results.MisconfiguredDeployments) > 0 {
		deploymentBlocks := []interface{}{
			SlackSectionBlock{
				Type: "section",
				Text: SlackTextBlock{
					Type: "mrkdwn",
					Text: "*Misconfigured Deployments*",
				},
			},
		}

		for i, deployment := range results.MisconfiguredDeployments {
			if i >= 5 { // Limit to 5 deployments
				deploymentBlocks = append(deploymentBlocks, SlackSectionBlock{
					Type: "section",
					Text: SlackTextBlock{
						Type: "mrkdwn",
						Text: fmt.Sprintf("... and %d more", len(results.MisconfiguredDeployments)-5),
					},
				})
				break
			}

			deploymentText := fmt.Sprintf("*%s*: %d/%d ready", 
				deployment.Name, deployment.ReadyReplicas, deployment.Replicas)
			if deployment.Reason != "" {
				deploymentText += fmt.Sprintf(" (%s)", deployment.Reason)
			}

			deploymentBlocks = append(deploymentBlocks, SlackSectionBlock{
				Type: "section",
				Text: SlackTextBlock{
					Type: "mrkdwn",
					Text: deploymentText,
				},
			})
		}

		attachments = append(attachments, SlackAttachment{
			Color:  "warning",
			Blocks: deploymentBlocks,
		})
	}

	// Create message
	message := SlackMessage{
		Text:        fmt.Sprintf("Kubernetes Diagnostic Results for Namespace: %s", results.Namespace),
		Blocks:      blocks,
		Attachments: attachments,
	}

	return message
}

// createSlackClusterReport creates a Slack message from a cluster report
func createSlackClusterReport(report ClusterReport) SlackMessage {
	// Create header
	headerText := "*Kubernetes Cluster Health Report*"
	
	// Create blocks
	blocks := []interface{}{
		SlackSectionBlock{
			Type: "section",
			Text: SlackTextBlock{
				Type: "mrkdwn",
				Text: headerText,
			},
		},
		SlackSectionBlock{
			Type: "section",
			Text: SlackTextBlock{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Time:* %s", report.Timestamp.Format(time.RFC1123)),
			},
		},
		SlackDividerBlock{
			Type: "divider",
		},
	}

	// Add cluster stats
	statsText := "*Cluster Statistics:*\n" +
		fmt.Sprintf("• Namespaces: %d\n", report.Stats.NamespaceCount) +
		fmt.Sprintf("• Pods: %d total (%d healthy, %d unhealthy)\n", 
			report.Stats.TotalPods, report.Stats.HealthyPods, report.Stats.UnhealthyPods) +
		fmt.Sprintf("• Deployments: %d total (%d healthy, %d misconfigured)\n", 
			report.Stats.TotalDeployments, report.Stats.HealthyDeployments, report.Stats.MisconfiguredDeployments) +
		fmt.Sprintf("• Failed Events: %d\n", report.Stats.FailedEventCount) +
		fmt.Sprintf("• Service Issues: %d", report.Stats.ServiceIssueCount)
	
	blocks = append(blocks, SlackSectionBlock{
		Type: "section",
		Text: SlackTextBlock{
			Type: "mrkdwn",
			Text: statsText,
		},
	})

	// Add health status
	healthStatus := "Healthy"
	healthColor := "good"
	if report.Stats.UnhealthyPods > 0 || report.Stats.MisconfiguredDeployments > 0 || 
	   report.Stats.FailedEventCount > 0 || report.Stats.ServiceIssueCount > 0 {
		healthStatus = "Unhealthy"
		healthColor = "danger"
	}
	
	blocks = append(blocks, SlackSectionBlock{
		Type: "section",
		Text: SlackTextBlock{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*Cluster Health:* %s", healthStatus),
		},
	})

	// Create attachments for namespaces with issues
	var attachments []interface{}
	
	for _, ns := range report.Namespaces {
		// Skip namespaces with no issues
		if len(ns.UnhealthyPods) == 0 && len(ns.MisconfiguredDeployments) == 0 && 
		   len(ns.FailedEvents) == 0 && len(ns.ServiceIssues) == 0 {
			continue
		}

		namespaceBlocks := []interface{}{
			SlackSectionBlock{
				Type: "section",
				Text: SlackTextBlock{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Namespace: %s*", ns.Name),
				},
			},
		}

		// Add namespace stats
		statsText := fmt.Sprintf("• Pods: %d total (%d healthy, %d unhealthy)\n", 
			len(ns.HealthyPods)+len(ns.UnhealthyPods), len(ns.HealthyPods), len(ns.UnhealthyPods)) +
			fmt.Sprintf("• Deployments: %d total (%d healthy, %d misconfigured)\n", 
				len(ns.HealthyDeployments)+len(ns.MisconfiguredDeployments), 
				len(ns.HealthyDeployments), len(ns.MisconfiguredDeployments)) +
			fmt.Sprintf("• Failed Events: %d\n", len(ns.FailedEvents)) +
			fmt.Sprintf("• Service Issues: %d", len(ns.ServiceIssues))
		
		namespaceBlocks = append(namespaceBlocks, SlackSectionBlock{
			Type: "section",
			Text: SlackTextBlock{
				Type: "mrkdwn",
				Text: statsText,
			},
		})

		// Add issues
		if len(ns.UnhealthyPods) > 0 || len(ns.MisconfiguredDeployments) > 0 || len(ns.ServiceIssues) > 0 {
			issuesText := "*Issues:*\n"
			
			// Add unhealthy pods
			if len(ns.UnhealthyPods) > 0 {
				issuesText += "*Unhealthy Pods:*\n"
				for i, pod := range ns.UnhealthyPods {
					if i >= 3 { // Limit to 3 pods
						issuesText += fmt.Sprintf("... and %d more\n", len(ns.UnhealthyPods)-3)
						break
					}
					issuesText += fmt.Sprintf("• %s: %s", pod.Name, pod.Status)
					if pod.Reason != "" {
						issuesText += fmt.Sprintf(" (%s)", pod.Reason)
					}
					issuesText += "\n"
				}
			}
			
			// Add misconfigured deployments
			if len(ns.MisconfiguredDeployments) > 0 {
				issuesText += "*Misconfigured Deployments:*\n"
				for i, deployment := range ns.MisconfiguredDeployments {
					if i >= 3 { // Limit to 3 deployments
						issuesText += fmt.Sprintf("... and %d more\n", len(ns.MisconfiguredDeployments)-3)
						break
					}
					issuesText += fmt.Sprintf("• %s: %d/%d ready", 
						deployment.Name, deployment.ReadyReplicas, deployment.Replicas)
					if deployment.Reason != "" {
						issuesText += fmt.Sprintf(" (%s)", deployment.Reason)
					}
					issuesText += "\n"
				}
			}
			
			// Add service issues
			if len(ns.ServiceIssues) > 0 {
				issuesText += "*Service Issues:*\n"
				for i, service := range ns.ServiceIssues {
					if i >= 3 { // Limit to 3 services
						issuesText += fmt.Sprintf("... and %d more\n", len(ns.ServiceIssues)-3)
						break
					}
					issuesText += fmt.Sprintf("• %s: %d endpoints", 
						service.Name, service.EndpointCount)
					if service.Reason != "" {
						issuesText += fmt.Sprintf(" (%s)", service.Reason)
					}
					issuesText += "\n"
				}
			}
			
			namespaceBlocks = append(namespaceBlocks, SlackSectionBlock{
				Type: "section",
				Text: SlackTextBlock{
					Type: "mrkdwn",
					Text: issuesText,
				},
			})
		}

		attachments = append(attachments, SlackAttachment{
			Color:  "warning",
			Blocks: namespaceBlocks,
		})
	}

	// Create message
	message := SlackMessage{
		Text:        "Kubernetes Cluster Health Report",
		Blocks:      blocks,
		Attachments: attachments,
	}

	return message
}