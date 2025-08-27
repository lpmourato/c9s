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

type cloudRunDataSource struct {
	projectID string
	client    *run.ProjectsLocationsServicesService
	provider  *cloudrun.Provider
}

func newCloudRunDataSource(projectID string) (DataSource, error) {
	ctx := context.Background()

	// Create Cloud Run client
	runService, err := run.NewService(ctx, option.WithScopes(run.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run client: %v", err)
	}

	// Get the services service which we'll use for API calls
	servicesService := run.NewProjectsLocationsServicesService(runService)

	baseProvider, err := gcp.NewServiceProvider(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create service provider: %v", err)
	}

	provider := cloudrun.NewProvider(baseProvider)

	return &cloudRunDataSource{
		projectID: projectID,
		client:    servicesService,
		provider:  provider,
	}, nil
}

func (ds *cloudRunDataSource) GetServices() ([]model.Service, error) {
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
			// Skip regions that fail but continue with others
			continue
		}
		allServices = append(allServices, services...)
	}

	return allServices, nil
}

func (ds *cloudRunDataSource) GetProvider() model.CloudRunProvider {
	return ds.provider
}

func (ds *cloudRunDataSource) GetServicesByRegion(region string) ([]model.Service, error) {
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
		services = append(services, cloudrun.NewCloudRunGCPService(svc, region))
	}

	return services, nil
}

func (ds *cloudRunDataSource) GetServiceDetails(name, region string) (*model.ServiceDetails, error) {
	if ds.projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if region == "" {
		return nil, fmt.Errorf("region is required")
	}
	if name == "" {
		return nil, fmt.Errorf("service name is required")
	}

	ctx := context.Background()
	return ds.provider.GetServiceDetails(ctx, name, region)
}
