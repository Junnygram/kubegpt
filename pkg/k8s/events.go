package k8s

import (
	"encoding/json"
	"fmt"
	"time"
)

// GetFailedEvents gets all failed events in the specified namespace
func (c *Client) GetFailedEvents(namespace string) ([]interface{}, error) {
	// Get all events in the namespace
	output, err := c.ExecuteKubectl("get", "events", "-n", namespace, "-o", "json")
	if err != nil {
		return nil, err
	}

	// Parse the JSON output
	var eventList struct {
		Items []struct {
			Type      string    `json:"type"`
			Reason    string    `json:"reason"`
			Message   string    `json:"message"`
			Count     int       `json:"count"`
			FirstTimestamp string `json:"firstTimestamp"`
			LastTimestamp  string `json:"lastTimestamp"`
			InvolvedObject struct {
				Kind      string `json:"kind"`
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"involvedObject"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &eventList); err != nil {
		return nil, fmt.Errorf("error parsing event list: %w", err)
	}

	// Filter failed events
	var failedEvents []interface{}
	for _, event := range eventList.Items {
		// Only include Warning events
		if event.Type != "Warning" {
			continue
		}

		// Skip old events (more than 1 hour old)
		if lastTime, err := time.Parse(time.RFC3339, event.LastTimestamp); err == nil {
			if time.Since(lastTime) > time.Hour {
				continue
			}
		}

		// Add the event to the list
		failedEvents = append(failedEvents, event)
	}

	return failedEvents, nil
}