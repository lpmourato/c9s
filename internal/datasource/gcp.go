package datasource

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/option"
	run "google.golang.org/api/run/v1"

	"github.com/lpmourato/c9s/internal/domain/cloudrun"
	"github.com/lpmourato/c9s/internal/model"
)

type gcpDataSource struct {
	projectID string
	client    *run.ProjectsLocationsServicesService
}

func newGCPDataSource(projectID string) (DataSource, error) {
	ctx := context.Background()

	// Create Cloud Run client
	runService, err := run.NewService(ctx, option.WithScopes(run.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run client: %v", err)
	}

	// Get the services service which we'll use for API calls
	servicesService := run.NewProjectsLocationsServicesService(runService)

	return &gcpDataSource{
		projectID: projectID,
		client:    servicesService,
	}, nil
}

func (ds *gcpDataSource) GetServices() ([]model.Service, error) {
	if ds.projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	// List services in all regions
	regions := []string{
		"us-central1",
		"us-east1",
		"us-east4",
		"us-west1",
		"europe-west1",
		"asia-east1",
	}

	var allServices []model.Service
	for _, region := range regions {
		services, err := ds.GetServicesByRegion(region)
		if err != nil {
			// Log error but continue with other regions
			fmt.Printf("Warning: failed to get services in %s: %v\n", region, err)
			continue
		}
		allServices = append(allServices, services...)
	}

	return allServices, nil
}

func (ds *gcpDataSource) GetServicesByRegion(region string) ([]model.Service, error) {
	if ds.projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if region == "" {
		return nil, fmt.Errorf("region is required")
	}

	// Format parent path for Cloud Run API
	parent := fmt.Sprintf("projects/%s/locations/%s", ds.projectID, region)

	// List services in the region
	resp, err := ds.client.List(parent).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list services in %s: %v", region, err)
	}

	var services []model.Service
	for _, svc := range resp.Items {
		// Get traffic allocation
		var traffic string
		if len(svc.Status.Traffic) == 0 {
			traffic = "0%"
		} else if len(svc.Status.Traffic) == 1 {
			traffic = "100%"
		} else {
			traffic = ""
			for i, tr := range svc.Status.Traffic {
				if i > 0 {
					traffic += ", "
				}
				if tr.RevisionName != "" {
					rev := tr.RevisionName[strings.LastIndex(tr.RevisionName, "-")+1:]
					traffic += fmt.Sprintf("%s (%d%%)", rev, tr.Percent)
				}
			}
		}

		// Parse last deploy time
		var lastDeploy time.Time
		if svc.Status.ObservedGeneration > 0 && len(svc.Status.Conditions) > 0 {
			for _, cond := range svc.Status.Conditions {
				if cond.LastTransitionTime != "" {
					if t, err := time.Parse(time.RFC3339, cond.LastTransitionTime); err == nil {
						if lastDeploy.IsZero() || t.After(lastDeploy) {
							lastDeploy = t
						}
					}
				}
			}
		}

		// Create CloudRunService
		service := &cloudrun.CloudRunService{
			Name:       svc.Metadata.Name,
			Region:     region,
			URL:        svc.Status.Url,
			Status:     svc.Status.Conditions[0].Status,
			LastDeploy: lastDeploy,
			Traffic:    traffic,
		}

		services = append(services, service)
	}

	return services, nil
}
