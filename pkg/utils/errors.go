package utils

import (
	"fmt"
	"strings"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypePermission represents a permission error
	ErrorTypePermission ErrorType = "permission"
	// ErrorTypeNotFound represents a not found error
	ErrorTypeNotFound ErrorType = "not-found"
	// ErrorTypeConnection represents a connection error
	ErrorTypeConnection ErrorType = "connection"
	// ErrorTypeConfiguration represents a configuration error
	ErrorTypeConfiguration ErrorType = "configuration"
	// ErrorTypeResource represents a resource error
	ErrorTypeResource ErrorType = "resource"
	// ErrorTypeUnknown represents an unknown error
	ErrorTypeUnknown ErrorType = "unknown"
)

// DetectErrorType detects the type of error from an error message
func DetectErrorType(errorMsg string) ErrorType {
	errorMsg = strings.ToLower(errorMsg)

	// Check for permission errors
	if strings.Contains(errorMsg, "forbidden") ||
		strings.Contains(errorMsg, "unauthorized") ||
		strings.Contains(errorMsg, "permission denied") ||
		strings.Contains(errorMsg, "cannot get") ||
		strings.Contains(errorMsg, "cannot list") ||
		strings.Contains(errorMsg, "cannot create") ||
		strings.Contains(errorMsg, "cannot update") ||
		strings.Contains(errorMsg, "cannot delete") {
		return ErrorTypePermission
	}

	// Check for not found errors
	if strings.Contains(errorMsg, "not found") ||
		strings.Contains(errorMsg, "no such") ||
		strings.Contains(errorMsg, "doesn't exist") ||
		strings.Contains(errorMsg, "does not exist") {
		return ErrorTypeNotFound
	}

	// Check for connection errors
	if strings.Contains(errorMsg, "connection") ||
		strings.Contains(errorMsg, "dial") ||
		strings.Contains(errorMsg, "timeout") ||
		strings.Contains(errorMsg, "refused") ||
		strings.Contains(errorMsg, "unreachable") ||
		strings.Contains(errorMsg, "network") {
		return ErrorTypeConnection
	}

	// Check for configuration errors
	if strings.Contains(errorMsg, "configuration") ||
		strings.Contains(errorMsg, "config") ||
		strings.Contains(errorMsg, "invalid") ||
		strings.Contains(errorMsg, "missing") ||
		strings.Contains(errorMsg, "required") {
		return ErrorTypeConfiguration
	}

	// Check for resource errors
	if strings.Contains(errorMsg, "resource") ||
		strings.Contains(errorMsg, "quota") ||
		strings.Contains(errorMsg, "limit") ||
		strings.Contains(errorMsg, "insufficient") ||
		strings.Contains(errorMsg, "exceeded") {
		return ErrorTypeResource
	}

	return ErrorTypeUnknown
}

// FormatError formats an error message for display
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	errorMsg := err.Error()
	errorType := DetectErrorType(errorMsg)

	switch errorType {
	case ErrorTypePermission:
		return fmt.Sprintf("Permission Error: %s", errorMsg)
	case ErrorTypeNotFound:
		return fmt.Sprintf("Not Found Error: %s", errorMsg)
	case ErrorTypeConnection:
		return fmt.Sprintf("Connection Error: %s", errorMsg)
	case ErrorTypeConfiguration:
		return fmt.Sprintf("Configuration Error: %s", errorMsg)
	case ErrorTypeResource:
		return fmt.Sprintf("Resource Error: %s", errorMsg)
	default:
		return fmt.Sprintf("Error: %s", errorMsg)
	}
}

// SuggestFix suggests a fix for a common error
func SuggestFix(errorMsg string) string {
	errorType := DetectErrorType(errorMsg)
	errorMsg = strings.ToLower(errorMsg)

	switch errorType {
	case ErrorTypePermission:
		if strings.Contains(errorMsg, "serviceaccount") {
			return "This appears to be a permissions issue with your service account. Try creating a Role or ClusterRole with the necessary permissions and bind it to your service account."
		}
		return "This appears to be a permissions issue. Check that you have the necessary RBAC permissions to perform this action."

	case ErrorTypeNotFound:
		if strings.Contains(errorMsg, "pod") {
			return "The pod was not found. Check if it exists in the correct namespace or if it has been deleted."
		}
		if strings.Contains(errorMsg, "service") {
			return "The service was not found. Check if it exists in the correct namespace or if it has been deleted."
		}
		return "The resource was not found. Check if it exists in the correct namespace or if it has been deleted."

	case ErrorTypeConnection:
		if strings.Contains(errorMsg, "refused") {
			return "Connection was refused. Check if the Kubernetes API server is running and accessible."
		}
		return "This appears to be a connection issue. Check your network connectivity and make sure the Kubernetes API server is accessible."

	case ErrorTypeConfiguration:
		if strings.Contains(errorMsg, "kubeconfig") {
			return "There seems to be an issue with your kubeconfig file. Make sure it exists and is properly configured."
		}
		return "This appears to be a configuration issue. Check your configuration files and environment variables."

	case ErrorTypeResource:
		if strings.Contains(errorMsg, "memory") {
			return "There are insufficient memory resources. Consider increasing the memory limit or reducing the memory request."
		}
		if strings.Contains(errorMsg, "cpu") {
			return "There are insufficient CPU resources. Consider increasing the CPU limit or reducing the CPU request."
		}
		return "This appears to be a resource issue. Check your resource quotas and limits."

	default:
		return "No specific fix suggestion available for this error."
	}
}