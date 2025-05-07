#!/bin/bash
# KubeGPT Demo Script
# This script contains commands to demonstrate KubeGPT's capabilities

# Set environment variables for demo
export KUBEGPT_MOCK_AI=true  # Use mock AI responses for demo

# Show help
echo -e "\n\033[1;36m=== Showing KubeGPT Help ===\033[0m"
./kubegpt --help

# Show version
echo -e "\n\033[1;36m=== Showing KubeGPT Version ===\033[0m"
./kubegpt version

# Run diagnostic
echo -e "\n\033[1;36m=== Running Diagnostic on Current Namespace ===\033[0m"
./kubegpt diagnose

# Run diagnostic on specific namespace
echo -e "\n\033[1;36m=== Running Diagnostic on 'default' Namespace ===\033[0m"
./kubegpt diagnose --namespace default

# Run diagnostic with pods only
echo -e "\n\033[1;36m=== Running Diagnostic on Pods Only ===\033[0m"
./kubegpt diagnose --pods-only

# Explain a Kubernetes error
echo -e "\n\033[1;36m=== Explaining CrashLoopBackOff Error ===\033[0m"
./kubegpt explain "CrashLoopBackOff: container exited with code 1"

# Explain an ImagePullBackOff error
echo -e "\n\033[1;36m=== Explaining ImagePullBackOff Error ===\033[0m"
./kubegpt explain "ImagePullBackOff: Back-off pulling image myregistry.com/myapp:latest"

# Generate a report
echo -e "\n\033[1;36m=== Generating Cluster Health Report ===\033[0m"
./kubegpt report

# Generate fixes
echo -e "\n\033[1;36m=== Generating Fixes for Issues ===\033[0m"
./kubegpt diagnose --fix

# Generate markdown report
echo -e "\n\033[1;36m=== Generating Markdown Report ===\033[0m"
./kubegpt report --output markdown --file cluster-health.md
echo "Report saved to cluster-health.md"

echo -e "\n\033[1;32m=== Demo Complete ===\033[0m"