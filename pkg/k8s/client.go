package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Client represents a Kubernetes client
type Client struct {
	kubeconfig string
	namespace  string
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfig string) (*Client, error) {
	return &Client{
		kubeconfig: kubeconfig,
		namespace:  "",
	}, nil
}

// SetNamespace sets the namespace for the client
func (c *Client) SetNamespace(namespace string) {
	c.namespace = namespace
}

// GetCurrentNamespace gets the current namespace
func (c *Client) GetCurrentNamespace() string {
	if c.namespace != "" {
		return c.namespace
	}

	// If no namespace is specified, get the current namespace from kubectl
	cmd := c.kubectlCommand("config", "view", "--minify", "--output", "jsonpath={..namespace}")
	output, err := cmd.Output()
	if err != nil {
		// Try another method to get the current context's namespace
		contextCmd := c.kubectlCommand("config", "current-context")
		contextOutput, contextErr := contextCmd.Output()
		if contextErr == nil {
			currentContext := strings.TrimSpace(string(contextOutput))
			nsCmd := c.kubectlCommand("config", "get-contexts", currentContext, "--no-headers", "-o", "name")
			nsOutput, nsErr := nsCmd.Output()
			if nsErr == nil && string(nsOutput) != "" {
				parts := strings.Split(string(nsOutput), "/")
				if len(parts) > 2 {
					return parts[2]
				}
			}
		}
		// If all attempts fail, return an error message as namespace
		return "namespace-not-found"
	}

	namespace := string(output)
	if namespace == "" {
		// Try to get namespace from current context
		contextCmd := c.kubectlCommand("config", "current-context")
		contextOutput, contextErr := contextCmd.Output()
		if contextErr == nil {
			currentContext := strings.TrimSpace(string(contextOutput))
			if currentContext != "" {
				return currentContext + "-ns"
			}
		}
		return "current-context-ns"
	}

	return namespace
}

// NamespaceExists checks if a namespace exists
func (c *Client) NamespaceExists(namespace string) bool {
	// First try the standard way
	cmd := c.kubectlCommand("get", "namespace", namespace, "--no-headers", "--output", "name")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		return true
	}
	
	// If that fails, try listing all namespaces and check if our namespace is in the list
	allNamespacesCmd := c.kubectlCommand("get", "namespaces", "--no-headers", "-o", "custom-columns=:metadata.name")
	allOutput, allErr := allNamespacesCmd.Output()
	if allErr != nil {
		return false
	}
	
	namespaceList := strings.Split(string(allOutput), "\n")
	for _, ns := range namespaceList {
		if strings.TrimSpace(ns) == namespace {
			return true
		}
	}
	
	return false
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
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}
	
	if output == "" {
		return nil, fmt.Errorf("no namespaces found in the cluster")
	}
	
	return strings.Split(output, " "), nil
}

// GetUnhealthyPods returns a list of unhealthy pods
func (c *Client) GetUnhealthyPods(ctx context.Context) ([]PodIssue, error) {
	// Check if namespace exists
	if !c.NamespaceExists(c.GetCurrentNamespace()) {
		return nil, fmt.Errorf("namespace %q not found", c.GetCurrentNamespace())
	}

	// Get real unhealthy pods from the cluster
	pods, err := c.getRealUnhealthyPods()
	if err != nil {
		return nil, fmt.Errorf("failed to get unhealthy pods: %w", err)
	}
	
	// Return empty slice if no unhealthy pods found
	if len(pods) == 0 {
		return []PodIssue{}, nil
	}
	
	return pods, nil
}

