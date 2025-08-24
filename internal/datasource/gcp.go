package datasource

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	run "google.golang.org/api/run/v1"

	"github.com/lpmourato/c9s/internal/domain/cloudrun"
	"github.com/lpmourato/c9s/internal/gcp"
	"github.com/lpmourato/c9s/internal/model"
)

type gcpDataSource struct {
	projectID string
	client    *run.ProjectsLocationsServicesService
	provider  cloudrun.CloudRunProvider
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

	provider, err := gcp.NewServiceProvider(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create service provider: %v", err)
	}

	return &gcpDataSource{
		projectID: projectID,
		client:    servicesService,
		provider:  provider,
	}, nil
}

func (ds *gcpDataSource) GetServices() ([]model.Service, error) {
	if ds.projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	// List services in all regions
	regions := []string{
		"asia-east1",
		"europe-west1",
		"europe-west2",
		"europe-west3",
		"europe-west4",
		"us-central1",
		"us-east1",
		"us-east4",
		"us-west1",
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

func (ds *gcpDataSource) GetProvider() cloudrun.CloudRunProvider {
	return ds.provider
}

func (ds *gcpDataSource) GetServicesByRegion(region string) ([]model.Service, error) {
	if ds.projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if region == "" {
		return nil, fmt.Errorf("region is required")
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", ds.projectID, region)
	resp, err := ds.client.List(parent).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list services in %s: %v", region, err)
	}

	services := make([]model.Service, 0, len(resp.Items))
	for _, svc := range resp.Items {
		services = append(services, cloudrun.NewCloudRunServiceFromGCP(svc, region))
	}

	return services, nil
}
