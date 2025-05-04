package k8s

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventIssue represents a failed event
type EventIssue struct {
	Name           string
	Namespace      string
	Type           string
	Reason         string
	Message        string
	Count          int32
	FirstTimestamp time.Time
	LastTimestamp  time.Time
	InvolvedObject v1.ObjectReference
	Analysis       string
	Fix            string
}

// GetFailedEvents returns a list of failed events in the current namespace
func (c *Client) GetFailedEvents(ctx context.Context) ([]EventIssue, error) {
	events, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var failedEvents []EventIssue

	for _, event := range events.Items {
		// Skip normal events
		if event.Type != v1.EventTypeWarning {
			continue
		}

		// Skip old events (older than 1 hour)
		if time.Since(event.LastTimestamp.Time) > time.Hour {
			continue
		}

		eventIssue := EventIssue{
			Name:           event.Name,
			Namespace:      event.Namespace,
			Type:           event.Type,
			Reason:         event.Reason,
			Message:        event.Message,
			Count:          event.Count,
			FirstTimestamp: event.FirstTimestamp.Time,
			LastTimestamp:  event.LastTimestamp.Time,
			InvolvedObject: event.InvolvedObject,
		}

		failedEvents = append(failedEvents, eventIssue)
	}

	return failedEvents, nil
}

// GetResourceEvents returns events for a specific resource
func (c *Client) GetResourceEvents(ctx context.Context, kind, name string) ([]v1.Event, error) {
	fieldSelector := fmt.Sprintf("involvedObject.kind=%s,involvedObject.name=%s,involvedObject.namespace=%s", 
		kind, name, c.namespace)
	
	events, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events for %s %s: %w", kind, name, err)
	}

	return events.Items, nil
}

// GetRecentEvents returns recent events (within the specified duration)
func (c *Client) GetRecentEvents(ctx context.Context, duration time.Duration) ([]v1.Event, error) {
	events, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var recentEvents []v1.Event
	cutoffTime := time.Now().Add(-duration)

	for _, event := range events.Items {
		if event.LastTimestamp.Time.After(cutoffTime) {
			recentEvents = append(recentEvents, event)
		}
	}

	return recentEvents, nil
}

// GetEventsByType returns events of a specific type
func (c *Client) GetEventsByType(ctx context.Context, eventType string) ([]v1.Event, error) {
	events, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("type=%s", eventType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events of type %s: %w", eventType, err)
	}

	return events.Items, nil
}

// GetEventsByReason returns events with a specific reason
func (c *Client) GetEventsByReason(ctx context.Context, reason string) ([]v1.Event, error) {
	events, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("reason=%s", reason),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events with reason %s: %w", reason, err)
	}

	return events.Items, nil
}