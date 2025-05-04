package utils

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

// ParseYAML parses a YAML string into a map
func ParseYAML(yamlStr string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlStr), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return result, nil
}

// MarshalYAML marshals a map into a YAML string
func MarshalYAML(data interface{}) (string, error) {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(bytes), nil
}

// ReadYAMLFile reads a YAML file and parses it into a map
func ReadYAMLFile(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	return ParseYAML(string(data))
}

// WriteYAMLFile writes a map to a YAML file
func WriteYAMLFile(filename string, data interface{}) error {
	yamlStr, err := MarshalYAML(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(yamlStr), 0644)
}

// GenerateYAMLPatch generates a YAML patch from original and modified YAML
func GenerateYAMLPatch(original, modified string) (string, error) {
	// Parse original and modified YAML
	originalMap, err := ParseYAML(original)
	if err != nil {
		return "", err
	}
	modifiedMap, err := ParseYAML(modified)
	if err != nil {
		return "", err
	}

	// Generate patch
	patch := generatePatch(originalMap, modifiedMap)
	if len(patch) == 0 {
		return "", fmt.Errorf("no changes detected")
	}

	// Marshal patch to YAML
	patchYAML, err := MarshalYAML(patch)
	if err != nil {
		return "", err
	}

	return patchYAML, nil
}

// generatePatch generates a patch map from original and modified maps
func generatePatch(original, modified map[string]interface{}) map[string]interface{} {
	patch := make(map[string]interface{})

	// Add apiVersion and kind from original
	if apiVersion, ok := original["apiVersion"]; ok {
		patch["apiVersion"] = apiVersion
	}
	if kind, ok := original["kind"]; ok {
		patch["kind"] = kind
	}

	// Add metadata.name from original
	if metadata, ok := original["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"]; ok {
			if patch["metadata"] == nil {
				patch["metadata"] = make(map[string]interface{})
			}
			patch["metadata"].(map[string]interface{})["name"] = name
		}
	}

	// Add spec changes
	if originalSpec, ok := original["spec"].(map[string]interface{}); ok {
		if modifiedSpec, ok := modified["spec"].(map[string]interface{}); ok {
			specPatch := generateSpecPatch(originalSpec, modifiedSpec)
			if len(specPatch) > 0 {
				patch["spec"] = specPatch
			}
		}
	}

	return patch
}

// generateSpecPatch generates a patch for the spec section
func generateSpecPatch(original, modified map[string]interface{}) map[string]interface{} {
	patch := make(map[string]interface{})

	// Check for changes in modified
	for key, modValue := range modified {
		origValue, exists := original[key]
		if !exists {
			// New field
			patch[key] = modValue
		} else if !deepEqual(origValue, modValue) {
			// Changed field
			patch[key] = modValue
		}
	}

	return patch
}

// deepEqual checks if two values are deeply equal
func deepEqual(a, b interface{}) bool {
	// Simple case: both are nil or equal
	if a == nil && b == nil {
		return true
	}
	if a == b {
		return true
	}

	// Check maps
	aMap, aIsMap := a.(map[string]interface{})
	bMap, bIsMap := b.(map[string]interface{})
	if aIsMap && bIsMap {
		if len(aMap) != len(bMap) {
			return false
		}
		for k, v := range aMap {
			if !deepEqual(v, bMap[k]) {
				return false
			}
		}
		return true
	}

	// Check slices
	aSlice, aIsSlice := a.([]interface{})
	bSlice, bIsSlice := b.([]interface{})
	if aIsSlice && bIsSlice {
		if len(aSlice) != len(bSlice) {
			return false
		}
		for i := range aSlice {
			if !deepEqual(aSlice[i], bSlice[i]) {
				return false
			}
		}
		return true
	}

	// Different types or values
	return false
}

// FormatYAML formats a YAML string with proper indentation
func FormatYAML(yamlStr string) string {
	// Parse and re-marshal to ensure proper formatting
	var data interface{}
	err := yaml.Unmarshal([]byte(yamlStr), &data)
	if err != nil {
		return yamlStr // Return original if parsing fails
	}

	formatted, err := yaml.Marshal(data)
	if err != nil {
		return yamlStr // Return original if marshaling fails
	}

	return string(formatted)
}

// ExtractKind extracts the kind from a YAML string
func ExtractKind(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "kind:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// ExtractName extracts the metadata.name from a YAML string
func ExtractName(yamlStr string) string {
	inMetadata := false
	lines := strings.Split(yamlStr, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "metadata:" {
			inMetadata = true
			continue
		}
		if inMetadata && strings.HasPrefix(trimmed, "name:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}