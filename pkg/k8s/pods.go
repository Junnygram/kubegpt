package k8s

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GetUnhealthyPodsLegacy gets all unhealthy pods in the specified namespace
// This is kept for reference but not used
func (c *Client) GetUnhealthyPodsLegacy(namespace string) ([]PodIssue, error) {
	// Get all pods in the namespace
	output, err := c.ExecuteKubectl("get", "pods", "-n", namespace, "-o", "json")
	if err != nil {
		return nil, err
	}

	// Parse the JSON output
	var podList struct {
		Items []struct {
			Metadata struct {
				Name      string    `json:"name"`
				Namespace string    `json:"namespace"`
				CreationTimestamp string `json:"creationTimestamp"`
			} `json:"metadata"`
			Spec struct {
				NodeName string `json:"nodeName"`
			} `json:"spec"`
			Status struct {
				Phase   string `json:"phase"`
				Message string `json:"message"`
				Reason  string `json:"reason"`
				ContainerStatuses []struct {
					Name      string `json:"name"`
					Image     string `json:"image"`
					Ready     bool   `json:"ready"`
					RestartCount int  `json:"restartCount"`
					State     struct {
						Waiting struct {
							Reason  string `json:"reason"`
							Message string `json:"message"`
						} `json:"waiting"`
						Terminated struct {
							Reason    string `json:"reason"`
							Message   string `json:"message"`
							ExitCode  int    `json:"exitCode"`
						} `json:"terminated"`
					} `json:"state"`
					LastState struct {
						Terminated struct {
							Reason    string `json:"reason"`
							Message   string `json:"message"`
							ExitCode  int    `json:"exitCode"`
						} `json:"terminated"`
					} `json:"lastState"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &podList); err != nil {
		return nil, fmt.Errorf("error parsing pod list: %w", err)
	}

	// Filter unhealthy pods
	var unhealthyPods []PodIssue
	for _, pod := range podList.Items {
		// Skip pods that are Running and have all containers ready
		if pod.Status.Phase == "Running" {
			allContainersReady := true
			for _, container := range pod.Status.ContainerStatuses {
				if !container.Ready || container.RestartCount > 5 {
					allContainersReady = false
					break
				}
			}
			if allContainersReady {
				continue
			}
		}

		// Skip pods that are Completed successfully
		if pod.Status.Phase == "Succeeded" {
			continue
		}

		// Create a PodIssue for the unhealthy pod
		podIssue := PodIssue{
			Name:      pod.Metadata.Name,
			Namespace: pod.Metadata.Namespace,
			Status:    pod.Status.Phase,
			Message:   pod.Status.Message,
			Reason:    pod.Status.Reason,
			Node:      pod.Spec.NodeName,
		}

		// Parse creation timestamp
		if creationTime, err := time.Parse(time.RFC3339, pod.Metadata.CreationTimestamp); err == nil {
			podIssue.Age = time.Since(creationTime)
		}

		// Add container issues
		for _, container := range pod.Status.ContainerStatuses {
			containerIssue := ContainerIssue{
				Name:     container.Name,
				Image:    container.Image,
				Ready:    container.Ready,
				Restarts: container.RestartCount,
			}

			// Determine container status and reason
			if !container.Ready {
				if container.State.Waiting.Reason != "" {
					containerIssue.Status = "Waiting"
					containerIssue.Reason = container.State.Waiting.Reason
					containerIssue.Message = container.State.Waiting.Message
				} else if container.State.Terminated.Reason != "" {
					containerIssue.Status = "Terminated"
					containerIssue.Reason = container.State.Terminated.Reason
					containerIssue.Message = container.State.Terminated.Message
					containerIssue.ExitCode = container.State.Terminated.ExitCode
				} else {
					containerIssue.Status = "Not Ready"
				}
			} else {
				containerIssue.Status = "Running"
				
				// Check for high restart count
				if container.RestartCount > 5 {
					containerIssue.Reason = "HighRestartCount"
					containerIssue.Message = fmt.Sprintf("Container has restarted %d times", container.RestartCount)
					
					// Check last termination reason
					if container.LastState.Terminated.Reason != "" {
						containerIssue.Message += fmt.Sprintf(", last exit reason: %s (code: %d)", 
							container.LastState.Terminated.Reason, 
							container.LastState.Terminated.ExitCode)
						containerIssue.ExitCode = container.LastState.Terminated.ExitCode
					}
				}
			}

			podIssue.Containers = append(podIssue.Containers, containerIssue)
		}

		// Get pod events
		podIssue.Events = c.getPodEvents(pod.Metadata.Name, pod.Metadata.Namespace)

		// Get pod logs for containers with issues
		podIssue.Logs = make(map[string]string)
		for _, container := range podIssue.Containers {
			if !container.Ready || container.Restarts > 5 {
				logs, _ := c.getPodLogs(pod.Metadata.Name, pod.Metadata.Namespace, container.Name, 50)
				podIssue.Logs[container.Name] = logs
			}
		}

		unhealthyPods = append(unhealthyPods, podIssue)
	}

	return unhealthyPods, nil
}

// getPodEvents gets events for a specific pod
func (c *Client) getPodEvents(podName, namespace string) []interface{} {
	output, err := c.ExecuteKubectl("get", "events", "-n", namespace, "--field-selector", fmt.Sprintf("involvedObject.name=%s", podName), "-o", "json")
	if err != nil {
		return nil
	}

	var eventList struct {
		Items []interface{} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &eventList); err != nil {
		return nil
	}

	return eventList.Items
}

// getPodLogs gets logs for a specific container in a pod
func (c *Client) getPodLogs(podName, namespace, containerName string, tailLines int) (string, error) {
	args := []string{"logs", "-n", namespace, podName, "-c", containerName}
	
	if tailLines > 0 {
		args = append(args, "--tail", strconv.Itoa(tailLines))
	}
	
	output, err := c.ExecuteKubectl(args...)
	if err != nil {
		// Try to get previous logs if current logs fail
		args = append(args, "--previous")
		output, err = c.ExecuteKubectl(args...)
		if err != nil {
			return "", err
		}
	}
	
	// Truncate logs if they're too long
	if len(output) > 2000 {
		lines := strings.Split(output, "\n")
		if len(lines) > 20 {
			// Take the first 5 and last 15 lines
			truncatedLines := append(lines[:5], append([]string{"...[logs truncated]..."}, lines[len(lines)-15:]...)...)
			output = strings.Join(truncatedLines, "\n")
		}
	}
	
	return output, nil
}