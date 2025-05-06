package k8s

import (
	"encoding/json"
	"fmt"
	"time"
)

// GetUnhealthyDeployments gets all unhealthy deployments in the specified namespace
func (c *Client) GetUnhealthyDeployments(namespace string) ([]DeploymentIssue, error) {
	// Get all deployments in the namespace
	output, err := c.ExecuteKubectl("get", "deployments", "-n", namespace, "-o", "json")
	if err != nil {
		return nil, err
	}

	// Parse the JSON output
	var deploymentList struct {
		Items []struct {
			Metadata struct {
				Name      string    `json:"name"`
				Namespace string    `json:"namespace"`
				CreationTimestamp string `json:"creationTimestamp"`
			} `json:"metadata"`
			Spec struct {
				Replicas int `json:"replicas"`
				Strategy struct {
					Type string `json:"type"`
				} `json:"strategy"`
			} `json:"spec"`
			Status struct {
				Replicas          int `json:"replicas"`
				ReadyReplicas     int `json:"readyReplicas"`
				UpdatedReplicas   int `json:"updatedReplicas"`
				AvailableReplicas int `json:"availableReplicas"`
				Conditions        []struct {
					Type    string `json:"type"`
					Status  string `json:"status"`
					Reason  string `json:"reason"`
					Message string `json:"message"`
				} `json:"conditions"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &deploymentList); err != nil {
		return nil, fmt.Errorf("error parsing deployment list: %w", err)
	}

	// Filter unhealthy deployments
	var unhealthyDeployments []DeploymentIssue
	for _, deployment := range deploymentList.Items {
		// Skip deployments with 0 replicas (scaled down)
		if deployment.Spec.Replicas == 0 {
			continue
		}

		// Skip healthy deployments
		if deployment.Status.ReadyReplicas == deployment.Spec.Replicas &&
			deployment.Status.UpdatedReplicas == deployment.Spec.Replicas &&
			deployment.Status.AvailableReplicas == deployment.Spec.Replicas {
			continue
		}

		// Create a DeploymentIssue for the unhealthy deployment
		deploymentIssue := DeploymentIssue{
			Name:              deployment.Metadata.Name,
			Namespace:         deployment.Metadata.Namespace,
			Replicas:          deployment.Spec.Replicas,
			ReadyReplicas:     deployment.Status.ReadyReplicas,
			UpdatedReplicas:   deployment.Status.UpdatedReplicas,
			AvailableReplicas: deployment.Status.AvailableReplicas,
			Strategy:          deployment.Spec.Strategy.Type,
		}

		// Parse creation timestamp
		if creationTime, err := time.Parse(time.RFC3339, deployment.Metadata.CreationTimestamp); err == nil {
			deploymentIssue.Age = time.Since(creationTime)
		}

		// Determine deployment status and reason
		if deployment.Status.ReadyReplicas < deployment.Spec.Replicas {
			deploymentIssue.Reason = "InsufficientReplicas"
			deploymentIssue.Message = fmt.Sprintf("Only %d/%d replicas are ready", 
				deployment.Status.ReadyReplicas, deployment.Spec.Replicas)
		}

		// Check conditions for more specific reasons
		for _, condition := range deployment.Status.Conditions {
			if condition.Status != "True" && condition.Type == "Available" {
				deploymentIssue.Reason = condition.Reason
				deploymentIssue.Message = condition.Message
				break
			}
		}

		// Add conditions
		deploymentIssue.Conditions = deployment.Status.Conditions

		// Get deployment events
		deploymentIssue.Events = c.getDeploymentEvents(deployment.Metadata.Name, deployment.Metadata.Namespace)

		unhealthyDeployments = append(unhealthyDeployments, deploymentIssue)
	}

	return unhealthyDeployments, nil
}

// getDeploymentEvents gets events for a specific deployment
func (c *Client) getDeploymentEvents(deploymentName, namespace string) []interface{} {
	output, err := c.ExecuteKubectl("get", "events", "-n", namespace, "--field-selector", fmt.Sprintf("involvedObject.name=%s", deploymentName), "-o", "json")
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