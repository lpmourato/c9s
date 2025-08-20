package gcp

import (
	"sort"
	"strings"

	run "google.golang.org/api/run/v2"
)

// ServiceSummary provides a simplified view of a Cloud Run service
type ServiceSummary struct {
	Name               string
	Region             string
	URL                string
	Status             string
	RawReadyState      string // Raw Ready condition state
	RawServingState    string // Raw Serving condition state
	ReadyMessage       string // Message from Ready condition
	ServingMessage     string // Message from Serving condition
	LastDeployTime     string
	Traffic            []TrafficTarget
	ActualInstance     int32  // Current number of instances running
	MinInstances       int32  // Minimum number of instances configured
	MaxInstances       int32  // Maximum number of instances allowed
	IsServing          bool   // Whether the service is actually serving traffic
	IsScaledToZero     bool   // Whether the service has scaled to zero
	IsFullyProvisioned bool   // Whether the service has all required resources
	Revision           string // Current active revision
}

// TrafficTarget represents a traffic split target
type TrafficTarget struct {
	RevisionName string
	Percent      int32
}

// ToSummary converts a Cloud Run service to a simplified summary
func ToSummary(s *run.GoogleCloudRunV2Service) *ServiceSummary {
	if s == nil {
		return &ServiceSummary{Status: "Unknown"}
	}

	// Extract region from name (format: projects/*/locations/*/services/*)
	parts := strings.Split(s.Name, "/")
	region := ""
	if len(parts) >= 4 {
		region = parts[3]
	}
	name := ""
	if len(parts) >= 6 {
		name = parts[5]
	}

	status := getServiceStatus(s)

	// Get raw states and check conditions
	var rawReadyState, rawServingState, readyMessage, servingMessage string
	var isServing, isFullyProvisioned bool
	for _, condition := range s.Conditions {
		switch condition.Type {
		case "Ready":
			rawReadyState = condition.State
			readyMessage = condition.Message
			isFullyProvisioned = condition.State == "CONDITION_MET"
		case "Serving":
			rawServingState = condition.State
			servingMessage = condition.Message
			isServing = condition.State == "CONDITION_MET"
		}
	}

	// Get scaling settings
	var minInstances, maxInstances int32
	if s.Template != nil && s.Template.Scaling != nil {
		minInstances = int32(s.Template.Scaling.MinInstanceCount)
		maxInstances = int32(s.Template.Scaling.MaxInstanceCount)
	}

	// Get current instances (from latest ready revision)
	var actualInstances int32
	if s.LatestReadyRevision != "" && s.TrafficStatuses != nil {
		for _, ts := range s.TrafficStatuses {
			if ts.Revision == s.LatestReadyRevision {
				// In Cloud Run, if there's traffic and the service is ready/serving,
				// there should be at least one instance
				if ts.Percent > 0 && isServing {
					actualInstances = 1
				}
				break
			}
		}
	}

	// Determine if service is scaled to zero
	isScaledToZero := isServing && actualInstances == 0

	summary := &ServiceSummary{
		Name:               name,
		Region:             region,
		URL:                s.Uri,
		Status:             status,
		RawReadyState:      rawReadyState,
		RawServingState:    rawServingState,
		ReadyMessage:       readyMessage,
		ServingMessage:     servingMessage,
		LastDeployTime:     s.UpdateTime,
		ActualInstance:     actualInstances,
		MinInstances:       minInstances,
		MaxInstances:       maxInstances,
		IsServing:          isServing,
		IsScaledToZero:     isScaledToZero,
		IsFullyProvisioned: isFullyProvisioned,
		Revision:           s.LatestReadyRevision,
	}

	// Handle traffic configuration
	if s.Traffic != nil {
		for _, t := range s.Traffic {
			revisionName := t.Revision
			// Extract just the revision portion for cleaner display
			if parts := strings.Split(revisionName, "/"); len(parts) > 0 {
				revisionName = parts[len(parts)-1]
				if strings.HasPrefix(revisionName, name+"-") {
					revisionName = strings.TrimPrefix(revisionName, name+"-")
				}
			}
			summary.Traffic = append(summary.Traffic, TrafficTarget{
				RevisionName: revisionName,
				Percent:      int32(t.Percent),
			})
		}

		// Sort traffic by percentage descending
		sort.Slice(summary.Traffic, func(i, j int) bool {
			return summary.Traffic[i].Percent > summary.Traffic[j].Percent
		})
	}

	return summary
}

// getServiceStatus returns a human-readable status for the service
func getServiceStatus(s *run.GoogleCloudRunV2Service) string {
	if s == nil {
		return "Unknown"
	}

	// Collect current state
	var readyState, servingState string
	var failureMessage string

	for _, condition := range s.Conditions {
		switch condition.Type {
		case "Ready":
			readyState = condition.State
			if condition.State == "CONDITION_FAILED" {
				failureMessage = condition.Message
			}
		case "Serving":
			servingState = condition.State
		}
	}

	// Handle active transitions first
	if s.Reconciling || s.LatestCreatedRevision != s.LatestReadyRevision {
		if s.LatestReadyRevision == "" {
			return "Creating"
		}
		return "Updating"
	}

	// Determine final status
	switch {
	case readyState == "CONDITION_MET" && servingState == "CONDITION_MET":
		return "Ready"
	case readyState == "CONDITION_FAILED":
		if failureMessage != "" {
			msg := strings.Split(failureMessage, ".")[0]
			msg = strings.TrimSpace(msg)
			if len(msg) > 50 {
				msg = msg[:47] + "..."
			}
			return "Failed: " + msg
		}
		return "Failed"
	case servingState == "CONDITION_FAILED":
		return "Stopped"
	case readyState == "CONDITION_PENDING" || servingState == "CONDITION_PENDING":
		return "Pending"
	case readyState == "" && servingState == "":
		return "Unknown"
	default:
		// Return the most specific state we have
		if readyState != "" {
			return readyState
		}
		if servingState != "" {
			return servingState
		}
		return "Unknown"
	}
}
