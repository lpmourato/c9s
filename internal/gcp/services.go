package gcp

import (
	"strings"

	run "google.golang.org/api/run/v2"
)

// ServiceSummary provides a simplified view of a Cloud Run service
type ServiceSummary struct {
	Name           string
	Region         string
	URL            string
	Status         string
	LastDeployTime string
	Traffic        []TrafficTarget
}

// TrafficTarget represents a traffic split target
type TrafficTarget struct {
	RevisionName string
	Percent      int32
}

// ToSummary converts a Cloud Run service to a simplified summary
func ToSummary(s *run.GoogleCloudRunV2Service) *ServiceSummary {
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

	summary := &ServiceSummary{
		Name:           name,
		Region:         region,
		URL:            s.Uri,
		Status:         getServiceStatus(s),
		LastDeployTime: s.UpdateTime,
	}

	if s.Template != nil && s.Template.Revision != "" {
		summary.Traffic = append(summary.Traffic, TrafficTarget{
			RevisionName: s.Template.Revision,
			Percent:      100,
		})
	}

	return summary
}

// getServiceStatus returns a human-readable status for the service
func getServiceStatus(s *run.GoogleCloudRunV2Service) string {
	if s.Conditions != nil {
		for _, condition := range s.Conditions {
			if condition.Type == "Ready" {
				if condition.State == "CONDITION_MET" {
					return "Ready"
				}
				return condition.State
			}
		}
	}
	return "Unknown"
}
