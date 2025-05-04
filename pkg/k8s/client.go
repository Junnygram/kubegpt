package k8s

import (
	"context"
	"fmt"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wraps the Kubernetes client-go functionality
type Client struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath string) (*Client, error) {
	var config *rest.Config
	var err error

	// Use provided kubeconfig or default path
	if kubeconfigPath == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	// Try to build config from kubeconfig file
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		// If that fails, try in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes client config: %w", err)
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	// Get current namespace from kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfigPath
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		namespace = "default"
	}

	return &Client{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// SetNamespace sets the namespace for the client
func (c *Client) SetNamespace(namespace string) {
	c.namespace = namespace
}

// GetCurrentNamespace returns the current namespace
func (c *Client) GetCurrentNamespace() string {
	return c.namespace
}

// GetAllNamespaces returns a list of all namespaces in the cluster
func (c *Client) GetAllNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var namespaceNames []string
	for _, ns := range namespaces.Items {
		namespaceNames = append(namespaceNames, ns.Name)
	}

	return namespaceNames, nil
}

// GetPodLogs retrieves logs for a specific pod
func (c *Client) GetPodLogs(ctx context.Context, podName, containerName string, tailLines int64) (string, error) {
	podLogOptions := v1.PodLogOptions{
		Container: containerName,
	}
	if tailLines > 0 {
		podLogOptions.TailLines = &tailLines
	}

	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(podName, &podLogOptions)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs for pod %s: %w", podName, err)
	}
	defer podLogs.Close()

	buf := make([]byte, 2048)
	var logs string
	for {
		n, err := podLogs.Read(buf)
		if err != nil {
			break
		}
		logs += string(buf[:n])
	}

	return logs, nil
}

// GetPodEvents retrieves events for a specific pod
func (c *Client) GetPodEvents(ctx context.Context, podName string) ([]v1.Event, error) {
	// Get the pod to get its UID
	pod, err := c.clientset.CoreV1().Pods(c.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s: %w", podName, err)
	}

	// List events for the pod
	fieldSelector := fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.uid=%s", 
		podName, c.namespace, pod.UID)
	events, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events for pod %s: %w", podName, err)
	}

	return events.Items, nil
}

// GetNodeInfo retrieves information about a specific node
func (c *Client) GetNodeInfo(ctx context.Context, nodeName string) (*v1.Node, error) {
	node, err := c.clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	return node, nil
}

// GetClusterInfo retrieves basic information about the cluster
func (c *Client) GetClusterInfo(ctx context.Context) (map[string]string, error) {
	info := make(map[string]string)

	// Get Kubernetes version
	version, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}
	info["version"] = version.String()

	// Get node count
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	info["nodes"] = fmt.Sprintf("%d", len(nodes.Items))

	// Get namespace count
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	info["namespaces"] = fmt.Sprintf("%d", len(namespaces.Items))

	return info, nil
}