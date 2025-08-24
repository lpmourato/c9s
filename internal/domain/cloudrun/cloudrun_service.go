package cloudrun

import (
	"fmt"
	"strings"
	"time"

	run "google.golang.org/api/run/v1"

	"github.com/lpmourato/c9s/internal/model"
)

// CloudRunGCPService represents a GCP implementation of Cloud Run service
type CloudRunGCPService struct {
	*model.CloudRunService
	rawService *run.Service
}

// NewCloudRunGCPService creates a new Cloud Run service instance from GCP service
func NewCloudRunGCPService(svc *run.Service, region string) *CloudRunGCPService {
	service := &CloudRunGCPService{
		CloudRunService: &model.CloudRunService{
			Name:   svc.Metadata.Name,
			Region: region,
			URL:    svc.Status.Url,
		},
		rawService: svc,
	}
	service.RefreshStatus() // This will set Status, LastDeploy, and Traffic
	return service
}

// GetLastDeployTime returns the most recent deployment time from service conditions
func (s *CloudRunGCPService) GetLastDeployTime() time.Time {
	if s.rawService.Status.ObservedGeneration == 0 || len(s.rawService.Status.Conditions) == 0 {
		return time.Time{}
	}

	var lastDeploy time.Time
	for _, cond := range s.rawService.Status.Conditions {
		if cond.LastTransitionTime == "" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, cond.LastTransitionTime); err == nil {
			if lastDeploy.IsZero() || t.After(lastDeploy) {
				lastDeploy = t
			}
		}
	}
	return lastDeploy
}

// GetTrafficAllocation returns a formatted string representing traffic distribution
func (s *CloudRunGCPService) GetTrafficAllocation() string {
	traffic := s.rawService.Status.Traffic
	if len(traffic) == 0 {
		return "0%"
	}
	if len(traffic) == 1 {
		return "100%"
	}

	var allocations []string
	for _, tr := range traffic {
		if tr.RevisionName == "" {
			continue
		}
		rev := tr.RevisionName[strings.LastIndex(tr.RevisionName, "-")+1:]
		allocations = append(allocations, fmt.Sprintf("%s (%d%%)", rev, tr.Percent))
	}
	return strings.Join(allocations, ", ")
}

// RefreshStatus updates the service status from the raw GCP service data
func (s *CloudRunGCPService) RefreshStatus() {
	s.Status = s.rawService.Status.Conditions[0].Status
	s.LastDeploy = s.GetLastDeployTime()
	s.Traffic = s.GetTrafficAllocation()
}
