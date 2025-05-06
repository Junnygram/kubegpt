package utils

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// ConvertFormat converts between YAML and JSON formats
func ConvertFormat(content, inputFormat, outputFormat string) (string, error) {
	// Convert from input format to interface{}
	var data interface{}
	var err error

	switch inputFormat {
	case "yaml":
		err = yaml.Unmarshal([]byte(content), &data)
	case "json":
		err = json.Unmarshal([]byte(content), &data)
	default:
		return "", fmt.Errorf("unsupported input format: %s", inputFormat)
	}

	if err != nil {
		return "", fmt.Errorf("error parsing %s: %w", inputFormat, err)
	}

	// Convert from interface{} to output format
	var result []byte

	switch outputFormat {
	case "yaml":
		result, err = yaml.Marshal(data)
	case "json":
		result, err = json.MarshalIndent(data, "", "  ")
	default:
		return "", fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if err != nil {
		return "", fmt.Errorf("error generating %s: %w", outputFormat, err)
	}

	return string(result), nil
}