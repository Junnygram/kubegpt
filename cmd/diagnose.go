package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/junioroyewunmi/kubegpt/pkg/ai"
	"github.com/junioroyewunmi/kubegpt/pkg/k8s"
	"github.com/junioroyewunmi/kubegpt/pkg/utils"
)

var (
	podsOnly    bool
	deploymentsOnly bool
	allNamespaces bool
)

// diagnoseCmd represents the diagnose command
var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose issues in your Kubernetes cluster",
	Long: `Diagnose issues in your Kubernetes cluster using Amazon Q Developer.

This command will:
1. Connect to your Kubernetes cluster
2. Identify unhealthy resources (pods, deployments, etc.)
3. Analyze the issues using Amazon Q Developer
4. Provide explanations and suggested fixes

Examples:
  # Diagnose issues in the current namespace
  kubegpt diagnose

  # Diagnose issues in a specific namespace
  kubegpt diagnose --namespace monitoring

  # Diagnose only pod issues
  kubegpt diagnose --pods-only

  # Generate YAML patches to fix issues
  kubegpt diagnose --fix
`,
	Run: func(cmd *cobra.Command, args []string) {
		runDiagnose()
	},
}

func init() {
	diagnoseCmd.Flags().BoolVar(&podsOnly, "pods-only", false, "only diagnose pod issues")
	diagnoseCmd.Flags().BoolVar(&deploymentsOnly, "deployments-only", false, "only diagnose deployment issues")
	diagnoseCmd.Flags().BoolVar(&allNamespaces, "all-namespaces", false, "diagnose issues in all namespaces")
}

func runDiagnose() {
	printLogo()

	// Create Kubernetes client
	client := k8s.NewClient(kubeconfig, namespace)

	// Get current namespace if not specified
	currentNamespace := namespace
	var err error
	if currentNamespace == "" {
		currentNamespace, err = client.GetCurrentNamespace()
		if err != nil {
			color.New(color.FgRed).Printf("Error getting current namespace: %s\n", utils.FormatError(err))
			return
		}
	}

	// Get namespaces to diagnose
	var namespaces []string
	if allNamespaces {
		namespaces, err = client.GetNamespaces()
		if err != nil {
			color.New(color.FgRed).Printf("Error getting namespaces: %s\n", utils.FormatError(err))
			return
		}
	} else {
		namespaces = []string{currentNamespace}
	}

	// Create Amazon Q client
	amazonQClient := ai.NewAmazonQClient()

	// Diagnose each namespace
	for _, ns := range namespaces {
		diagnoseNamespace(client, amazonQClient, ns)
	}
}

