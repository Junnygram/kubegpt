package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/yourusername/kubegpt/pkg/k8s"
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
	return c.callAmazonQ(prompt)
}

// AnalyzeDeploymentIssue analyzes a deployment issue using Amazon Q
func (c *AmazonQClient) AnalyzeDeploymentIssue(deployment k8s.DeploymentIssue) (string, error) {
	// Build prompt for Amazon Q
	prompt := buildDeploymentIssuePrompt(deployment)
	
	// Call Amazon Q
	return c.callAmazonQ(prompt)
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
	return c.callAmazonQ(prompt)
}

// ExplainLogs explains Kubernetes logs using Amazon Q
func (c *AmazonQClient) ExplainLogs(logs string) (string, error) {
	// Build prompt for Amazon Q
	prompt := fmt.Sprintf(`
As a Kubernetes expert, please analyze these logs and explain:
1. What issues or errors are present
2. Root causes of any problems
3. How to fix the issues
4. Specific kubectl commands that might help diagnose or fix the issues

Logs:
%s
`, logs)
	
	// Call Amazon Q
	return c.callAmazonQ(prompt)
}

// ExplainYAML explains a Kubernetes YAML configuration using Amazon Q
func (c *AmazonQClient) ExplainYAML(yaml string) (string, error) {
	// Build prompt for Amazon Q
	prompt := fmt.Sprintf(`
As a Kubernetes expert, please analyze this YAML configuration and explain:
1. What this configuration does
2. Any issues, misconfigurations, or security concerns
3. Best practices that should be applied
4. Specific improvements that could be made

YAML:
%s
`, yaml)
	
	// Call Amazon Q
	return c.callAmazonQ(prompt)
}

// ExplainGeneric provides a generic explanation using Amazon Q
func (c *AmazonQClient) ExplainGeneric(input string) (string, error) {
	// Build prompt for Amazon Q
	prompt := fmt.Sprintf(`
As a Kubernetes expert, please analyze this information and provide insights:
1. What is this information about
2. Any issues or concerns
3. Recommendations or best practices
4. Specific kubectl commands that might be helpful

Input:
%s
`, input)
	
	// Call Amazon Q
	return c.callAmazonQ(prompt)
}

// GenerateFix generates a fix for a Kubernetes issue using Amazon Q
func (c *AmazonQClient) GenerateFix(input, inputType string) (string, error) {
	// Build prompt for Amazon Q based on input type
	var prompt string
	
	switch inputType {
	case "error":
		prompt = fmt.Sprintf(`
As a Kubernetes expert, please generate a YAML patch or kubectl commands to fix this error:

Error:
%s

Please provide:
1. A brief explanation of the fix
2. YAML patch or kubectl commands to apply the fix
3. Any additional steps needed
`, input)
	case "yaml":
		prompt = fmt.Sprintf(`
As a Kubernetes expert, please generate an improved version of this YAML configuration:

Original YAML:
%s

Please provide:
1. A brief explanation of the improvements
2. The complete improved YAML
3. Any additional steps needed
`, input)
	default:
		prompt = fmt.Sprintf(`
As a Kubernetes expert, please generate a fix for this issue:

Issue:
%s

Please provide:
1. A brief explanation of the fix
2. YAML patch or kubectl commands to apply the fix
3. Any additional steps needed
`, input)
	}
	
	// Call Amazon Q
	return c.callAmazonQ(prompt)
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
	return c.callAmazonQ(prompt)
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
	return c.callAmazonQ(prompt)
}

