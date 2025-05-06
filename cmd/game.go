package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	difficulty string
)

// gameCmd represents the game command
var gameCmd = &cobra.Command{
	Use:   "game",
	Short: "Play a Kubernetes troubleshooting game",
	Long: `Play an interactive game that tests your Kubernetes troubleshooting skills.
The game presents you with common Kubernetes issues and challenges you to solve them.

Amazon Q Developer will provide hints and explanations to help you learn.

Examples:
  # Start a new game with easy difficulty
  kubegpt game --difficulty easy

  # Start a new game with hard difficulty
  kubegpt game --difficulty hard
`,
	Run: func(cmd *cobra.Command, args []string) {
		runGame()
	},
}

func init() {
	rootCmd.AddCommand(gameCmd)
	gameCmd.Flags().StringVar(&difficulty, "difficulty", "medium", "game difficulty (easy, medium, hard)")
}

func runGame() {
	printLogo()
	color.New(color.FgGreen, color.Bold).Println("🎮 KUBERNETES TROUBLESHOOTING CHALLENGE 🎮")
	fmt.Println()
	
	color.New(color.FgYellow).Println("Welcome to the Kubernetes Troubleshooting Challenge!")
	fmt.Println("Test your skills by diagnosing and fixing common Kubernetes issues.")
	fmt.Println("Amazon Q Developer will be your guide and judge.")
	fmt.Println()
	
	color.New(color.FgCyan).Printf("Difficulty: %s\n", strings.ToUpper(difficulty))
	fmt.Println()
	
	// Start the game
	startGame()
}

func startGame() {
	scenarios := getScenarios()
	score := 0
	totalScenarios := len(scenarios)
	
	for i, scenario := range scenarios {
		color.New(color.FgMagenta, color.Bold).Printf("SCENARIO %d/%d: %s\n", i+1, totalScenarios, scenario.title)
		fmt.Println()
		color.New(color.FgWhite).Println(scenario.description)
		fmt.Println()
		
		color.New(color.FgCyan).Println("Resources:")
		for _, resource := range scenario.resources {
			fmt.Printf("- %s\n", resource)
		}
		fmt.Println()
		
		color.New(color.FgRed).Println("Error/Issue:")
		color.New(color.FgRed).Println(scenario.error)
		fmt.Println()
		
		// Simulate thinking time
		color.New(color.FgYellow).Print("Analyzing with Amazon Q Developer")
		for i := 0; i < 3; i++ {
			time.Sleep(500 * time.Millisecond)
			fmt.Print(".")
		}
		fmt.Println()
		fmt.Println()
		
		// Show hint
		color.New(color.FgGreen).Println("💡 HINT:")
		fmt.Println(scenario.hint)
		fmt.Println()
		
		// Ask for solution
		color.New(color.FgYellow).Println("What would you do to fix this issue? (Press Enter when ready to see the solution)")
		fmt.Scanln()
		
		// Show solution
		color.New(color.FgGreen).Println("✅ SOLUTION:")
		fmt.Println(scenario.solution)
		fmt.Println()
		
		// Ask if they got it right
		color.New(color.FgYellow).Print("Did you get it right? (y/n): ")
		var answer string
		fmt.Scanln(&answer)
		
		if strings.ToLower(answer) == "y" {
			score++
			color.New(color.FgGreen).Println("Great job! +1 point")
		} else {
			color.New(color.FgCyan).Println("Keep learning! No worries.")
		}
		
		fmt.Println()
		color.New(color.FgYellow).Println("Press Enter to continue...")
		fmt.Scanln()
		fmt.Println()
		fmt.Println("------------------------------------------------")
		fmt.Println()
	}
	
	// Show final score
	fmt.Println()
	color.New(color.FgGreen, color.Bold).Println("🏆 GAME COMPLETE! 🏆")
	fmt.Println()
	color.New(color.FgYellow).Printf("Your score: %d/%d\n", score, totalScenarios)
	fmt.Println()
	
	// Show rating based on score
	percentage := float64(score) / float64(totalScenarios) * 100
	if percentage >= 90 {
		color.New(color.FgGreen, color.Bold).Println("Rating: Kubernetes Master! 🌟🌟🌟🌟🌟")
	} else if percentage >= 70 {
		color.New(color.FgGreen).Println("Rating: Kubernetes Expert! 🌟🌟🌟🌟")
	} else if percentage >= 50 {
		color.New(color.FgYellow).Println("Rating: Kubernetes Practitioner! 🌟🌟🌟")
	} else if percentage >= 30 {
		color.New(color.FgYellow).Println("Rating: Kubernetes Apprentice! 🌟🌟")
	} else {
		color.New(color.FgRed).Println("Rating: Kubernetes Novice! 🌟")
		color.New(color.FgCyan).Println("Don't worry! Keep practicing and learning!")
	}
	
	fmt.Println()
	color.New(color.FgCyan).Println("Thanks for playing! Run 'kubegpt game' again to play with different scenarios.")
	fmt.Println()
}

type gameScenario struct {
	title       string
	description string
	resources   []string
	error       string
	hint        string
	solution    string
}