func diagnoseNamespace(client *k8s.Client, amazonQClient *ai.AmazonQClient, namespace string) {
	color.New(color.FgYellow).Printf("Diagnosing issues in namespace: %s\n\n", namespace)

	// Track unhealthy resources
	var unhealthyPods []k8s.PodIssue
	var unhealthyDeployments []k8s.DeploymentIssue
	var failedEvents []interface{}

	// Check pods
	if !deploymentsOnly {
		fmt.Println("Checking pods...")
		pods, err := client.GetUnhealthyPods(namespace)
		if err != nil {
			color.New(color.FgRed).Printf("Error checking pods: %s\n", utils.FormatError(err))
		} else {
			unhealthyPods = pods
			color.New(color.FgYellow).Printf("Found %d unhealthy pods\n\n", len(pods))
		}
	}

	// Check deployments
	if !podsOnly {
		fmt.Println("Checking deployments...")
		deployments, err := client.GetUnhealthyDeployments(namespace)
		if err != nil {
			color.New(color.FgRed).Printf("Error checking deployments: %s\n", utils.FormatError(err))
		} else {
			unhealthyDeployments = deployments
			color.New(color.FgYellow).Printf("Found %d unhealthy deployments\n\n", len(deployments))
		}
	}

	// Check events
	if !podsOnly && !deploymentsOnly {
		fmt.Println("Checking events...")
		events, err := client.GetFailedEvents(namespace)
		if err != nil {
			color.New(color.FgRed).Printf("Error checking events: %s\n", utils.FormatError(err))
		} else {
			failedEvents = events
			color.New(color.FgYellow).Printf("Found %d failed events\n\n", len(events))
		}
	}

	// Analyze issues with Amazon Q
	if len(unhealthyPods) > 0 || len(unhealthyDeployments) > 0 {
		color.New(color.FgCyan).Println("Analyzing issues with Amazon Q...")

		// Analyze pod issues
		for i := range unhealthyPods {
			color.New(color.FgCyan).Printf("Analyzing pod %s...\n", unhealthyPods[i].Name)
			analysis, err := amazonQClient.AnalyzePodIssue(unhealthyPods[i])
			if err != nil {
				color.New(color.FgRed).Printf("Error analyzing pod: %s\n", utils.FormatError(err))
			} else {
				unhealthyPods[i].Analysis = analysis
			}

			// Generate fix if requested
			if fix {
				color.New(color.FgCyan).Printf("Generating fix for pod %s...\n", unhealthyPods[i].Name)
				fixSuggestion, err := amazonQClient.GeneratePodFix(unhealthyPods[i])
				if err != nil {
					color.New(color.FgRed).Printf("Error generating fix: %s\n", utils.FormatError(err))
				} else {
					unhealthyPods[i].Fix = fixSuggestion
				}
			}
		}

		// Analyze deployment issues
		for i := range unhealthyDeployments {
			color.New(color.FgCyan).Printf("Analyzing deployment %s...\n", unhealthyDeployments[i].Name)
			analysis, err := amazonQClient.AnalyzeDeploymentIssue(unhealthyDeployments[i])
			if err != nil {
				color.New(color.FgRed).Printf("Error analyzing deployment: %s\n", utils.FormatError(err))
			} else {
				unhealthyDeployments[i].Analysis = analysis
			}

			// Generate fix if requested
			if fix {
				color.New(color.FgCyan).Printf("Generating fix for deployment %s...\n", unhealthyDeployments[i].Name)
				fixSuggestion, err := amazonQClient.GenerateDeploymentFix(unhealthyDeployments[i])
				if err != nil {
					color.New(color.FgRed).Printf("Error generating fix: %s\n", utils.FormatError(err))
				} else {
					unhealthyDeployments[i].Fix = fixSuggestion
				}
			}
		}
	}

	// Print diagnostic results
	fmt.Println()
	color.New(color.FgGreen, color.Bold).Printf("Diagnostic Results for Namespace: %s\n", namespace)
	color.New(color.FgWhite).Printf("Time: %s\n\n", time.Now().Format(time.RFC1123))

	// Print summary
	color.New(color.FgYellow, color.Bold).Println("Summary:")
	color.New(color.FgWhite).Printf("- Unhealthy Pods: %d\n", len(unhealthyPods))
	color.New(color.FgWhite).Printf("- Unhealthy Deployments: %d\n", len(unhealthyDeployments))
	color.New(color.FgWhite).Printf("- Failed Events: %d\n", len(failedEvents))
	fmt.Println()

	// Print pod issues
	if len(unhealthyPods) > 0 {
		color.New(color.FgYellow, color.Bold).Println("Unhealthy Pods:")
		fmt.Println()

		for i, pod := range unhealthyPods {
			color.New(color.FgCyan, color.Bold).Printf("[%d] Pod: %s\n", i+1, pod.Name)
			color.New(color.FgWhite).Printf("    Status: %s\n", pod.Status)
			
			if pod.Reason != "" {
				color.New(color.FgWhite).Printf("    Reason: %s\n", pod.Reason)
			}
			
			if pod.Message != "" {
				color.New(color.FgWhite).Printf("    Message: %s\n", pod.Message)
			}
			
			if len(pod.Containers) > 0 {
				color.New(color.FgWhite).Println("    Container Issues:")
				for _, container := range pod.Containers {
					color.New(color.FgWhite).Printf("    - %s: %s (Restarts: %d)\n", container.Name, container.Status, container.Restarts)
					if container.Reason != "" {
						color.New(color.FgWhite).Printf("      Reason: %s\n", container.Reason)
					}
					if container.Message != "" {
						color.New(color.FgWhite).Printf("      Message: %s\n", container.Message)
					}
				}
			}
			
			fmt.Println()
			
			if pod.Analysis != "" {
				color.New(color.FgGreen).Println("    Analysis:")
				fmt.Printf("    %s\n", strings.ReplaceAll(pod.Analysis, "\n", "\n    "))
				fmt.Println()
			}
			
			if pod.Fix != "" {
				color.New(color.FgGreen).Println("    Suggested Fix:")
				fmt.Printf("    %s\n", strings.ReplaceAll(pod.Fix, "\n", "\n    "))
				fmt.Println()
			}
			
			fmt.Println()
		}
	}

	// Print deployment issues
	if len(unhealthyDeployments) > 0 {
		color.New(color.FgYellow, color.Bold).Println("Unhealthy Deployments:")
		fmt.Println()

		for i, deployment := range unhealthyDeployments {
			color.New(color.FgCyan, color.Bold).Printf("[%d] Deployment: %s\n", i+1, deployment.Name)
			color.New(color.FgWhite).Printf("    Replicas: %d/%d ready\n", deployment.ReadyReplicas, deployment.Replicas)
			
			if deployment.Reason != "" {
				color.New(color.FgWhite).Printf("    Reason: %s\n", deployment.Reason)
			}
			
			if deployment.Message != "" {
				color.New(color.FgWhite).Printf("    Message: %s\n", deployment.Message)
			}
			
			fmt.Println()
			
			if deployment.Analysis != "" {
				color.New(color.FgGreen).Println("    Analysis:")
				fmt.Printf("    %s\n", strings.ReplaceAll(deployment.Analysis, "\n", "\n    "))
				fmt.Println()
			}
			
			if deployment.Fix != "" {
				color.New(color.FgGreen).Println("    Suggested Fix:")
				fmt.Printf("    %s\n", strings.ReplaceAll(deployment.Fix, "\n", "\n    "))
				fmt.Println()
			}
			
			fmt.Println()
		}
	}

	// Print failed events
	if len(failedEvents) > 0 {
		color.New(color.FgYellow, color.Bold).Println("Failed Events:")
		fmt.Println()

		// Print events (simplified for now)
		for i, event := range failedEvents {
			if eventMap, ok := event.(map[string]interface{}); ok {
				reason := eventMap["reason"]
				message := eventMap["message"]
				involvedObject := eventMap["involvedObject"]
				
				if involvedObjectMap, ok := involvedObject.(map[string]interface{}); ok {
					kind := involvedObjectMap["kind"]
					name := involvedObjectMap["name"]
					
					color.New(color.FgCyan, color.Bold).Printf("[%d] %s: %s\n", i+1, kind, name)
					color.New(color.FgWhite).Printf("    Reason: %s\n", reason)
					color.New(color.FgWhite).Printf("    Message: %s\n", message)
					fmt.Println()
				}
			}
		}
	}

	// No issues found
	if len(unhealthyPods) == 0 && len(unhealthyDeployments) == 0 && len(failedEvents) == 0 {
		color.New(color.FgGreen).Println("No issues found in this namespace! 🎉")
		fmt.Println()
	}
}