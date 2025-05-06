package k8s

import (
	"time"
)

// PodIssue represents an issue with a pod
type PodIssue struct {
	Name       string
	Namespace  string
	Status     string
	Message    string
	Reason     string
	Node       string
	Age        time.Duration
	Containers []ContainerIssue
	Events     []interface{}
	Logs       map[string]string
	Analysis   string
	Fix        string
}

// ContainerIssue represents an issue with a container
type ContainerIssue struct {
	Name      string
	Image     string
	Ready     bool
	Status    string
	Restarts  int
	Reason    string
	Message   string
	ExitCode  int
}

// DeploymentIssue represents an issue with a deployment
type DeploymentIssue struct {
	Name             string
	Namespace        string
	Replicas         int
	ReadyReplicas    int
	UpdatedReplicas  int
	AvailableReplicas int
	Strategy         string
	Age              time.Duration
	Message          string
	Reason           string
	Conditions       interface{}
	Events           []interface{}
	Analysis         string
	Fix              string
}