// getRealUnhealthyPods attempts to get real unhealthy pods from the cluster
func (c *Client) getRealUnhealthyPods() ([]PodIssue, error) {
	output, err := c.ExecuteKubectl("get", "pods", "-n", c.GetCurrentNamespace(), "-o", "json")
	if err != nil {
		return nil, err
	}

	var podList struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Status struct {
				Phase             string `json:"phase"`
				ContainerStatuses []struct {
					Name         string `json:"name"`
					Ready        bool   `json:"ready"`
					RestartCount int    `json:"restartCount"`
					State        struct {
						Waiting struct {
							Reason  string `json:"reason"`
							Message string `json:"message"`
						} `json:"waiting"`
					} `json:"state"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &podList); err != nil {
		return nil, err
	}

	var unhealthyPods []PodIssue
	for _, pod := range podList.Items {
		// Check if pod is unhealthy
		isUnhealthy := false
		if pod.Status.Phase != "Running" && pod.Status.Phase != "Succeeded" {
			isUnhealthy = true
		} else {
			for _, container := range pod.Status.ContainerStatuses {
				if !container.Ready || container.RestartCount > 5 {
					isUnhealthy = true
					break
				}
			}
		}

		if isUnhealthy {
			podIssue := PodIssue{
				Name:      pod.Metadata.Name,
				Namespace: pod.Metadata.Namespace,
				Status:    pod.Status.Phase,
			}

			// Add container issues
			for _, container := range pod.Status.ContainerStatuses {
				containerIssue := ContainerIssue{
					Name:     container.Name,
					Ready:    container.Ready,
					Restarts: container.RestartCount,
				}

				if !container.Ready {
					containerIssue.Status = "Not Ready"
					if container.State.Waiting.Reason != "" {
						containerIssue.Reason = container.State.Waiting.Reason
						containerIssue.Message = container.State.Waiting.Message
					}
				} else {
					containerIssue.Status = "Running"
				}

				podIssue.Containers = append(podIssue.Containers, containerIssue)
			}

			unhealthyPods = append(unhealthyPods, podIssue)
		}
	}

	return unhealthyPods, nil
}

// GetFailedEvents returns a list of failed events
func (c *Client) GetFailedEvents(ctx context.Context) ([]interface{}, error) {
	// Check if namespace exists
	if !c.NamespaceExists(c.GetCurrentNamespace()) {
		return nil, fmt.Errorf("namespace %q not found", c.GetCurrentNamespace())
	}

	// Get real failed events from the cluster
	events, err := c.getRealFailedEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	
	// Return empty slice if no failed events found
	if len(events) == 0 {
		return []interface{}{}, nil
	}
	
	return events, nil
}

// getRealFailedEvents attempts to get real failed events from the cluster
func (c *Client) getRealFailedEvents() ([]interface{}, error) {
	// First try to get warning events in the current namespace
	output, err := c.ExecuteKubectl("get", "events", "-n", c.GetCurrentNamespace(), "--field-selector=type=Warning", "-o", "json")
	if err != nil {
		// Try getting all events if specific selector fails
		output, err = c.ExecuteKubectl("get", "events", "-n", c.GetCurrentNamespace(), "-o", "json")
		if err != nil {
			return nil, err
		}
	}

	var eventList struct {
		Items []map[string]interface{} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &eventList); err != nil {
		return nil, err
	}

	// Filter for warning and error events if we got all events
	var filteredEvents []map[string]interface{}
	for _, event := range eventList.Items {
		eventType, ok := event["type"].(string)
		if !ok {
			continue
		}
		
		// Keep Warning events and any events with "Error" or "Failed" in the reason
		if eventType == "Warning" {
			filteredEvents = append(filteredEvents, event)
		} else {
			reason, ok := event["reason"].(string)
			if ok && (strings.Contains(reason, "Error") || strings.Contains(reason, "Failed")) {
				filteredEvents = append(filteredEvents, event)
			}
		}
	}
	
	// Enrich events with additional context
	for i := range filteredEvents {
		// Add namespace and object fields for easier reference
		if involvedObj, ok := filteredEvents[i]["involvedObject"].(map[string]interface{}); ok {
			kind := involvedObj["kind"]
			name := involvedObj["name"]
			namespace, hasNs := involvedObj["namespace"]
			
			// Add simplified involvedObject for compatibility
			filteredEvents[i]["involvedObject"] = map[string]interface{}{
				"kind": kind,
				"name": name,
			}
			
			// Add object field for easier reference
			if kind != nil && name != nil {
				filteredEvents[i]["object"] = fmt.Sprintf("%s/%s", kind, name)
			}
			
			// Add namespace field for easier reference
			if hasNs {
				filteredEvents[i]["namespace"] = namespace
			} else if ns, ok := filteredEvents[i]["metadata"].(map[string]interface{})["namespace"]; ok {
				filteredEvents[i]["namespace"] = ns
			}
		}
		
		// Format age for easier reading
		if timestamp, ok := filteredEvents[i]["lastTimestamp"].(string); ok {
			// Try to parse the timestamp
			if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
				age := time.Since(t)
				if age < time.Minute {
					filteredEvents[i]["age"] = fmt.Sprintf("%ds", int(age.Seconds()))
				} else if age < time.Hour {
					filteredEvents[i]["age"] = fmt.Sprintf("%dm", int(age.Minutes()))
				} else if age < 24*time.Hour {
					filteredEvents[i]["age"] = fmt.Sprintf("%dh", int(age.Hours()))
				} else {
					filteredEvents[i]["age"] = fmt.Sprintf("%dd", int(age.Hours()/24))
				}
			}
		}
	}

	// Convert []map[string]interface{} to []interface{}
	events := make([]interface{}, len(filteredEvents))
	for i, item := range filteredEvents {
		events[i] = item
	}
	
	return events, nil
}

// GetMisconfiguredDeployments returns a list of misconfigured deployments
func (c *Client) GetMisconfiguredDeployments(ctx context.Context) ([]DeploymentIssue, error) {
	// Check if namespace exists
	if !c.NamespaceExists(c.GetCurrentNamespace()) {
		return nil, fmt.Errorf("namespace %q not found", c.GetCurrentNamespace())
	}

	// Get real misconfigured deployments from the cluster
	deployments, err := c.getRealMisconfiguredDeployments()
	if err != nil {
		return nil, fmt.Errorf("failed to get misconfigured deployments: %w", err)
	}
	
	// Return empty slice if no misconfigured deployments found
	if len(deployments) == 0 {
		return []DeploymentIssue{}, nil
	}
	
	return deployments, nil
}

// getRealMisconfiguredDeployments attempts to get real misconfigured deployments from the cluster
func (c *Client) getRealMisconfiguredDeployments() ([]DeploymentIssue, error) {
	output, err := c.ExecuteKubectl("get", "deployments", "-n", c.GetCurrentNamespace(), "-o", "json")
	if err != nil {
		return nil, err
	}

	var deploymentList struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
				Labels    map[string]string `json:"labels"`
			} `json:"metadata"`
			Spec struct {
				Replicas int `json:"replicas"`
				Strategy struct {
					Type string `json:"type"`
				} `json:"strategy"`
				Selector struct {
					MatchLabels map[string]string `json:"matchLabels"`
				} `json:"selector"`
			} `json:"spec"`
			Status struct {
				Replicas            int `json:"replicas"`
				ReadyReplicas       int `json:"readyReplicas"`
				UpdatedReplicas     int `json:"updatedReplicas"`
				AvailableReplicas   int `json:"availableReplicas"`
				UnavailableReplicas int `json:"unavailableReplicas"`
				Conditions          []struct {
					Type    string `json:"type"`
					Status  string `json:"status"`
					Reason  string `json:"reason"`
					Message string `json:"message"`
				} `json:"conditions"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &deploymentList); err != nil {
		return nil, err
	}

	var misconfiguredDeployments []DeploymentIssue
	for _, deployment := range deploymentList.Items {
		// Check if deployment is misconfigured
		isMisconfigured := false
		
		// Check for various misconfiguration conditions
		if deployment.Status.ReadyReplicas < deployment.Spec.Replicas {
			isMisconfigured = true
		}
		
		// Check for stalled rollouts
		if deployment.Status.UpdatedReplicas < deployment.Spec.Replicas {
			isMisconfigured = true
		}
		
		// Check for non-true conditions
		hasNonTrueCondition := false
		for _, condition := range deployment.Status.Conditions {
			if condition.Status != "True" {
				hasNonTrueCondition = true
				break
			}
		}
		
		if isMisconfigured || hasNonTrueCondition {
			deploymentIssue := DeploymentIssue{
				Name:              deployment.Metadata.Name,
				Namespace:         deployment.Metadata.Namespace,
				Replicas:          deployment.Spec.Replicas,
				ReadyReplicas:     deployment.Status.ReadyReplicas,
				UpdatedReplicas:   deployment.Status.UpdatedReplicas,
				AvailableReplicas: deployment.Status.AvailableReplicas,
				Strategy:          deployment.Spec.Strategy.Type,
			}

			// Add conditions
			var conditions []map[string]interface{}
			for _, condition := range deployment.Status.Conditions {
				if condition.Status != "True" {
					conditions = append(conditions, map[string]interface{}{
						"type":    condition.Type,
						"status":  condition.Status,
						"reason":  condition.Reason,
						"message": condition.Message,
					})

					// Use the first non-true condition for the deployment issue
					if deploymentIssue.Reason == "" {
						deploymentIssue.Reason = condition.Reason
						deploymentIssue.Message = condition.Message
					}
				}
			}

			if len(conditions) > 0 {
				deploymentIssue.Conditions = conditions
			}
			
			// If we don't have a reason yet, create one based on the deployment status
			if deploymentIssue.Reason == "" {
				if deployment.Status.ReadyReplicas < deployment.Spec.Replicas {
					deploymentIssue.Reason = "InsufficientReadyReplicas"
					deploymentIssue.Message = fmt.Sprintf("Deployment has %d/%d ready replicas", 
						deployment.Status.ReadyReplicas, deployment.Spec.Replicas)
				} else if deployment.Status.UpdatedReplicas < deployment.Spec.Replicas {
					deploymentIssue.Reason = "StalledRollout"
					deploymentIssue.Message = fmt.Sprintf("Deployment rollout is stalled with %d/%d updated replicas", 
						deployment.Status.UpdatedReplicas, deployment.Spec.Replicas)
				}
			}
			
			// Get events related to this deployment
			eventsOutput, err := c.ExecuteKubectl("get", "events", "--field-selector=involvedObject.name="+deployment.Metadata.Name, 
				"-n", deployment.Metadata.Namespace, "-o", "json")
			
			if err == nil {
				var eventList struct {
					Items []map[string]interface{} `json:"items"`
				}
				
				if json.Unmarshal([]byte(eventsOutput), &eventList) == nil && len(eventList.Items) > 0 {
					events := make([]interface{}, len(eventList.Items))
					for i, item := range eventList.Items {
						events[i] = item
					}
					deploymentIssue.Events = events
				}
			}

			misconfiguredDeployments = append(misconfiguredDeployments, deploymentIssue)
		}
	}

	return misconfiguredDeployments, nil
}

// GetServiceIssues returns a list of service issues
func (c *Client) GetServiceIssues(ctx context.Context) ([]interface{}, error) {
	// Check if namespace exists
	if !c.NamespaceExists(c.GetCurrentNamespace()) {
		return nil, fmt.Errorf("namespace %q not found", c.GetCurrentNamespace())
	}

	// Get real service issues from the cluster
	services, err := c.getRealServiceIssues()
	if err != nil {
		return nil, fmt.Errorf("failed to get service issues: %w", err)
	}
	
	// Return empty slice if no service issues found
	if len(services) == 0 {
		return []interface{}{}, nil
	}
	
	return services, nil
}

// getRealServiceIssues attempts to get real service issues from the cluster
func (c *Client) getRealServiceIssues() ([]interface{}, error) {
	// Get all services
	output, err := c.ExecuteKubectl("get", "services", "-n", c.GetCurrentNamespace(), "-o", "json")
	if err != nil {
		return nil, err
	}

	var serviceList struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Spec struct {
				Type       string            `json:"type"`
				Selector   map[string]string `json:"selector"`
				ClusterIP  string            `json:"clusterIP"`
				Ports      []interface{}     `json:"ports"`
			} `json:"spec"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &serviceList); err != nil {
		return nil, err
	}

	var serviceIssues []interface{}
	for _, service := range serviceList.Items {
		// Skip services without selectors (like ExternalName services)
		if len(service.Spec.Selector) == 0 {
			continue
		}
		
		// Get endpoints for this service
		endpointsOutput, err := c.ExecuteKubectl("get", "endpoints", service.Metadata.Name, "-n", c.GetCurrentNamespace(), "-o", "json")
		if err != nil {
			// If we can't get endpoints, consider it an issue
			serviceIssues = append(serviceIssues, map[string]interface{}{
				"name":      service.Metadata.Name,
				"namespace": service.Metadata.Namespace,
				"type":      service.Spec.Type,
				"issue":     "Cannot retrieve endpoints",
				"message":   fmt.Sprintf("Error getting endpoints: %v", err),
			})
			continue
		}

		var endpoints struct {
			Subsets []struct {
				Addresses []interface{} `json:"addresses"`
				Ports     []interface{} `json:"ports"`
			} `json:"subsets"`
		}

		if err := json.Unmarshal([]byte(endpointsOutput), &endpoints); err != nil {
			serviceIssues = append(serviceIssues, map[string]interface{}{
				"name":      service.Metadata.Name,
				"namespace": service.Metadata.Namespace,
				"type":      service.Spec.Type,
				"issue":     "Invalid endpoint data",
				"message":   fmt.Sprintf("Error parsing endpoints: %v", err),
			})
			continue
		}

		// Check if service has no endpoints
		hasEndpoints := false
		for _, subset := range endpoints.Subsets {
			if len(subset.Addresses) > 0 {
				hasEndpoints = true
				break
			}
		}

		if !hasEndpoints {
			// Get pods matching the service selector to provide more context
			selectorString := []string{}
			for k, v := range service.Spec.Selector {
				selectorString = append(selectorString, fmt.Sprintf("%s=%s", k, v))
			}
			
			podListCmd := fmt.Sprintf("get pods -n %s -l %s", 
				c.GetCurrentNamespace(), 
				strings.Join(selectorString, ","))
			
			podsOutput, _ := c.ExecuteKubectl(strings.Split(podListCmd, " ")...)
			
			message := "Service has no endpoint pods"
			if podsOutput != "" {
				message = fmt.Sprintf("Service has no endpoint pods. Matching pods status: %s", podsOutput)
			}
			
			serviceIssues = append(serviceIssues, map[string]interface{}{
				"name":      service.Metadata.Name,
				"namespace": service.Metadata.Namespace,
				"type":      service.Spec.Type,
				"issue":     "No endpoints available",
				"message":   message,
				"selector":  service.Spec.Selector,
			})
		}
	}

	return serviceIssues, nil
}
