package utils

import (
	"fmt"
	"strings"
)

// FormatError formats an error message for display
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()
	
	// Remove common kubectl error prefixes
	msg = strings.TrimPrefix(msg, "error: ")
	msg = strings.TrimPrefix(msg, "Error: ")
	
	// Clean up kubectl error output
	if strings.Contains(msg, "kubectl error:") {
		parts := strings.SplitN(msg, "Output:", 2)
		if len(parts) > 1 {
			msg = strings.TrimSpace(parts[1])
		}
	}
	
	return msg
}

// WrapError wraps an error with a context message
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}