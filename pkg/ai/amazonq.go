package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/junioroyewunmi/kubegpt/pkg/k8s"
)

// AmazonQClient is a client for interacting with Amazon Q Developer
type AmazonQClient struct {
	cliPath string
}

// NewAmazonQClient creates a new Amazon Q client
func NewAmazonQClient() *AmazonQClient {
	// Try to find the Amazon Q CLI
	cliPath, err := exec.LookPath("amazon-q")
	if err != nil {
		// If not found, use the default path
		cliPath = "amazon-q"
	}

	return &AmazonQClient{
		cliPath: cliPath,
	}
}

// AnalyzePodIssue analyzes a pod issue using Amazon Q
func (c *AmazonQClient) AnalyzePodIssue(pod k8s.PodIssue) (string, error) {
	// Build prompt for Amazon Q
	prompt := buildPodIssuePrompt(pod)
	
	// Call Amazon Q
	return c.CallAmazonQ(prompt)
}

// AnalyzeDeploymentIssue analyzes a deployment issue using Amazon Q
func (c *AmazonQClient) AnalyzeDeploymentIssue(deployment k8s.DeploymentIssue) (string, error) {
	// Build prompt for Amazon Q
	prompt := buildDeploymentIssuePrompt(deployment)
	
	// Call Amazon Q
	return c.CallAmazonQ(prompt)
}

// ExplainError explains a Kubernetes error using Amazon Q
func (c *AmazonQClient) ExplainError(errorMsg string) (string, error) {
	// Build prompt for Amazon Q
	prompt := fmt.Sprintf(`
As a Kubernetes expert, please analyze this error message and explain:
1. What the error means
2. Likely causes
3. How to fix it
4. Specific kubectl commands that might help diagnose or fix the issue

Error message:
%s
`, errorMsg)
	
	// Call Amazon Q
	return c.CallAmazonQ(prompt)
}

// GeneratePodFix generates a fix for a pod issue using Amazon Q
func (c *AmazonQClient) GeneratePodFix(pod k8s.PodIssue) (string, error) {
	// Build prompt for Amazon Q
	prompt := fmt.Sprintf(`
As a Kubernetes expert, please generate a fix for this pod issue:

Pod: %s
Status: %s
Message: %s
Reason: %s

Container issues:
%s

Events:
%s

Please provide:
1. A brief explanation of the fix
2. YAML patch or kubectl commands to apply the fix
3. Any additional steps needed
`, 
		pod.Name,
		pod.Status,
		pod.Message,
		pod.Reason,
		formatContainerIssues(pod.Containers),
		formatEvents(pod.Events),
	)
	
	// Call Amazon Q
	return c.CallAmazonQ(prompt)
}

// GenerateDeploymentFix generates a fix for a deployment issue using Amazon Q
func (c *AmazonQClient) GenerateDeploymentFix(deployment k8s.DeploymentIssue) (string, error) {
	// Build prompt for Amazon Q
	prompt := fmt.Sprintf(`
As a Kubernetes expert, please generate a fix for this deployment issue:

Deployment: %s
Replicas: %d/%d ready
Message: %s
Reason: %s

Conditions:
%s

Events:
%s

Please provide:
1. A brief explanation of the fix
2. YAML patch or kubectl commands to apply the fix
3. Any additional steps needed
`, 
		deployment.Name,
		deployment.ReadyReplicas,
		deployment.Replicas,
		deployment.Message,
		deployment.Reason,
		formatDeploymentConditions(deployment.Conditions),
		formatEvents(deployment.Events),
	)
	
	// Call Amazon Q
	return c.CallAmazonQ(prompt)
}

// CallAmazonQ calls the Amazon Q CLI with a prompt
func (c *AmazonQClient) CallAmazonQ(prompt string) (string, error) {
	// Check if we should use the CLI or mock responses for development
	if os.Getenv("KUBEGPT_MOCK_AI") == "true" {
		return c.mockResponse(prompt), nil
	}

	// Create a temporary file for the prompt
	promptFile, err := os.CreateTemp("", "kubegpt-prompt-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(promptFile.Name())

	// Write the prompt to the file
	if _, err := promptFile.WriteString(prompt); err != nil {
		return "", fmt.Errorf("failed to write prompt to file: %w", err)
	}
	promptFile.Close()

	// Create a command to call Amazon Q CLI
	cmd := exec.Command(c.cliPath, "chat", "--prompt-file", promptFile.Name())
	
	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		// If the CLI is not installed, return a helpful message
		if stderr.String() == "" {
			return "", fmt.Errorf("Amazon Q CLI not found or not installed. Please install it using 'pip install amazon-q-developer-cli'")
		}
		return "", fmt.Errorf("failed to call Amazon Q: %w\n%s", err, stderr.String())
	}

	// Return the response
	return stdout.String(), nil
}

