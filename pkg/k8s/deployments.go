package k8s

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentIssue represents an issue with a deployment
type DeploymentIssue struct {
	Name           string
	Namespace      string
	Replicas       int32
	ReadyReplicas  int32
	UpdatedReplicas int32
	AvailableReplicas int32
	Strategy       string
	Conditions     []appsv1.DeploymentCondition
	Events         []v1.Event
	Age            time.Duration
	Message        string
	Reason         string
	Analysis       string
	Fix            string
}

// ServiceIssue represents an issue with a service
type ServiceIssue struct {
	Name           string
	Namespace      string
	Type           string
	ClusterIP      string
	ExternalIP     string
	Ports          []v1.ServicePort
	Selector       map[string]string
	Events         []v1.Event
	Age            time.Duration
	Message        string
	Reason         string
	Analysis       string
	Fix            string
	EndpointCount  int
}

// GetMisconfiguredDeployments returns a list of misconfigured deployments in the current namespace
func (c *Client) GetMisconfiguredDeployments(ctx context.Context) ([]DeploymentIssue, error) {
	deployments, err := c.clientset.AppsV1().Deployments(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var misconfiguredDeployments []DeploymentIssue

	for _, deployment := range deployments.Items {
		// Skip if deployment is healthy
		if isDeploymentHealthy(&deployment) {
			continue
		}

		// Create deployment issue
		deploymentIssue := DeploymentIssue{
			Name:             deployment.Name,
			Namespace:        deployment.Namespace,
			Replicas:         *deployment.Spec.Replicas,
			ReadyReplicas:    deployment.Status.ReadyReplicas,
			UpdatedReplicas:  deployment.Status.UpdatedReplicas,
			AvailableReplicas: deployment.Status.AvailableReplicas,
			Strategy:         string(deployment.Spec.Strategy.Type),
			Conditions:       deployment.Status.Conditions,
			Age:              time.Since(deployment.CreationTimestamp.Time),
		}

		// Get deployment events
		events, err := c.GetResourceEvents(ctx, "Deployment", deployment.Name)
		if err == nil {
			deploymentIssue.Events = events
		}

		// Add deployment message and reason
		for _, condition := range deployment.Status.Conditions {
			if condition.Status != v1.ConditionTrue {
				deploymentIssue.Message = condition.Message
				deploymentIssue.Reason = condition.Reason
				break
			}
		}

		misconfiguredDeployments = append(misconfiguredDeployments, deploymentIssue)
	}

	return misconfiguredDeployments, nil
}

// GetDeploymentStats returns lists of healthy and misconfigured deployments
func (c *Client) GetDeploymentStats(ctx context.Context) ([]appsv1.Deployment, []DeploymentIssue, error) {
	deployments, err := c.clientset.AppsV1().Deployments(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var healthyDeployments []appsv1.Deployment
	var misconfiguredDeploymentIssues []DeploymentIssue

	for _, deployment := range deployments.Items {
		if isDeploymentHealthy(&deployment) {
			healthyDeployments = append(healthyDeployments, deployment)
		} else {
			// Create deployment issue (simplified version without events)
			deploymentIssue := DeploymentIssue{
				Name:             deployment.Name,
				Namespace:        deployment.Namespace,
				Replicas:         *deployment.Spec.Replicas,
				ReadyReplicas:    deployment.Status.ReadyReplicas,
				UpdatedReplicas:  deployment.Status.UpdatedReplicas,
				AvailableReplicas: deployment.Status.AvailableReplicas,
				Strategy:         string(deployment.Spec.Strategy.Type),
				Conditions:       deployment.Status.Conditions,
				Age:              time.Since(deployment.CreationTimestamp.Time),
			}

			// Add deployment message and reason
			for _, condition := range deployment.Status.Conditions {
				if condition.Status != v1.ConditionTrue {
					deploymentIssue.Message = condition.Message
					deploymentIssue.Reason = condition.Reason
					break
				}
			}

			misconfiguredDeploymentIssues = append(misconfiguredDeploymentIssues, deploymentIssue)
		}
	}

	return healthyDeployments, misconfiguredDeploymentIssues, nil
}

// GetServiceIssues returns a list of service issues in the current namespace
func (c *Client) GetServiceIssues(ctx context.Context) ([]ServiceIssue, error) {
	services, err := c.clientset.CoreV1().Services(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var serviceIssues []ServiceIssue

	for _, service := range services.Items {
		// Skip if service is healthy
		if isServiceHealthy(ctx, c, &service) {
			continue
		}

		// Create service issue
		serviceIssue := ServiceIssue{
			Name:      service.Name,
			Namespace: service.Namespace,
			Type:      string(service.Spec.Type),
			ClusterIP: service.Spec.ClusterIP,
			Ports:     service.Spec.Ports,
			Selector:  service.Spec.Selector,
			Age:       time.Since(service.CreationTimestamp.Time),
		}

		// Get external IP if available
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			serviceIssue.ExternalIP = service.Status.LoadBalancer.Ingress[0].IP
			if serviceIssue.ExternalIP == "" {
				serviceIssue.ExternalIP = service.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		// Get service events
		events, err := c.GetResourceEvents(ctx, "Service", service.Name)
		if err == nil {
			serviceIssue.Events = events
		}

		// Get endpoint count
		endpoints, err := c.clientset.CoreV1().Endpoints(c.namespace).Get(ctx, service.Name, metav1.GetOptions{})
		if err == nil {
			count := 0
			for _, subset := range endpoints.Subsets {
				count += len(subset.Addresses)
			}
			serviceIssue.EndpointCount = count
		}

		// Set message and reason based on endpoint count
		if serviceIssue.EndpointCount == 0 {
			serviceIssue.Message = "No endpoints available"
			serviceIssue.Reason = "EndpointsNotFound"
		}

		serviceIssues = append(serviceIssues, serviceIssue)
	}

	return serviceIssues, nil
}

// isDeploymentHealthy checks if a deployment is healthy
func isDeploymentHealthy(deployment *appsv1.Deployment) bool {
	// Check if desired replicas match ready replicas
	if deployment.Status.ReadyReplicas != *deployment.Spec.Replicas {
		return false
	}

	// Check if all replicas are updated
	if deployment.Status.UpdatedReplicas != *deployment.Spec.Replicas {
		return false
	}

	// Check if all replicas are available
	if deployment.Status.AvailableReplicas != *deployment.Spec.Replicas {
		return false
	}

	// Check deployment conditions
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable && condition.Status != v1.ConditionTrue {
			return false
		}
	}

	return true
}

// isServiceHealthy checks if a service is healthy
func isServiceHealthy(ctx context.Context, c *Client, service *v1.Service) bool {
	// Skip headless services
	if service.Spec.ClusterIP == "None" {
		return true
	}

	// Skip services without selectors (manually managed endpoints)
	if len(service.Spec.Selector) == 0 {
		return true
	}

	// Check if service has endpoints
	endpoints, err := c.clientset.CoreV1().Endpoints(service.Namespace).Get(ctx, service.Name, metav1.GetOptions{})
	if err != nil {
		return false
	}

	// Check if endpoints have addresses
	for _, subset := range endpoints.Subsets {
		if len(subset.Addresses) > 0 {
			return true
		}
	}

	return false
}