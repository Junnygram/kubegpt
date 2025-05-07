---

````markdown
# KubeGPT: AI-Powered Kubernetes Troubleshooting Assistant

![KubeGPT Logo](https://via.placeholder.com/800x200?text=KubeGPT+CLI)

KubeGPT is an AI-powered Kubernetes CLI tool that helps DevOps engineers and SREs diagnose, explain, and fix issues in their clusters with the help of Amazon Q Developer.

---

## 🧪 Demo Script

### Introduction

I'm excited to demonstrate KubeGPT, an AI-powered Kubernetes troubleshooting assistant that helps DevOps engineers and SREs diagnose and fix issues in their Kubernetes clusters.

### Setup

Before we begin, ensure you have:

- A working Kubernetes cluster configured with `kubectl`
- The `kubegpt` binary in your `PATH`
- [Amazon Q Developer CLI](https://docs.aws.amazon.com/amazonq/latest/developerguide/) installed

### Demo Flow

#### 1. Basic Usage

```bash
./kubegpt --help
./kubegpt version
```

#### 2. Diagnosing Cluster Issues

```bash
./kubegpt diagnose
./kubegpt diagnose --namespace monitoring
```

KubeGPT:

- Detects unhealthy pods
- Flags misconfigurations
- Analyzes problems using Amazon Q
- Suggests actionable fixes

#### 3. Explaining Kubernetes Errors

```bash
./kubegpt explain "CrashLoopBackOff: container exited with code 1"
kubectl logs my-pod | ./kubegpt explain
kubectl get deployment my-deployment -o yaml | ./kubegpt explain
```

#### 4. Generating Reports

```bash
./kubegpt report
./kubegpt report --all-namespaces
./kubegpt report --output markdown --file cluster-health.md
```

#### 5. Generating Fixes

```bash
./kubegpt diagnose --fix
```

---

## 💡 Features

- **AI Analysis**: Explains logs, events, YAML configs using Amazon Q Developer
- **Smart Diagnostics**: Detects common issues across pods, deployments, events
- **Fix Suggestions**: Offers YAML patches and kubectl commands
- **Report Generation**: Output in terminal, Markdown, or send to Slack
- **IaC Conversion**: Converts resources to Terraform, Pulumi, CDK, JSON, etc.
- **Interactive Game**: Test your Kubernetes knowledge

---

## 🔧 Installation

### Prerequisites

- Go 1.21+
- `kubectl` configured
- Amazon Q CLI installed:

  ```bash
  brew list amazon-q
  ```

### Build from Source

```bash
git clone https://github.com/yourusername/kubegpt.git
cd kubegpt
go build -o kubegpt
```

---

## 🚀 Usage & Command Reference

### Diagnose Command

```bash
./kubegpt diagnose
./kubegpt diagnose --namespace kube-system
./kubegpt diagnose --all-namespaces
./kubegpt diagnose --pods-only
./kubegpt diagnose --deployments-only
./kubegpt diagnose --fix
```

### Explain Command

```bash
./kubegpt explain "CrashLoopBackOff: container exited with code 1"
./kubegpt explain -f deployment.yaml
kubectl logs my-pod | ./kubegpt explain
```

### Transform Command

```bash
./kubegpt transform --target-lang terraform -f deployment.yaml -o deployment.tf
./kubegpt transform --target-lang pulumi-py -f deployment.yaml -o pulumi_app.py
./kubegpt transform --input-format yaml --output-format json -f deployment.yaml -o deployment.json
```

### Report Command

```bash
./kubegpt report
./kubegpt report --all-namespaces
./kubegpt report --output markdown --file cluster-health.md
./kubegpt report --output slack --slack-webhook https://hooks.slack.com/services/...
```

### Game Command

```bash
./kubegpt game
./kubegpt game --difficulty easy
./kubegpt game --difficulty medium
./kubegpt game --difficulty hard
```

### Version Command

```bash
./kubegpt version
```

---

## 🌍 Global Flags

```bash
./kubegpt --config /path/to/config.yaml [command]
./kubegpt --kubeconfig /path/to/kubeconfig [command]
./kubegpt --namespace default [command]
./kubegpt --verbose [command]
```

---

## 🧬 Environment Variables

```bash
export KUBEGPT_MOCK_AI=true        # Use mock AI output
export KUBECONFIG=/path/to/config  # Set custom kubeconfig
```

---

## 🧰 Full Workflow Example

```bash
./kubegpt diagnose
kubectl describe pod problematic-pod | ./kubegpt explain
./kubegpt diagnose --fix
./kubegpt transform --target-lang terraform -f fixed-deployment.yaml -o deployment.tf
./kubegpt report --output markdown --file cluster-health.md
```

---

## 🎮 Game Example

```bash
./kubegpt game
# Follow prompts to choose difficulty and answer questions.
```

---

## 💡 Pro Tips

- Use `--namespace` to narrow down diagnostics in large clusters.
- Pipe `kubectl` outputs directly into KubeGPT for inline explanations.
- Use `KUBEGPT_MOCK_AI=true` in dev environments to skip real AI analysis.
- Save diagnostic outputs to share with your team:

  ```bash
  ./kubegpt diagnose > diagnostic.txt
  ```

---

## 📬 Support & Contribution

If you'd like to contribute or have questions, open an issue on [GitHub](https://github.com/yourusername/kubegpt) or reach out on [Twitter](https://twitter.com/yourhandle).

---

## 📌 Conclusion

KubeGPT empowers SREs and platform engineers by:

- Reducing time to identify Kubernetes issues
- Leveraging AI for clear explanations
- Suggesting real fixes
- Converting insights into actionable IaC

Thank you for exploring KubeGPT!

```

```