// mockResponse generates a mock response for development
func (c *AmazonQClient) mockResponse(prompt string) string {
	// For development and testing, return a mock response
	if strings.Contains(prompt, "CrashLoopBackOff") {
		return `
## Issue Analysis: CrashLoopBackOff

### What's happening:
The container is repeatedly crashing and Kubernetes is trying to restart it, but it keeps failing.

### Likely causes:
1. Application error inside the container
2. Missing configuration or environment variables
3. Resource constraints (memory/CPU)
4. Incorrect command or arguments

### How to fix:

1. Check the container logs:
   kubectl logs <pod-name> -c <container-name>

2. Check for resource constraints:
   kubectl describe pod <pod-name>

3. If it's a memory issue, increase the memory limit:
   resources:
     requests:
       memory: "128Mi"
     limits:
       memory: "256Mi"

4. If it's a configuration issue, check environment variables and config maps:
   kubectl describe configmap <configmap-name>
`
	} else if strings.Contains(prompt, "transform") || strings.Contains(prompt, "Terraform") || strings.Contains(prompt, "Pulumi") || strings.Contains(prompt, "CDK") {
		// Mock response for transformation
		if strings.Contains(prompt, "terraform") {
			return `
provider "kubernetes" {
  config_path = "~/.kube/config"
}

resource "kubernetes_deployment" "example" {
  metadata {
    name = "example-deployment"
    namespace = "default"
    labels = {
      app = "example"
    }
  }

  spec {
    replicas = 3

    selector {
      match_labels = {
        app = "example"
      }
    }

    template {
      metadata {
        labels = {
          app = "example"
        }
      }

      spec {
        container {
          image = "nginx:1.21.6"
          name  = "example"

          resources {
            limits = {
              cpu    = "0.5"
              memory = "512Mi"
            }
            requests = {
              cpu    = "250m"
              memory = "256Mi"
            }
          }

          liveness_probe {
            http_get {
              path = "/"
              port = 80
            }
            initial_delay_seconds = 30
            period_seconds        = 10
          }
        }
      }
    }
  }
}
`
		} else if strings.Contains(prompt, "pulumi-py") {
			return `
import pulumi
import pulumi_kubernetes as k8s

# Create a Kubernetes Deployment
app_labels = {"app": "example"}
deployment = k8s.apps.v1.Deployment(
    "example-deployment",
    metadata=k8s.meta.v1.ObjectMetaArgs(
        name="example-deployment",
        namespace="default",
        labels=app_labels,
    ),
    spec=k8s.apps.v1.DeploymentSpecArgs(
        replicas=3,
        selector=k8s.meta.v1.LabelSelectorArgs(
            match_labels=app_labels,
        ),
        template=k8s.core.v1.PodTemplateSpecArgs(
            metadata=k8s.meta.v1.ObjectMetaArgs(
                labels=app_labels,
            ),
            spec=k8s.core.v1.PodSpecArgs(
                containers=[
                    k8s.core.v1.ContainerArgs(
                        name="example",
                        image="nginx:1.21.6",
                        resources=k8s.core.v1.ResourceRequirementsArgs(
                            limits={
                                "cpu": "0.5",
                                "memory": "512Mi",
                            },
                            requests={
                                "cpu": "250m",
                                "memory": "256Mi",
                            },
                        ),
                        liveness_probe=k8s.core.v1.ProbeArgs(
                            http_get=k8s.core.v1.HTTPGetActionArgs(
                                path="/",
                                port=80,
                            ),
                            initial_delay_seconds=30,
                            period_seconds=10,
                        ),
                    )
                ],
            ),
        ),
    ),
)

pulumi.export("deployment_name", deployment.metadata.name)
`
		} else {
			return `
import * as k8s from '@pulumi/kubernetes';

// Create a Kubernetes Deployment
const appLabels = { app: 'example' };
const deployment = new k8s.apps.v1.Deployment('example-deployment', {
    metadata: {
        name: 'example-deployment',
        namespace: 'default',
        labels: appLabels,
    },
    spec: {
        replicas: 3,
        selector: {
            matchLabels: appLabels,
        },
        template: {
            metadata: {
                labels: appLabels,
            },
            spec: {
                containers: [{
                    name: 'example',
                    image: 'nginx:1.21.6',
                    resources: {
                        limits: {
                            cpu: '0.5',
                            memory: '512Mi',
                        },
                        requests: {
                            cpu: '250m',
                            memory: '256Mi',
                        },
                    },
                    livenessProbe: {
                        httpGet: {
                            path: '/',
                            port: 80,
                        },
                        initialDelaySeconds: 30,
                        periodSeconds: 10,
                    },
                }],
            },
        },
    },
});

export const deploymentName = deployment.metadata.name;
`
		}
	} else {
		return `
## Kubernetes Issue Analysis

Based on the information provided, here are my observations and recommendations:

### Potential Issues:
1. Resource constraints (CPU/memory)
2. Configuration errors
3. Network connectivity problems
4. Permission issues

### Recommended Actions:

1. Check pod logs for detailed error messages:
   kubectl logs <pod-name> -n <namespace>

2. Describe the resource for more details:
   kubectl describe <resource-type> <resource-name> -n <namespace>

3. Verify configuration:
   kubectl get <resource-type> <resource-name> -n <namespace> -o yaml
`
	}
}