// callAmazonQ calls the Amazon Q CLI with a prompt
func (c *AmazonQClient) callAmazonQ(prompt string) (string, error) {
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
   \`\`\`
   kubectl logs <pod-name> -c <container-name>
   \`\`\`

2. Check for resource constraints:
   \`\`\`
   kubectl describe pod <pod-name>
   \`\`\`

3. If it's a memory issue, increase the memory limit:
   \`\`\`yaml
   resources:
     requests:
       memory: "128Mi"
     limits:
       memory: "256Mi"
   \`\`\`

4. If it's a configuration issue, check environment variables and config maps:
   \`\`\`
   kubectl describe configmap <configmap-name>
   \`\`\`

5. Try running the container locally to debug:
   \`\`\`
   docker run --rm -it <image-name> <command>
   \`\`\`
`
	} else if strings.Contains(prompt, "ImagePullBackOff") {
		return `
## Issue Analysis: ImagePullBackOff

### What's happening:
Kubernetes is unable to pull the container image from the registry.

### Likely causes:
1. Image doesn't exist or wrong image name
2. Registry authentication issues
3. Network connectivity problems
4. Rate limiting by the registry

### How to fix:

1. Verify the image name and tag:
   \`\`\`
   kubectl describe pod <pod-name>
   \`\`\`

2. Check if you need registry credentials:
   \`\`\`
   kubectl create secret docker-registry regcred --docker-server=<registry> --docker-username=<username> --docker-password=<password>
   \`\`\`

3. Update the pod to use the credentials:
   \`\`\`yaml
   spec:
     imagePullSecrets:
     - name: regcred
   \`\`\`

4. Check if you can pull the image manually:
   \`\`\`
   docker pull <image-name>:<tag>
   \`\`\`

5. If using a private registry, ensure your cluster has network access to it.
`
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
   \`\`\`
   kubectl logs <pod-name> -n <namespace>
   \`\`\`

2. Describe the resource for more details:
   \`\`\`
   kubectl describe <resource-type> <resource-name> -n <namespace>
   \`\`\`

3. Verify configuration:
   \`\`\`
   kubectl get <resource-type> <resource-name> -n <namespace> -o yaml
   \`\`\`

4. Check events in the namespace:
   \`\`\`
   kubectl get events -n <namespace> --sort-by='.lastTimestamp'
   \`\`\`

5. Ensure proper RBAC permissions are in place if applicable.

For more specific guidance, please provide additional details about the issue.
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
	sb.WriteString(fmt.Sprintf("Age: %s\n", pod.Age.String()))
	sb.WriteString(fmt.Sprintf("Node: %s\n", pod.Node))
	
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
	
	// Add logs (limited to avoid huge prompts)
	if len(pod.Logs) > 0 {
		sb.WriteString("\nContainer logs (most recent):\n")
		for container, logs := range pod.Logs {
			// Limit logs to last 20 lines
			logLines := strings.Split(logs, "\n")
			if len(logLines) > 20 {
				logLines = logLines[len(logLines)-20:]
			}
			limitedLogs := strings.Join(logLines, "\n")
			
			sb.WriteString(fmt.Sprintf("\n--- %s logs ---\n%s\n", container, limitedLogs))
		}
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
	sb.WriteString(fmt.Sprintf("Updated replicas: %d/%d\n", deployment.UpdatedReplicas, deployment.Replicas))
	sb.WriteString(fmt.Sprintf("Available replicas: %d/%d\n", deployment.AvailableReplicas, deployment.Replicas))
	sb.WriteString(fmt.Sprintf("Strategy: %s\n", deployment.Strategy))
	sb.WriteString(fmt.Sprintf("Age: %s\n", deployment.Age.String()))
	
	if deployment.Message != "" {
		sb.WriteString(fmt.Sprintf("Message: %s\n", deployment.Message))
	}
	
	if deployment.Reason != "" {
		sb.WriteString(fmt.Sprintf("Reason: %s\n", deployment.Reason))
	}
	
	// Add conditions
	if len(deployment.Conditions) > 0 {
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
		sb.WriteString(fmt.Sprintf("  Image: %s\n", container.Image))
		sb.WriteString(fmt.Sprintf("  Ready: %t\n", container.Ready))
		sb.WriteString(fmt.Sprintf("  Status: %s\n", container.Status))
		sb.WriteString(fmt.Sprintf("  Restarts: %d\n", container.Restarts))
		
		if container.Reason != "" {
			sb.WriteString(fmt.Sprintf("  Reason: %s\n", container.Reason))
		}
		
		if container.Message != "" {
			sb.WriteString(fmt.Sprintf("  Message: %s\n", container.Message))
		}
		
		if container.ExitCode != 0 {
			sb.WriteString(fmt.Sprintf("  Exit Code: %d\n", container.ExitCode))
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