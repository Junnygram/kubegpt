package ai

import (
	"bytes"
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
	prompt := fmt.Sprintf(`
As a Kubernetes expert, please analyze this pod issue:

Pod: %s
Namespace: %s
Status: %s
Message: %s
Reason: %s

Please provide:
1. A diagnosis of the issue
2. Likely root causes
3. Recommended solutions
4. Specific kubectl commands to help diagnose or fix the issue
`, pod.Name, pod.Namespace, pod.Status, pod.Message, pod.Reason)

	// Call Amazon Q
	return c.CallAmazonQ(prompt)
}

// AnalyzeDeploymentIssue analyzes a deployment issue using Amazon Q
func (c *AmazonQClient) AnalyzeDeploymentIssue(deployment k8s.DeploymentIssue) (string, error) {
	// Build prompt for Amazon Q
	prompt := fmt.Sprintf(`
As a Kubernetes expert, please analyze this deployment issue:

Deployment: %s
Namespace: %s
Replicas: %d/%d ready
Message: %s
Reason: %s

Please provide:
1. A diagnosis of the issue
2. Likely root causes
3. Recommended solutions
4. Specific kubectl commands to help diagnose or fix the issue
`, deployment.Name, deployment.Namespace, deployment.ReadyReplicas, deployment.Replicas, deployment.Message, deployment.Reason)

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

Please provide:
1. A brief explanation of the fix
2. YAML patch or kubectl commands to apply the fix
3. Any additional steps needed
`,
		pod.Name,
		pod.Status,
		pod.Message,
		pod.Reason,
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
	)

	// Call Amazon Q
	return c.CallAmazonQ(prompt)
}

// GenerateResponse generates a response based on a custom prompt
func (c *AmazonQClient) GenerateResponse(prompt string) (string, error) {
	// Call Amazon Q with the provided prompt
	return c.CallAmazonQ(prompt)
}

// CallAmazonQ calls the Amazon Q CLI with a prompt
func (c *AmazonQClient) CallAmazonQ(prompt string) (string, error) {
	// Check if mock mode is enabled - default to using real Amazon Q
	useMock := false
	if os.Getenv("KUBEGPT_MOCK_AI") == "true" {
		useMock = true
	}

	if useMock {
		// Use mock response in development mode
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
		// If the CLI is not installed or fails, return a helpful message
		return "", fmt.Errorf("Failed to run Amazon Q CLI: %w. Error output: %s", err, stderr.String())
	}

	// Return the response
	return stdout.String(), nil
}

// mockResponse generates a mock response for development
func (c *AmazonQClient) mockResponse(prompt string) string {
	// Check for common error types
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
	} else if strings.Contains(prompt, "ImagePullBackOff") {
		return `
## Issue Analysis: ImagePullBackOff

### What's happening:
Kubernetes is unable to pull the specified container image from a container registry.

### Likely causes:
1. Incorrect image name or tag
2. Image does not exist in the registry
3. Private registry requires authentication
4. Network connectivity issues between the cluster and registry
5. Rate limiting by the registry (e.g., Docker Hub)

### How to fix:

1. Verify the image name and tag are correct:
   kubectl describe pod <pod-name>

2. Check if you can manually pull the image:
   docker pull <image-name>:<tag>

3. If using a private registry, create a pull secret:
   kubectl create secret docker-registry regcred --docker-server=<registry-server> --docker-username=<username> --docker-password=<password>

4. Update the deployment to use the pull secret:
   kubectl patch deployment <deployment-name> -p '{"spec":{"template":{"spec":{"imagePullSecrets":[{"name":"regcred"}]}}}}'

5. Check network connectivity from the nodes to the registry:
   kubectl debug node/<node-name> -it --image=ubuntu
`
	} else if strings.Contains(prompt, "Readiness probe failed") {
		return `
## Issue Analysis: Readiness Probe Failure

### What's happening:
The container's readiness probe is failing, which means Kubernetes won't route traffic to this pod.

### Likely causes:
1. Application is not ready to accept traffic
2. Incorrect probe configuration (wrong port, path, or timing)
3. Application is experiencing internal errors
4. Network connectivity issues within the cluster
5. Resource constraints affecting application startup

### How to fix:

1. Check the application logs for errors:
   kubectl logs <pod-name> -c <container-name>

2. Verify the readiness probe configuration:
   kubectl describe pod <pod-name>

3. Test the endpoint manually from within the cluster:
   kubectl exec -it <another-pod> -- curl http://<pod-ip>:<port>/path

4. Adjust the probe timing if the application needs more time to start:
   kubectl patch deployment <deployment-name> -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container-name>","readinessProbe":{"initialDelaySeconds":30,"periodSeconds":10}}]}}}}'

5. Check for resource constraints:
   kubectl top pod <pod-name>
`
	} else if strings.Contains(prompt, "OOMKilled") {
		return `
## Issue Analysis: OOMKilled

### What's happening:
The container was terminated because it exceeded its memory limit.

### Likely causes:
1. Memory leak in the application
2. Memory limit set too low for the application's needs
3. Temporary spike in memory usage
4. Java applications with default heap size settings
5. Large data processing without proper memory management

### How to fix:

1. Check the memory usage pattern before the OOM:
   kubectl describe pod <pod-name>

2. Increase the memory limit:
   kubectl patch deployment <deployment-name> -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container-name>","resources":{"limits":{"memory":"512Mi"},"requests":{"memory":"256Mi"}}}]}}}}'

3. For Java applications, set JVM memory limits:
   Add environment variables: JAVA_OPTS="-Xmx256m -Xms128m"

4. Analyze the application for memory leaks:
   kubectl exec -it <pod-name> -- jmap -dump:format=b,file=/tmp/heap.bin <java-pid>
   kubectl cp <pod-name>:/tmp/heap.bin ./heap.bin

5. Consider using a memory profiler or monitoring tool:
   kubectl top pod <pod-name> --containers
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
5. Container image problems
6. Storage-related issues

### Recommended Actions:

1. Check pod logs for detailed error messages:
   kubectl logs <pod-name> -n <namespace>

2. Describe the resource for more details:
   kubectl describe <resource-type> <resource-name> -n <namespace>

3. Verify configuration:
   kubectl get <resource-type> <resource-name> -n <namespace> -o yaml

4. Check events in the namespace:
   kubectl get events -n <namespace> --sort-by='.lastTimestamp'

5. For more specific guidance, please provide additional details about the issue.
`
	}
}