// Helper functions to format data for prompts

func buildPodIssuePrompt(pod k8s.PodIssue) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("As a Kubernetes expert, please analyze this pod issue:\n\n"))
	sb.WriteString(fmt.Sprintf("Pod: %s\n", pod.Name))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", pod.Namespace))
	sb.WriteString(fmt.Sprintf("Status: %s\n", pod.Status))
	
	if pod.Message != "" {
		sb.WriteString(fmt.Sprintf("Message: %s\n", pod.Message))
	}
	
	if pod.Reason != "" {
		sb.WriteString(fmt.Sprintf("Reason: %s\n", pod.Reason))
	}
	
	// Add container issues
	if len(pod.Containers) > 0 {
		sb.WriteString("\nContainer issues:\n")
		sb.WriteString(formatContainerIssues(pod.Containers))
	}
	
	// Add events
	if len(pod.Events) > 0 {
		sb.WriteString("\nEvents:\n")
		sb.WriteString(formatEvents(pod.Events))
	}
	
	sb.WriteString("\nPlease provide:\n")
	sb.WriteString("1. A diagnosis of the issue\n")
	sb.WriteString("2. Likely root causes\n")
	sb.WriteString("3. Recommended solutions\n")
	sb.WriteString("4. Specific kubectl commands to help diagnose or fix the issue\n")
	
	return sb.String()
}

func buildDeploymentIssuePrompt(deployment k8s.DeploymentIssue) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("As a Kubernetes expert, please analyze this deployment issue:\n\n"))
	sb.WriteString(fmt.Sprintf("Deployment: %s\n", deployment.Name))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", deployment.Namespace))
	sb.WriteString(fmt.Sprintf("Replicas: %d/%d ready\n", deployment.ReadyReplicas, deployment.Replicas))
	
	if deployment.Message != "" {
		sb.WriteString(fmt.Sprintf("Message: %s\n", deployment.Message))
	}
	
	if deployment.Reason != "" {
		sb.WriteString(fmt.Sprintf("Reason: %s\n", deployment.Reason))
	}
	
	// Add conditions
	if deployment.Conditions != nil {
		sb.WriteString("\nConditions:\n")
		sb.WriteString(formatDeploymentConditions(deployment.Conditions))
	}
	
	// Add events
	if len(deployment.Events) > 0 {
		sb.WriteString("\nEvents:\n")
		sb.WriteString(formatEvents(deployment.Events))
	}
	
	sb.WriteString("\nPlease provide:\n")
	sb.WriteString("1. A diagnosis of the issue\n")
	sb.WriteString("2. Likely root causes\n")
	sb.WriteString("3. Recommended solutions\n")
	sb.WriteString("4. Specific kubectl commands to help diagnose or fix the issue\n")
	
	return sb.String()
}

func formatContainerIssues(containers []k8s.ContainerIssue) string {
	var sb strings.Builder
	
	for _, container := range containers {
		sb.WriteString(fmt.Sprintf("- Container: %s\n", container.Name))
		sb.WriteString(fmt.Sprintf("  Status: %s\n", container.Status))
		sb.WriteString(fmt.Sprintf("  Restarts: %d\n", container.Restarts))
		
		if container.Reason != "" {
			sb.WriteString(fmt.Sprintf("  Reason: %s\n", container.Reason))
		}
		
		if container.Message != "" {
			sb.WriteString(fmt.Sprintf("  Message: %s\n", container.Message))
		}
		
		sb.WriteString("\n")
	}
	
	return sb.String()
}

func formatEvents(events []interface{}) string {
	var sb strings.Builder
	
	// Convert events to JSON for consistent formatting
	eventsJSON, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return "Error formatting events"
	}
	
	sb.WriteString(string(eventsJSON))
	return sb.String()
}

func formatDeploymentConditions(conditions interface{}) string {
	var sb strings.Builder
	
	// Convert conditions to JSON for consistent formatting
	conditionsJSON, err := json.MarshalIndent(conditions, "", "  ")
	if err != nil {
		return "Error formatting conditions"
	}
	
	sb.WriteString(string(conditionsJSON))
	return sb.String()
}