package ai

// This file contains prompt templates for Amazon Q

// PodIssuePromptTemplate is the template for pod issue analysis
const PodIssuePromptTemplate = `
As a Kubernetes expert, please analyze this pod issue:

Pod: {{.Name}}
Namespace: {{.Namespace}}
Status: {{.Status}}
Age: {{.Age}}
Node: {{.Node}}
{{if .Message}}Message: {{.Message}}{{end}}
{{if .Reason}}Reason: {{.Reason}}{{end}}

Container issues:
{{.ContainerIssues}}

Events:
{{.Events}}

{{if .Logs}}
Container logs (most recent):
{{.Logs}}
{{end}}

Please provide:
1. A diagnosis of the issue
2. Likely root causes
3. Recommended solutions
4. Specific kubectl commands to help diagnose or fix the issue
`

// DeploymentIssuePromptTemplate is the template for deployment issue analysis
const DeploymentIssuePromptTemplate = `
As a Kubernetes expert, please analyze this deployment issue:

Deployment: {{.Name}}
Namespace: {{.Namespace}}
Replicas: {{.ReadyReplicas}}/{{.Replicas}} ready
Updated replicas: {{.UpdatedReplicas}}/{{.Replicas}}
Available replicas: {{.AvailableReplicas}}/{{.Replicas}}
Strategy: {{.Strategy}}
Age: {{.Age}}
{{if .Message}}Message: {{.Message}}{{end}}
{{if .Reason}}Reason: {{.Reason}}{{end}}

Conditions:
{{.Conditions}}

Events:
{{.Events}}

Please provide:
1. A diagnosis of the issue
2. Likely root causes
3. Recommended solutions
4. Specific kubectl commands to help diagnose or fix the issue
`

// ServiceIssuePromptTemplate is the template for service issue analysis
const ServiceIssuePromptTemplate = `
As a Kubernetes expert, please analyze this service issue:

Service: {{.Name}}
Namespace: {{.Namespace}}
Type: {{.Type}}
ClusterIP: {{.ClusterIP}}
{{if .ExternalIP}}ExternalIP: {{.ExternalIP}}{{end}}
Selector: {{.Selector}}
Endpoints: {{.EndpointCount}}
Age: {{.Age}}
{{if .Message}}Message: {{.Message}}{{end}}
{{if .Reason}}Reason: {{.Reason}}{{end}}

Events:
{{.Events}}

Please provide:
1. A diagnosis of the issue
2. Likely root causes
3. Recommended solutions
4. Specific kubectl commands to help diagnose or fix the issue
`

// ErrorPromptTemplate is the template for error analysis
const ErrorPromptTemplate = `
As a Kubernetes expert, please analyze this error message and explain:
1. What the error means
2. Likely causes
3. How to fix it
4. Specific kubectl commands that might help diagnose or fix the issue

Error message:
{{.ErrorMessage}}
`

// LogsPromptTemplate is the template for logs analysis
const LogsPromptTemplate = `
As a Kubernetes expert, please analyze these logs and explain:
1. What issues or errors are present
2. Root causes of any problems
3. How to fix the issues
4. Specific kubectl commands that might help diagnose or fix the issues

Logs:
{{.Logs}}
`

// YAMLPromptTemplate is the template for YAML analysis
const YAMLPromptTemplate = `
As a Kubernetes expert, please analyze this YAML configuration and explain:
1. What this configuration does
2. Any issues, misconfigurations, or security concerns
3. Best practices that should be applied
4. Specific improvements that could be made

YAML:
{{.YAML}}
`

// FixPromptTemplate is the template for generating fixes
const FixPromptTemplate = `
As a Kubernetes expert, please generate a fix for this issue:

{{.IssueDescription}}

Please provide:
1. A brief explanation of the fix
2. YAML patch or kubectl commands to apply the fix
3. Any additional steps needed
`

// ClusterReportPromptTemplate is the template for generating cluster reports
const ClusterReportPromptTemplate = `
As a Kubernetes expert, please analyze this cluster health report and provide insights:

Cluster Overview:
- Total namespaces: {{.NamespaceCount}}
- Total pods: {{.TotalPods}} ({{.HealthyPods}} healthy, {{.UnhealthyPods}} unhealthy)
- Total deployments: {{.TotalDeployments}} ({{.HealthyDeployments}} healthy, {{.MisconfiguredDeployments}} misconfigured)
- Failed events in the last hour: {{.FailedEventCount}}

{{if .UnhealthyResources}}
Unhealthy Resources:
{{.UnhealthyResources}}
{{end}}

Please provide:
1. An assessment of the cluster health
2. Key issues that need attention
3. Recommended actions to improve cluster health
4. Prioritization of issues (what to fix first)
`