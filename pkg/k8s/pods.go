package k8s

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodIssue represents an issue with a pod
type PodIssue struct {
	Name       string
	Namespace  string
	Status     string
	Containers []ContainerIssue
	Events     []v1.Event
	Logs       map[string]string // Container name -> logs
	Node       string
	Age        time.Duration
	Message    string
	Reason     string
	Analysis   string
	Fix        string
}

// ContainerIssue represents an issue with a container
type ContainerIssue struct {
	Name      string
	Image     string
	Ready     bool
	Status    string
	Restarts  int32
	Message   string
	Reason    string
	ExitCode  int32
	StartTime *metav1.Time
}

// GetUnhealthyPods returns a list of unhealthy pods in the current namespace
func (c *Client) GetUnhealthyPods(ctx context.Context) ([]PodIssue, error) {
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var unhealthyPods []PodIssue

	for _, pod := range pods.Items {
		// Skip if pod is running and ready
		if isPodHealthy(&pod) {
			continue
		}

		// Create pod issue
		podIssue := PodIssue{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Node:      pod.Spec.NodeName,
			Age:       time.Since(pod.CreationTimestamp.Time),
		}

		// Add container issues
		for _, containerStatus := range pod.Status.ContainerStatuses {
			containerIssue := ContainerIssue{
				Name:     containerStatus.Name,
				Image:    containerStatus.Image,
				Ready:    containerStatus.Ready,
				Restarts: containerStatus.RestartCount,
			}

			// Get container state
			if containerStatus.State.Waiting != nil {
				containerIssue.Status = "Waiting"
				containerIssue.Reason = containerStatus.State.Waiting.Reason
				containerIssue.Message = containerStatus.State.Waiting.Message
			} else if containerStatus.State.Running != nil {
				containerIssue.Status = "Running"
				containerIssue.StartTime = &metav1.Time{Time: containerStatus.State.Running.StartedAt.Time}
			} else if containerStatus.State.Terminated != nil {
				containerIssue.Status = "Terminated"
				containerIssue.Reason = containerStatus.State.Terminated.Reason
				containerIssue.Message = containerStatus.State.Terminated.Message
				containerIssue.ExitCode = containerStatus.State.Terminated.ExitCode
			}

			podIssue.Containers = append(podIssue.Containers, containerIssue)
		}

		// Get pod events
		events, err := c.GetPodEvents(ctx, pod.Name)
		if err == nil {
			podIssue.Events = events
		}

		// Get logs for each container
		podIssue.Logs = make(map[string]string)
		for _, container := range pod.Spec.Containers {
			logs, err := c.GetPodLogs(ctx, pod.Name, container.Name, 50)
			if err == nil {
				podIssue.Logs[container.Name] = logs
			}
		}

		// Add pod message and reason
		for _, condition := range pod.Status.Conditions {
			if condition.Status == v1.ConditionFalse {
				podIssue.Message = condition.Message
				podIssue.Reason = condition.Reason
				break
			}
		}

		unhealthyPods = append(unhealthyPods, podIssue)
	}

	return unhealthyPods, nil
}

// GetPodStats returns lists of healthy and unhealthy pods
func (c *Client) GetPodStats(ctx context.Context) ([]v1.Pod, []PodIssue, error) {
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var healthyPods []v1.Pod
	var unhealthyPodIssues []PodIssue

	for _, pod := range pods.Items {
		if isPodHealthy(&pod) {
			healthyPods = append(healthyPods, pod)
		} else {
			// Create pod issue (simplified version without logs and events)
			podIssue := PodIssue{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				Status:    string(pod.Status.Phase),
				Node:      pod.Spec.NodeName,
				Age:       time.Since(pod.CreationTimestamp.Time),
			}

			// Add container issues
			for _, containerStatus := range pod.Status.ContainerStatuses {
				containerIssue := ContainerIssue{
					Name:     containerStatus.Name,
					Image:    containerStatus.Image,
					Ready:    containerStatus.Ready,
					Restarts: containerStatus.RestartCount,
				}

				// Get container state
				if containerStatus.State.Waiting != nil {
					containerIssue.Status = "Waiting"
					containerIssue.Reason = containerStatus.State.Waiting.Reason
					containerIssue.Message = containerStatus.State.Waiting.Message
				} else if containerStatus.State.Running != nil {
					containerIssue.Status = "Running"
					containerIssue.StartTime = &metav1.Time{Time: containerStatus.State.Running.StartedAt.Time}
				} else if containerStatus.State.Terminated != nil {
					containerIssue.Status = "Terminated"
					containerIssue.Reason = containerStatus.State.Terminated.Reason
					containerIssue.Message = containerStatus.State.Terminated.Message
					containerIssue.ExitCode = containerStatus.State.Terminated.ExitCode
				}

				podIssue.Containers = append(podIssue.Containers, containerIssue)
			}

			// Add pod message and reason
			for _, condition := range pod.Status.Conditions {
				if condition.Status == v1.ConditionFalse {
					podIssue.Message = condition.Message
					podIssue.Reason = condition.Reason
					break
				}
			}

			unhealthyPodIssues = append(unhealthyPodIssues, podIssue)
		}
	}

	return healthyPods, unhealthyPodIssues, nil
}

// isPodHealthy checks if a pod is healthy (running and all containers ready)
func isPodHealthy(pod *v1.Pod) bool {
	if pod.Status.Phase != v1.PodRunning {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady && condition.Status != v1.ConditionTrue {
			return false
		}
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if !containerStatus.Ready || containerStatus.RestartCount > 5 {
			return false
		}
	}

	return true
}