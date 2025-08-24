package cloudrun

import (
	"fmt"
	"strings"
	"time"

	run "google.golang.org/api/run/v1"

	"github.com/lpmourato/c9s/internal/model"
)

// getTrafficAllocation returns a formatted string representing traffic distribution
func getTrafficAllocation(traffic []*run.TrafficTarget) string {
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

// getLastDeployTime returns the most recent deployment time from service conditions
func getLastDeployTime(svc *run.Service) time.Time {
	if svc.Status.ObservedGeneration == 0 || len(svc.Status.Conditions) == 0 {
		return time.Time{}
	}

	var lastDeploy time.Time
	for _, cond := range svc.Status.Conditions {
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

// NewCloudRunServiceFromGCP creates a new CloudRunService from a GCP service object
func NewCloudRunServiceFromGCP(svc *run.Service, region string) *model.CloudRunService {
	return &model.CloudRunService{
		Name:       svc.Metadata.Name,
		Region:     region,
		URL:        svc.Status.Url,
		Status:     svc.Status.Conditions[0].Status,
		LastDeploy: getLastDeployTime(svc),
		Traffic:    getTrafficAllocation(svc.Status.Traffic),
	}
}
