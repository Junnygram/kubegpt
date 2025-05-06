# KubeGPT CLI

KubeGPT is an AI-powered Kubernetes troubleshooting assistant that helps DevOps engineers and SREs diagnose and fix issues in their Kubernetes clusters.

![KubeGPT Logo](https://via.placeholder.com/800x200?text=KubeGPT+CLI)

## Features

- **AI-Powered Analysis**: Analyze Kubernetes logs, error messages, and YAML configurations using Amazon Q Developer
- **Cluster Diagnostics**: Automatically identify unhealthy pods, failed events, and misconfigured deployments
- **Smart Troubleshooting**: Get explanations of issues and suggested kubectl commands to fix them
- **Fix Generation**: Generate YAML patches to fix common Kubernetes issues
- **Comprehensive Reporting**: Generate reports in terminal, markdown, or send to Slack
- **Interactive Learning Game**: Play a Kubernetes troubleshooting game to improve your skills
- **Resource Transformation**: Convert Kubernetes YAML to Terraform, Pulumi, AWS CDK, and more

## Installation

### Prerequisites

- Go 1.21 or later
- Kubernetes cluster and kubectl configured
- Amazon Q Developer CLI installed (`pip install amazon-q-developer-cli`)

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/kubegpt.git
   cd kubegpt
   ```

2. Build the binary:
   ```bash
   go build -o kubegpt
   ```

3. Move the binary to your PATH:
   ```bash
   sudo mv kubegpt /usr/local/bin/
   ```

## Usage

### Diagnosing Cluster Issues

Run a diagnostic on your current namespace:

```bash
kubegpt diagnose
```

Diagnose issues in a specific namespace:

```bash
kubegpt diagnose --namespace monitoring
```

Generate YAML patches to fix issues:

```bash
kubegpt diagnose --fix
```

### Explaining Kubernetes Errors

Explain a specific error message:

```bash
kubegpt explain "CrashLoopBackOff: container exited with code 1"
```

Analyze logs from kubectl:

```bash
kubectl logs my-pod | kubegpt explain
```

Analyze a YAML configuration:

```bash
kubectl get deployment my-deployment -o yaml | kubegpt explain
```

### Generating Reports

Generate a cluster health report:

```bash
kubegpt report
```

Generate a report for all namespaces:

```bash
kubegpt report --all-namespaces
```

Save the report as markdown:

```bash
kubegpt report --output markdown --file cluster-health.md
```

Send the report to Slack:

```bash
kubegpt report --output slack --slack-webhook https://hooks.slack.com/services/...
```

### Playing the Kubernetes Troubleshooting Game

Start a new game with default difficulty:

```bash
kubegpt game
```

Start a game with easy difficulty:

```bash
kubegpt game --difficulty easy
```

Start a challenging game:

```bash
kubegpt game --difficulty hard
```

### Transforming Kubernetes Resources

Convert Kubernetes YAML to Terraform:

```bash
kubegpt transform --target-lang terraform -f deployment.yaml -o deployment.tf
```

Convert Kubernetes YAML to Pulumi Python:

```bash
kubegpt transform --target-lang pulumi-py -f deployment.yaml -o pulumi_app.py
```

Convert between YAML and JSON:

```bash
kubegpt transform --input-format yaml --output-format json -f deployment.yaml -o deployment.json
```

## Configuration

KubeGPT uses your kubectl configuration by default. You can specify a different kubeconfig file using the `--kubeconfig` flag:

```bash
kubegpt --kubeconfig=/path/to/kubeconfig diagnose
```

## Amazon Q Developer Integration

KubeGPT uses Amazon Q Developer CLI to analyze Kubernetes issues. Make sure you have it installed and configured:

```bash
pip install amazon-q-developer-cli
amazon-q configure
```

For development and testing without Amazon Q, you can use the mock mode:

```bash
export KUBEGPT_MOCK_AI=true
kubegpt diagnose
```

## Examples

### Diagnosing a CrashLoopBackOff Issue

```bash
$ kubegpt diagnose --pods-only
```

Output:
```
 _    _    _            _____  _____  _______
| |  / |  | |          / ____||  __ \|__   __|
| | / /| |_| |__   ___| |  __ | |__) |  | |   
| |/ / | __| '_ \ / _ \ | |_ ||  ___/   | |   
|   <  | |_| |_) |  __/ |__| || |       | |   
|_|\_\  \__|_.__/ \___|\_____||_|       |_|   
                                             
AI-powered Kubernetes troubleshooting assistant
--------------------------------------------

Diagnosing issues in namespace: default

Checking pods...
Found 2 unhealthy pods

Analyzing issues with Amazon Q...
Analyzing pod frontend-6d4cf56db6-abc12...
Analyzing pod database-5c7cf34fd8-ghi56...

Diagnostic Results for Namespace: default
Time: Wed, 15 May 2024 14:23:45 UTC

Summary:
- Unhealthy Pods: 2
- Failed Events: 0
- Misconfigured Deployments: 0
- Service Issues: 0

Unhealthy Pods:

[1] Pod: frontend-6d4cf56db6-abc12
    Status: Running
    Container Issues:
    - frontend: Running (Restarts: 3)
      Reason: Unhealthy
      Message: Readiness probe failed: HTTP probe failed with statuscode: 500

    Analysis:
    The frontend pod is experiencing readiness probe failures with HTTP status code 500.
    
    Likely causes:
    1. The application is not properly handling health check requests
    2. The application is failing to connect to a backend service
    3. The readiness probe configuration might be incorrect
    
    Recommended solutions:
    1. Check the application logs for errors
    2. Verify connectivity to dependent services
    3. Adjust the readiness probe configuration if needed
    
    Commands to help diagnose:
    kubectl logs frontend-6d4cf56db6-abc12 -n default
    kubectl describe pod frontend-6d4cf56db6-abc12 -n default
    kubectl get endpoints -n default
    kubectl exec -it frontend-6d4cf56db6-abc12 -n default -- curl localhost:80/health

    Suggested Fix:
    You can try adjusting the readiness probe to give the application more time to start:
    
    kubectl patch deployment frontend -n default --patch '
    spec:
      template:
        spec:
          containers:
          - name: frontend
            readinessProbe:
              httpGet:
                path: /health
                port: 80
              initialDelaySeconds: 30
              periodSeconds: 10
              failureThreshold: 3
    '
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.