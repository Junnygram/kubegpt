package cmd

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
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
  # Start a new game with default difficulty
  kubegpt game

  # Start a game with easy difficulty
  kubegpt game --difficulty easy

  # Start a game with hard difficulty
  kubegpt game --difficulty hard
`,
	Run: func(cmd *cobra.Command, args []string) {
		runGame()
	},
}

func init() {
	rootCmd.AddCommand(gameCmd)
	gameCmd.Flags().StringVar(&difficulty, "difficulty", "", "game difficulty (easy, medium, hard)")
}

func runGame() {
	printLogo()
	color.New(color.FgGreen, color.Bold).Println("🎮 KUBERNETES TROUBLESHOOTING CHALLENGE 🎮")
	fmt.Println()

	color.New(color.FgYellow).Println("Welcome to the Kubernetes Troubleshooting Challenge!")
	fmt.Println("Test your skills by diagnosing and fixing common Kubernetes issues.")
	fmt.Println("Amazon Q Developer will be your guide and judge.")
	fmt.Println()

	// Get username
	reader := bufio.NewReader(os.Stdin)
	color.New(color.FgCyan).Print("Enter your username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "Player"
	}

	// Get difficulty if not provided as flag
	selectedDifficulty := strings.ToLower(difficulty)
	if selectedDifficulty == "" {
		color.New(color.FgCyan).Println("\nSelect difficulty level:")
		color.New(color.FgGreen).Println("1. Easy (10 questions)")
		color.New(color.FgYellow).Println("2. Medium (20 questions)")
		color.New(color.FgRed).Println("3. Hard (30 questions)")
		color.New(color.FgCyan).Print("\nEnter your choice (1-3): ")

		diffChoice, _ := reader.ReadString('\n')
		diffChoice = strings.TrimSpace(diffChoice)

		switch diffChoice {
		case "1":
			selectedDifficulty = "easy"
		case "2":
			selectedDifficulty = "medium"
		case "3":
			selectedDifficulty = "hard"
		default:
			selectedDifficulty = "medium" // Default to medium if invalid input
		}
	}

	// Set number of questions based on difficulty
	numQuestions := 20 // Default medium
	switch selectedDifficulty {
	case "easy":
		numQuestions = 10
	case "medium":
		numQuestions = 20
	case "hard":
		numQuestions = 30
	}

	color.New(color.FgCyan).Printf("\nDifficulty: %s (%d questions)\n", strings.ToUpper(selectedDifficulty), numQuestions)
	fmt.Println()

	color.New(color.FgYellow).Println("Each question has a 30-second time limit.")
	color.New(color.FgYellow).Println("Choose the best answer from the options provided (A, B, or C).")
	color.New(color.FgYellow).Println("Press Enter when you're ready to start...")
	reader.ReadString('\n')

	// Start the game
	startGame(username, selectedDifficulty, numQuestions)
}

func startGame(username, difficulty string, numQuestions int) {
	allScenarios := getScenarios()

	// Limit scenarios based on difficulty
	scenarios := allScenarios
	if len(scenarios) > numQuestions {
		scenarios = scenarios[:numQuestions]
	}

	score := 0
	totalScenarios := len(scenarios)
	startTime := time.Now()

	reader := bufio.NewReader(os.Stdin)

	for i, scenario := range scenarios {
		color.New(color.FgMagenta, color.Bold).Printf("QUESTION %d/%d: %s\n", i+1, totalScenarios, scenario.title)
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
			time.Sleep(300 * time.Millisecond)
			fmt.Print(".")
		}
		fmt.Println()
		fmt.Println()

		// Show hint
		color.New(color.FgGreen).Println("💡 HINT:")
		fmt.Println(scenario.hint)
		fmt.Println()

		// Show multiple choice options
		color.New(color.FgYellow).Println("What would you do to fix this issue?")
		color.New(color.FgWhite).Printf("A) %s\n", scenario.optionA)
		color.New(color.FgWhite).Printf("B) %s\n", scenario.optionB)
		color.New(color.FgWhite).Printf("C) %s\n", scenario.optionC)
		fmt.Println()

		// Start timer for this question
		questionStart := time.Now()
		timeLeft := 30 * time.Second

		// Create a channel for timeout
		timeout := make(chan bool, 1)
		go func() {
			time.Sleep(timeLeft)
			timeout <- true
		}()

		// Ask for answer
		color.New(color.FgYellow).Printf("Time remaining: 30s - Enter your answer (A, B, or C): ")

		// Wait for either user input or timeout
		var userAnswer string
		inputCh := make(chan string)
		go func() {
			answer, _ := reader.ReadString('\n')
			inputCh <- strings.TrimSpace(answer)
		}()

		// Wait for either input or timeout
		select {
		case userAnswer = <-inputCh:
			// User provided input before timeout
			elapsed := time.Since(questionStart)
			timeLeft = 30*time.Second - elapsed
			if timeLeft < 0 {
				timeLeft = 0
			}
			color.New(color.FgYellow).Printf("Time remaining: %.0fs\n", timeLeft.Seconds())
		case <-timeout:
			// Time's up
			color.New(color.FgRed).Println("Time's up!")
			userAnswer = ""
		}

		// Check answer

		userAnswer = strings.ToUpper(userAnswer)
		if userAnswer == scenario.correctAnswer {

			score++
			color.New(color.FgGreen).Println("✅ Correct! +1 point")
		} else {
			color.New(color.FgRed).Println("❌ Incorrect!")
		}

		// Show solution
		color.New(color.FgGreen).Println("\n📝 SOLUTION EXPLANATION:")
		fmt.Println(scenario.solution)
		fmt.Println()

		// Show correct answer
		color.New(color.FgGreen).Printf("The correct answer was: %s\n", scenario.correctAnswer)
		fmt.Println()

		color.New(color.FgYellow).Println("Press Enter to continue...")
		reader.ReadString('\n')
		fmt.Println()
		fmt.Println("------------------------------------------------")
		fmt.Println()
	}

	// Calculate total time
	totalTime := time.Since(startTime)

	// Show final score
	fmt.Println()
	color.New(color.FgGreen, color.Bold).Println("🏆 GAME COMPLETE! 🏆")
	fmt.Println()
	color.New(color.FgYellow).Printf("Player: %s\n", username)
	color.New(color.FgYellow).Printf("Difficulty: %s\n", strings.ToUpper(difficulty))
	color.New(color.FgYellow).Printf("Score: %d/%d\n", score, totalScenarios)
	color.New(color.FgYellow).Printf("Time: %.1f minutes\n", totalTime.Minutes())
	fmt.Println()

	// Show rating based on score
	percentage := float64(score) / float64(totalScenarios) * 100
	var rating string

	if percentage >= 90 {
		rating = "Kubernetes Master! 🌟🌟🌟🌟🌟"
		color.New(color.FgGreen, color.Bold).Println("Rating: " + rating)
	} else if percentage >= 70 {
		rating = "Kubernetes Expert! 🌟🌟🌟🌟"
		color.New(color.FgGreen).Println("Rating: " + rating)
	} else if percentage >= 50 {
		rating = "Kubernetes Practitioner! 🌟🌟🌟"
		color.New(color.FgYellow).Println("Rating: " + rating)
	} else if percentage >= 30 {
		rating = "Kubernetes Apprentice! 🌟🌟"
		color.New(color.FgYellow).Println("Rating: " + rating)
	} else {
		rating = "Kubernetes Novice! 🌟"
		color.New(color.FgRed).Println("Rating: " + rating)
		color.New(color.FgCyan).Println("Don't worry! Keep practicing and learning!")
	}

	// Save results to CSV
	saveGameResult(username, difficulty, score, totalScenarios, totalTime.Minutes(), rating)

	fmt.Println()
	color.New(color.FgCyan).Println("Thanks for playing! Run 'kubegpt game' again to play with different scenarios.")
	fmt.Println()
}

// saveGameResult saves the game result to a CSV file
func saveGameResult(username, difficulty string, score, total int, timeMinutes float64, rating string) {
	// Use the current directory instead of home directory
	resultsFile := "game_results.csv"
	color.New(color.FgGreen).Printf("Saving results to: %s\n", resultsFile)

	var writer *csv.Writer

	fileExists := true
	if _, err := os.Stat(resultsFile); os.IsNotExist(err) {
		fileExists = false
	}

	if !fileExists {
		// Create file with headers
		file, err := os.Create(resultsFile)
		if err != nil {
			color.New(color.FgRed).Printf("Error creating results file: %v\n", err)
			return
		}
		defer file.Close()

		writer = csv.NewWriter(file)
		err = writer.Write([]string{"Timestamp", "Username", "Difficulty", "Score", "Total", "Percentage", "Time (minutes)", "Rating"})
		if err != nil {
			color.New(color.FgRed).Printf("Error writing headers: %v\n", err)
			return
		}
		color.New(color.FgGreen).Println("Created new results file with headers")
	} else {
		// Append to existing file
		file, err := os.OpenFile(resultsFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			color.New(color.FgRed).Printf("Error opening results file: %v\n", err)
			return
		}
		defer file.Close()

		writer = csv.NewWriter(file)
		color.New(color.FgGreen).Println("Opened existing results file for appending")
	}

	// Write result
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	percentage := float64(score) / float64(total) * 100

	record := []string{
		timestamp,
		username,
		difficulty,
		fmt.Sprintf("%d", score),
		fmt.Sprintf("%d", total),
		fmt.Sprintf("%.1f%%", percentage),
		fmt.Sprintf("%.1f", timeMinutes),
		rating,
	}

	err := writer.Write(record)
	if err != nil {
		color.New(color.FgRed).Printf("Error writing record: %v\n", err)
		return
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		color.New(color.FgRed).Printf("Error flushing writer: %v\n", err)
		return
	}

	color.New(color.FgGreen).Printf("\nYour result has been saved to %s!\n", resultsFile)
}

type gameScenario struct {
	title         string
	description   string
	resources     []string
	error         string
	hint          string
	optionA       string
	optionB       string
	optionC       string
	correctAnswer string
	solution      string
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
			error:         "Pod Status: CrashLoopBackOff\nContainer payment-api has restarted 5 times in the last 10 minutes\nLast Exit Code: 137",
			hint:          "Exit code 137 typically indicates that the container was terminated due to an OOM (Out of Memory) kill. The container is likely trying to use more memory than its limit allows.",
			optionA:       "Restart the pod using kubectl delete pod payment-service-5d7f9b7b59-xvz2p",
			optionB:       "Increase the memory limit for the container in the deployment",
			optionC:       "Check the container logs for application errors",
			correctAnswer: "B",
			solution:      "Exit code 137 indicates an Out of Memory (OOM) kill. The container is being terminated because it's trying to use more memory than its configured limit. Increasing the memory limit will allow the container to use more memory and prevent it from being killed.\n\nThe correct approach is to increase the memory limit using:\n\nkubectl patch deployment payment-service -n default --patch '\nspec:\n  template:\n    spec:\n      containers:\n      - name: payment-api\n        resources:\n          limits:\n            memory: \"512Mi\"\n          requests:\n            memory: \"256Mi\"\n'",
		},
		{
			title:       "The Pending Pod",
			description: "A new pod has been stuck in 'Pending' state for over 30 minutes.",
			resources: []string{
				"Pod: analytics-processor-1",
				"Node: worker-node-1, worker-node-2, worker-node-3",
			},
			error:         "Pod Status: Pending\nWarning: 0/3 nodes are available: 3 Insufficient memory.",
			hint:          "The scheduler can't find a node with enough available memory to schedule the pod. Either the pod is requesting too much memory, or your nodes are already heavily utilized.",
			optionA:       "Delete and recreate the pod",
			optionB:       "Scale up the cluster by adding more nodes",
			optionC:       "Reduce the memory request of the pod",
			correctAnswer: "C",
			solution:      "The error message indicates that none of the nodes have sufficient memory to schedule the pod. While scaling up the cluster (option B) would work, the most immediate and resource-efficient solution is to reduce the memory request of the pod if possible.\n\nThe correct approach is to reduce the memory request using:\n\nkubectl patch pod analytics-processor-1 -p '{\"spec\":{\"containers\":[{\"name\":\"analytics\",\"resources\":{\"requests\":{\"memory\":\"256Mi\"}}}]}}'",
		},
		{
			title:       "The Unreachable Service",
			description: "Users are reporting that they can't access the frontend service, but the pods appear to be running.",
			resources: []string{
				"Service: frontend-svc",
				"Deployment: frontend",
				"Pods: frontend-7d9f7b7b59-abc1, frontend-7d9f7b7b59-def2, frontend-7d9f7b7b59-ghi3",
			},
			error:         "Service endpoint not responding\nAll pods show as Running\nkubectl get endpoints frontend-svc returns no endpoints",
			hint:          "The service selector might not be matching any pods. Check if the labels in the service selector match the labels on your pods.",
			optionA:       "Restart the pods in the deployment",
			optionB:       "Update the service selector to match the pod labels",
			optionC:       "Create a new service with the correct port",
			correctAnswer: "B",
			solution:      "The issue is that the service selector doesn't match the labels on the pods. This is evident from the fact that 'kubectl get endpoints frontend-svc returns no endpoints' even though the pods are running.\n\nThe correct approach is to update the service selector to match the pod labels:\n\nkubectl get pods --show-labels\n# Notice the pods have label 'app=frontend-v2' instead of 'app=frontend'\n\nkubectl patch service frontend-svc -p '{\"spec\":{\"selector\":{\"app\":\"frontend-v2\"}}}'",
		},
		{
			title:       "The Failed Ingress",
			description: "Your new ingress resource isn't routing traffic to your service.",
			resources: []string{
				"Ingress: api-ingress",
				"Service: api-service",
				"Pods: api-deployment-5d7f9b7b59-abc1, api-deployment-5d7f9b7b59-def2",
			},
			error:         "Ingress shows address but returns 404 Not Found\nIngress controller logs show: \"service api-service not found\"",
			hint:          "The ingress resource might be referencing a service in the wrong namespace, or the service name might be incorrect.",
			optionA:       "Update the ingress to specify the correct namespace for the service",
			optionB:       "Recreate the service with the correct name",
			optionC:       "Restart the ingress controller",
			correctAnswer: "A",
			solution:      "The error message \"service api-service not found\" suggests that the ingress controller can't find the service, even though it exists. This typically happens when the ingress is in a different namespace than the service.\n\nThe correct approach is to update the ingress to specify the correct namespace:\n\nkubectl get svc --all-namespaces | grep api-service\n# Notice the service is in the 'backend' namespace\n\nkubectl patch ingress api-ingress -p '{\"spec\":{\"rules\":[{\"http\":{\"paths\":[{\"path\":\"/api\",\"pathType\":\"Prefix\",\"backend\":{\"service\":{\"name\":\"api-service\",\"port\":{\"number\":80}},\"namespace\":\"backend\"}}]}}]}}'",
		},
		{
			title:       "The Secret Permission",
			description: "A pod can't start because it can't access a required secret.",
			resources: []string{
				"Pod: secure-app-1",
				"Secret: api-credentials",
				"ServiceAccount: app-service-account",
			},
			error:         "Pod Status: Error\nError: secrets \"api-credentials\" is forbidden: User \"system:serviceaccount:default:app-service-account\" cannot get resource \"secrets\" in API group \"\" in the namespace \"secure-namespace\"",
			hint:          "The pod's service account doesn't have permission to access the secret in the specified namespace. You need to create a Role and RoleBinding to grant access.",
			optionA:       "Create a Role and RoleBinding to grant the service account access to the secret",
			optionB:       "Move the secret to the same namespace as the pod",
			optionC:       "Use a different service account with more permissions",
			correctAnswer: "A",
			solution:      "The error message indicates a permission issue - the service account doesn't have permission to access the secret in the specified namespace. The proper Kubernetes way to handle this is to create a Role that allows access to the specific secret, and a RoleBinding that grants this role to the service account.\n\nThe correct approach is to create a Role and RoleBinding:\n\nkubectl create role secret-reader --verb=get --resource=secrets --resource-name=api-credentials -n secure-namespace\n\nkubectl create rolebinding sa-secret-reader --role=secret-reader --serviceaccount=default:app-service-account -n secure-namespace",
		},
		{
			title:       "The Evicted Pod",
			description: "Your monitoring system alerts that several pods have been evicted from your production cluster.",
			resources: []string{
				"Pods: cache-redis-0, cache-redis-1, cache-redis-2",
				"Node: worker-node-4",
			},
			error:         "Pod Status: Evicted\nReason: The node was low on resource: ephemeral-storage.\nNode Status: Ready, Pressure: DiskPressure",
			hint:          "Pod evictions often occur when a node is under resource pressure. In this case, the node is experiencing disk pressure due to low ephemeral storage.",
			optionA:       "Immediately cordon and drain the node for maintenance",
			optionB:       "Clean up unused images and containers on the node",
			optionC:       "Increase the toleration time for disk pressure",
			correctAnswer: "B",
			solution:      "The node is experiencing disk pressure due to low ephemeral storage. While cordoning the node would prevent new pods from being scheduled, it doesn't solve the underlying issue. The most immediate solution is to free up disk space by cleaning up unused images and containers.\n\nThe correct approach is to clean up the node:\n\nkubectl debug node/worker-node-4 -it -- chroot /host crictl rmi --prune\nkubectl debug node/worker-node-4 -it -- chroot /host crictl rm $(crictl ps -a -q --state exited)\n\nAfter freeing up space, the evicted pods need to be recreated:\nkubectl get pod cache-redis-0 cache-redis-1 cache-redis-2 -o name | xargs kubectl delete\n# StatefulSet controller will automatically create new pods",
		},
		{
			title:       "The Network Policy Blockade",
			description: "After implementing network policies, your frontend service can't communicate with your backend API.",
			resources: []string{
				"Pod: frontend-app-7d9cb6b88c-xpl43",
				"Pod: backend-api-5f7d8b9c67-ztq29",
				"Service: backend-api-svc",
				"NetworkPolicy: backend-restrict",
			},
			error:         "curl: (7) Failed to connect to backend-api-svc port 8080: Connection timed out\nNetworkPolicy 'backend-restrict' is active",
			hint:          "The network policy may be too restrictive. It needs to explicitly allow traffic from the frontend namespace/pods to the backend pods.",
			optionA:       "Temporarily delete the network policy to restore communication",
			optionB:       "Add a new ingress rule to the network policy to allow traffic from frontend pods",
			optionC:       "Create a service mesh to bypass network policies",
			correctAnswer: "B",
			solution:      "Network policies in Kubernetes are deny-by-default, meaning once a policy is applied to a pod, all traffic not explicitly allowed is denied. The correct approach is to modify the network policy to allow traffic from the frontend pods.\n\nThe correct solution is to add an ingress rule to the network policy:\n\nkubectl patch networkpolicy backend-restrict -n backend --type=json -p='[{\"op\": \"add\", \"path\": \"/spec/ingress/-\", \"value\": {\"from\": [{\"namespaceSelector\": {\"matchLabels\": {\"name\": \"frontend\"}}, \"podSelector\": {\"matchLabels\": {\"app\": \"frontend-app\"}}}], \"ports\": [{\"port\": 8080, \"protocol\": \"TCP\"}]}}]'",
		},

		{
			title:       "The Evicted Pod",
			description: "Your monitoring system alerts that several pods have been evicted from your production cluster.",
			resources: []string{
				"Pods: cache-redis-0, cache-redis-1, cache-redis-2",
				"Node: worker-node-4",
			},
			error:         "Pod Status: Evicted\nReason: The node was low on resource: ephemeral-storage.\nNode Status: Ready, Pressure: DiskPressure",
			hint:          "Pod evictions often occur when a node is under resource pressure. In this case, the node is experiencing disk pressure due to low ephemeral storage.",
			optionA:       "Immediately cordon and drain the node for maintenance",
			optionB:       "Clean up unused images and containers on the node",
			optionC:       "Increase the toleration time for disk pressure",
			correctAnswer: "B",
			solution:      "The node is experiencing disk pressure due to low ephemeral storage. While cordoning the node would prevent new pods from being scheduled, it doesn't solve the underlying issue. The most immediate solution is to free up disk space by cleaning up unused images and containers.\n\nThe correct approach is to clean up the node:\n\nkubectl debug node/worker-node-4 -it -- chroot /host crictl rmi --prune\nkubectl debug node/worker-node-4 -it -- chroot /host crictl rm $(crictl ps -a -q --state exited)\n\nAfter freeing up space, the evicted pods need to be recreated:\nkubectl get pod cache-redis-0 cache-redis-1 cache-redis-2 -o name | xargs kubectl delete\n# StatefulSet controller will automatically create new pods",
		},
		{
			title:       "The Network Policy Blockade",
			description: "After implementing network policies, your frontend service can't communicate with your backend API.",
			resources: []string{
				"Pod: frontend-app-7d9cb6b88c-xpl43",
				"Pod: backend-api-5f7d8b9c67-ztq29",
				"Service: backend-api-svc",
				"NetworkPolicy: backend-restrict",
			},
			error:         "curl: (7) Failed to connect to backend-api-svc port 8080: Connection timed out\nNetworkPolicy 'backend-restrict' is active",
			hint:          "The network policy may be too restrictive. It needs to explicitly allow traffic from the frontend namespace/pods to the backend pods.",
			optionA:       "Temporarily delete the network policy to restore communication",
			optionB:       "Add a new ingress rule to the network policy to allow traffic from frontend pods",
			optionC:       "Change the backend service to use NodePort instead of ClusterIP",
			correctAnswer: "B",
			solution:      "The network policy is blocking traffic from the frontend to the backend service. While deleting the policy would work, it's not the correct approach as it would remove all protection. Changing the service type doesn't address the network policy issue.\n\nThe correct approach is to update the network policy to allow traffic from the frontend:\n\nkubectl get pod frontend-app-7d9cb6b88c-xpl43 --show-labels\n# Take note of the labels, e.g., app=frontend\n\nkubectl patch networkpolicy backend-restrict -n backend-namespace --patch '\nspec:\n  ingress:\n  - from:\n    - namespaceSelector:\n        matchLabels:\n          name: frontend-namespace\n      podSelector:\n        matchLabels:\n          app: frontend\n    ports:\n    - protocol: TCP\n      port: 8080\n'",
		},
		{
			title:       "The Resource Quota Limit",
			description: "Your team can't deploy a new application because the resource quota has been exceeded.",
			resources: []string{
				"Namespace: development",
				"ResourceQuota: dev-quota",
				"Deployment: new-application",
			},
			error:         "Error: failed quota: dev-quota: must specify limits.memory\nCurrent Usage: requests.cpu: 800m/1000m, limits.cpu: 1200m/2000m, requests.memory: 1.2Gi/2Gi",
			hint:          "The resource quota requires that all containers specify memory limits, and you're also approaching the CPU request limit for the namespace.",
			optionA:       "Increase the resource quota limits for the namespace",
			optionB:       "Add memory limits to the new application deployment",
			optionC:       "Deploy the application to a different namespace with no quota",
			correctAnswer: "B",
			solution:      "The error indicates that the resource quota requires all containers to specify memory limits, which are missing from the new deployment. Moving to another namespace or increasing the quota would work but isn't the correct approach - you should comply with the quota requirements.\n\nThe correct approach is to add memory limits to the deployment:\n\nkubectl patch deployment new-application -n development --patch '\nspec:\n  template:\n    spec:\n      containers:\n      - name: app-container\n        resources:\n          limits:\n            memory: \"256Mi\"\n            cpu: \"200m\"\n          requests:\n            memory: \"128Mi\"\n            cpu: \"100m\"\n'",
		},
		{
			title:       "The Persistent Volume Claim Stuck",
			description: "A newly created pod is stuck in 'ContainerCreating' state because its PVC is pending.",
			resources: []string{
				"Pod: database-0",
				"PVC: data-database-0",
				"StorageClass: standard",
			},
			error:         "Pod Status: ContainerCreating\nPVC Status: Pending\nWarning: ProvisioningFailed: Failed to provision volume: connection error: desc = \"transport: Error while dialing dial tcp: lookup storage-provisioner on 10.96.0.10:53: no such host\"",
			hint:          "The storage provisioner service is not available or misconfigured. Check if the storage provider is running correctly.",
			optionA:       "Manually create a PV and patch the PVC to bind to it",
			optionB:       "Fix the storage provisioner service or use a different storage class",
			optionC:       "Delete and recreate the pod with a hostPath volume instead",
			correctAnswer: "B",
			solution:      "The error indicates that the storage provisioner service cannot be reached. This is a cluster configuration issue that needs to be addressed at the storage provisioner level.\n\nThe correct approach is to fix the storage provisioner or use a different storage class:\n\n# First check available storage classes\nkubectl get sc\n\n# If another storage class is available, patch the PVC to use it:\nkubectl patch pvc data-database-0 --patch '{\"spec\":{\"storageClassName\":\"alternative-storage-class\"}}'\n\n# If you need to fix the provisioner:\nkubectl get pods -n kube-system | grep storage-provisioner\n# Check the status and logs of the provisioner\nkubectl describe pod storage-provisioner-pod-name -n kube-system\nkubectl logs storage-provisioner-pod-name -n kube-system",
		},
		{
			title:       "The ConfigMap Confusion",
			description: "After updating a ConfigMap, your application is still using the old configuration values.",
			resources: []string{
				"ConfigMap: app-config",
				"Deployment: web-application",
				"Pods: web-application-6d9fb7c484-abc1, web-application-6d9fb7c484-def2",
			},
			error:         "Application logs: Using database host: db-old.example.com\nExpected value in ConfigMap: db-new.example.com\nConfigMap last updated: 2 hours ago",
			hint:          "Kubernetes doesn't automatically update existing pods when a ConfigMap changes. You need to trigger a rollout of the deployment.",
			optionA:       "Delete the pods manually to force recreation",
			optionB:       "Restart the deployment with kubectl rollout restart",
			optionC:       "Add an annotation to the pod template to force an update",
			correctAnswer: "B",
			solution:      "When a ConfigMap is updated, existing pods don't automatically receive the new values. You need to trigger a deployment rollout to create new pods with the updated ConfigMap.\n\nThe correct approach is to restart the deployment:\n\nkubectl rollout restart deployment web-application\n\n# Monitor the rollout\nkubectl rollout status deployment web-application\n\n# Verify the new pods have the updated config\nkubectl exec -it web-application-[new-pod-id] -- env | grep DB_HOST",
		},
		{
			title:       "The Init Container Blocker",
			description: "A pod is stuck in 'Init:CrashLoopBackOff' state and can't start its main containers.",
			resources: []string{
				"Pod: backend-api-6f7c9d8e5a-zxq43",
				"Deployment: backend-api",
			},
			error:         "Pod Status: Init:CrashLoopBackOff\nInit Container: check-db-ready\nInit Container Logs: error: could not connect to database at db-service:5432: connection refused",
			hint:          "The init container is failing because it can't connect to a database service that should be running before this pod starts.",
			optionA:       "Remove the init container from the pod spec",
			optionB:       "Ensure the database service is running and accessible",
			optionC:       "Increase the failure threshold for the init container",
			correctAnswer: "B",
			solution:      "The init container is designed to ensure dependencies are ready before the main container starts. It's failing because the database it's checking for isn't available. Removing the init container would break this dependency check.\n\nThe correct approach is to verify and fix the database service:\n\nkubectl get pods -l app=database\n# If not running, start the database\nkubectl get svc db-service\n# Verify the service exists and points to the right pods\n\n# If the database is in another namespace, make sure it's accessible from the pod's namespace\nkubectl get endpoints db-service\n# Check if the service has endpoints\n\n# Once the database is accessible, the init container will succeed and the pod will start",
		},
		{
			title:       "The Probe Failure",
			description: "Your application pod keeps restarting despite the application running correctly.",
			resources: []string{
				"Pod: web-server-7c8d9e6f5d-rtb42",
				"Deployment: web-server",
			},
			error:         "Pod Status: Running, but restarts: 7\nWarning: Liveness probe failed: HTTP probe failed with statuscode: 503",
			hint:          "The pod's liveness probe is failing, causing Kubernetes to restart the container even though the application might be functioning correctly.",
			optionA:       "Remove the liveness probe from the container",
			optionB:       "Adjust the liveness probe to match how the application reports health",
			optionC:       "Increase the initialDelaySeconds for the liveness probe",
			correctAnswer: "B",
			solution:      "The liveness probe is failing because it's not correctly configured to match how the application reports its health. Simply removing the probe would mean losing the automatic recovery benefit, and just increasing the delay doesn't fix the fundamental mismatch.\n\nThe correct approach is to adjust the liveness probe configuration:\n\nkubectl get deployment web-server -o yaml > web-server.yaml\n# Edit the file to update the liveness probe path, port, or headers to match the application's health endpoint\n\nkubectl patch deployment web-server --patch '\nspec:\n  template:\n    spec:\n      containers:\n      - name: web-server\n        livenessProbe:\n          httpGet:\n            path: /healthz\n            port: 8080\n          initialDelaySeconds: 15\n          periodSeconds: 10\n'",
		},
		{
			title:       "The Taints and Tolerations Challenge",
			description: "After upgrading your cluster, new pods won't schedule on specific nodes they previously used.",
			resources: []string{
				"Nodes: worker-node-5, worker-node-6, worker-node-7",
				"Deployment: gpu-processor",
				"Pods: gpu-processor-5d6f7g8h9i-abc1 (Pending)",
			},
			error:         "Pod Status: Pending\nWarning: 0/3 nodes are available: 3 node(s) had taint {dedicated=gpu:NoSchedule}, that the pod didn't tolerate.",
			hint:          "The nodes have been tainted to prevent general workloads from using these specialized GPU nodes. You need to add a matching toleration to your deployment.",
			optionA:       "Remove the taints from the nodes",
			optionB:       "Add tolerations to the deployment to match the node taints",
			optionC:       "Use node selectors instead of taints and tolerations",
			correctAnswer: "B",
			solution:      "The nodes have been tainted to reserve them for specific workloads (GPU in this case). Your pods need to have matching tolerations to be scheduled on these nodes. Removing the taints would allow any pods to schedule on these nodes, defeating the purpose of reserving them.\n\nThe correct approach is to add tolerations to the deployment:\n\nkubectl patch deployment gpu-processor --patch '\nspec:\n  template:\n    spec:\n      tolerations:\n      - key: \"dedicated\"\n        operator: \"Equal\"\n        value: \"gpu\"\n        effect: \"NoSchedule\"\n      nodeSelector:\n        gpu: \"true\"\n'",
		},
		{
			title:       "The ImagePullBackOff Mystery",
			description: "Your new deployment is failing with ImagePullBackOff errors.",
			resources: []string{
				"Deployment: customer-portal",
				"Pod: customer-portal-7d8e9f6c5b-jhk34",
				"Secret: registry-credentials",
			},
			error:         "Pod Status: ImagePullBackOff\nWarning: Failed to pull image \"private-registry.company.com/customer-portal:latest\": rpc error: code = Unknown desc = Error response from daemon: pull access denied, repository does not exist or may require authentication",
			hint:          "The pod can't pull the image because it either doesn't exist, or more likely, requires authentication to the private registry.",
			optionA:       "Use a different image tag from a public registry",
			optionB:       "Configure an imagePullSecret for the deployment",
			optionC:       "Add the registry URL to the insecureRegistries list in the Docker daemon config",
			correctAnswer: "B",
			solution:      "The error shows that the pod can't pull the image from your private registry due to authentication issues. You need to configure the deployment to use a secret containing registry credentials.\n\nThe correct approach is to create and use an imagePullSecret:\n\n# First, check if the secret exists\nkubectl get secret registry-credentials\n\n# If not, create it\nkubectl create secret docker-registry registry-credentials \\\n  --docker-server=private-registry.company.com \\\n  --docker-username=your-username \\\n  --docker-password=your-password\n\n# Update the deployment to use the secret\nkubectl patch deployment customer-portal --patch '\nspec:\n  template:\n    spec:\n      imagePullSecrets:\n      - name: registry-credentials\n'",
		},
		{
			title:       "The Horizontal Pod Autoscaler Dilemma",
			description: "Your application isn't scaling up during high load despite having an HPA configured.",
			resources: []string{
				"Deployment: webshop-frontend",
				"HPA: webshop-frontend-hpa",
				"Metrics Server: Running",
			},
			error:         "HPA Status: <unknown>/80%\nWarning: failed to get cpu utilization: unable to get metrics for resource cpu: no metrics returned from resource metrics API",
			hint:          "The HPA can't get the metrics needed to make scaling decisions. This could be due to missing resource requests in the deployment or issues with the metrics server.",
			optionA:       "Manually scale the deployment instead of using HPA",
			optionB:       "Add resource requests and limits to the deployment pods",
			optionC:       "Switch from CPU-based scaling to custom metrics",
			correctAnswer: "B",
			solution:      "The HPA needs pod resource requests defined to calculate utilization percentages. Without these, it can't determine when to scale. The error indicates it can't get CPU metrics, which is likely because CPU requests aren't defined in the pod spec.\n\nThe correct approach is to add resource requests to the deployment:\n\nkubectl patch deployment webshop-frontend --patch '\nspec:\n  template:\n    spec:\n      containers:\n      - name: webshop-app\n        resources:\n          requests:\n            cpu: \"200m\"\n            memory: \"256Mi\"\n          limits:\n            cpu: \"500m\"\n            memory: \"512Mi\"\n'",
		},
		{
			title:       "The Role-Based Access Control (RBAC) Issue",
			description: "Your CI/CD pipeline suddenly can't deploy updates to your cluster.",
			resources: []string{
				"ServiceAccount: ci-deployer",
				"ClusterRole: deployment-manager",
				"ClusterRoleBinding: ci-deployment-binding",
			},
			error:         "Error: deployments.apps is forbidden: User \"system:serviceaccount:ci-cd:ci-deployer\" cannot update resource \"deployments\" in API group \"apps\" in the namespace \"production\"",
			hint:          "The service account doesn't have permission to update deployments in the production namespace. Check your RBAC configuration.",
			optionA:       "Use a different service account with cluster-admin privileges",
			optionB:       "Update the ClusterRole to include the missing permissions",
			optionC:       "Create a namespace-specific Role and RoleBinding for the production namespace",
			correctAnswer: "C",
			solution:      "The error shows that the CI service account can't update deployments in the production namespace. While updating the ClusterRole would work, it's better to follow the principle of least privilege and create a namespace-specific Role for production access.\n\nThe correct approach is to create a Role and RoleBinding for the production namespace:\n\nkubectl create role deployment-manager --verb=get,list,watch,create,update,patch,delete --resource=deployments.apps -n production\n\nkubectl create rolebinding ci-deployment-binding --role=deployment-manager --serviceaccount=ci-cd:ci-deployer -n production\n\n# Verify the permissions\nkubectl auth can-i update deployments.apps --as=system:serviceaccount:ci-cd:ci-deployer -n production",
		},
	}
}