func getScenarios() []gameScenario {
	return []gameScenario{
		{
			title:       "The Crashing Container",
			description: "A critical microservice keeps crashing and restarting in your production cluster.",
			resources: []string{
				"Deployment: payment-service",
				"Pod: payment-service-5d7f9b7b59-xvz2p",
				"Container: payment-api",
			},
			error:    "Pod Status: CrashLoopBackOff\nContainer payment-api has restarted 5 times in the last 10 minutes\nLast Exit Code: 137",
			hint:     "Exit code 137 typically indicates that the container was terminated due to an OOM (Out of Memory) kill. The container is likely trying to use more memory than its limit allows.",
			solution: "Increase the memory limit for the container in the deployment:\n\nkubectl patch deployment payment-service -n default --patch '\nspec:\n  template:\n    spec:\n      containers:\n      - name: payment-api\n        resources:\n          limits:\n            memory: \"512Mi\"\n          requests:\n            memory: \"256Mi\"\n'",
		},
		{
			title:       "The Pending Pod",
			description: "A new pod has been stuck in 'Pending' state for over 30 minutes.",
			resources: []string{
				"Pod: analytics-processor-1",
				"Node: worker-node-1, worker-node-2, worker-node-3",
			},
			error:    "Pod Status: Pending\nWarning: 0/3 nodes are available: 3 Insufficient memory.",
			hint:     "The scheduler can't find a node with enough available memory to schedule the pod. Either the pod is requesting too much memory, or your nodes are already heavily utilized.",
			solution: "Options:\n1. Reduce the memory request of the pod:\n   kubectl patch pod analytics-processor-1 -p '{\"spec\":{\"containers\":[{\"name\":\"analytics\",\"resources\":{\"requests\":{\"memory\":\"256Mi\"}}}]}}'\n\n2. Scale up your cluster to add more nodes:\n   kubectl scale --replicas=4 nodegroup worker-nodes\n\n3. Evict less important pods to free up resources:\n   kubectl drain worker-node-2 --ignore-daemonsets --delete-emptydir-data",
		},
		{
			title:       "The Unreachable Service",
			description: "Users are reporting that they can't access the frontend service, but the pods appear to be running.",
			resources: []string{
				"Service: frontend-svc",
				"Deployment: frontend",
				"Pods: frontend-7d9f7b7b59-abc1, frontend-7d9f7b7b59-def2, frontend-7d9f7b7b59-ghi3",
			},
			error:    "Service endpoint not responding\nAll pods show as Running\nkubectl get endpoints frontend-svc returns no endpoints",
			hint:     "The service selector might not be matching any pods. Check if the labels in the service selector match the labels on your pods.",
			solution: "The service selector labels don't match the pod labels. Fix the service:\n\nkubectl get pods -l app=frontend --show-labels\n# Notice the pods have label 'app=frontend-v2' instead of 'app=frontend'\n\nkubectl patch service frontend-svc -p '{\"spec\":{\"selector\":{\"app\":\"frontend-v2\"}}}'",
		},
		{
			title:       "The Failed Ingress",
			description: "Your new ingress resource isn't routing traffic to your service.",
			resources: []string{
				"Ingress: api-ingress",
				"Service: api-service",
				"Pods: api-deployment-5d7f9b7b59-abc1, api-deployment-5d7f9b7b59-def2",
			},
			error:    "Ingress shows address but returns 404 Not Found\nIngress controller logs show: \"service api-service not found\"",
			hint:     "The ingress resource might be referencing a service in the wrong namespace, or the service name might be incorrect.",
			solution: "The ingress is likely in a different namespace than the service. Fix by specifying the correct namespace:\n\nkubectl get svc --all-namespaces | grep api-service\n# Notice the service is in the 'backend' namespace\n\nkubectl patch ingress api-ingress -p '{\"spec\":{\"rules\":[{\"http\":{\"paths\":[{\"path\":\"/api\",\"pathType\":\"Prefix\",\"backend\":{\"service\":{\"name\":\"api-service\",\"port\":{\"number\":80}},\"namespace\":\"backend\"}}]}}]}}'\n\nOr move the ingress to the same namespace as the service:\n\nkubectl get ingress api-ingress -o yaml | sed 's/namespace: default/namespace: backend/' | kubectl apply -f -",
		},
		{
			title:       "The Secret Permission",
			description: "A pod can't start because it can't access a required secret.",
			resources: []string{
				"Pod: secure-app-1",
				"Secret: api-credentials",
				"ServiceAccount: app-service-account",
			},
			error:    "Pod Status: Error\nError: secrets \"api-credentials\" is forbidden: User \"system:serviceaccount:default:app-service-account\" cannot get resource \"secrets\" in API group \"\" in the namespace \"secure-namespace\"",
			hint:     "The pod's service account doesn't have permission to access the secret in the specified namespace. You need to create a Role and RoleBinding to grant access.",
			solution: "Create a Role and RoleBinding to allow the service account to access the secret:\n\nkubectl create role secret-reader --verb=get --resource=secrets --resource-name=api-credentials -n secure-namespace\n\nkubectl create rolebinding sa-secret-reader --role=secret-reader --serviceaccount=default:app-service-account -n secure-namespace\n\nThen restart the pod:\nkubectl delete pod secure-app-1\nkubectl get pods # Verify the new pod starts correctly",
		},
	}
}