package k8s

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Client represents a Kubernetes client
type Client struct {
	kubeconfig string
	namespace  string
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfig, namespace string) *Client {
	return &Client{
		kubeconfig: kubeconfig,
		namespace:  namespace,
	}
}

// GetCurrentNamespace gets the current namespace
func (c *Client) GetCurrentNamespace() (string, error) {
	if c.namespace != "" {
		return c.namespace, nil
	}

	// If no namespace is specified, get the current namespace from kubectl
	cmd := c.kubectlCommand("config", "view", "--minify", "--output", "jsonpath={..namespace}")
	output, err := cmd.Output()
	if err != nil {
		return "default", nil // Default to "default" namespace if we can't determine the current one
	}

	namespace := string(output)
	if namespace == "" {
		namespace = "default"
	}

	return namespace, nil
}

// kubectlCommand creates a kubectl command with the specified arguments
func (c *Client) kubectlCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("kubectl", args...)
	
	// Set kubeconfig if specified
	if c.kubeconfig != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", c.kubeconfig))
	}
	
	return cmd
}

// ExecuteKubectl executes a kubectl command and returns the output
func (c *Client) ExecuteKubectl(args ...string) (string, error) {
	cmd := c.kubectlCommand(args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kubectl error: %w\nOutput: %s", err, string(output))
	}
	return string(output), nil
}

// GetNamespaces gets all namespaces
func (c *Client) GetNamespaces() ([]string, error) {
	output, err := c.ExecuteKubectl("get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return nil, err
	}
	
	if output == "" {
		return []string{}, nil
	}
	
	return strings.Split(output, " "), nil
}