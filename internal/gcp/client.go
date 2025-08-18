package gcp

import (
	"context"
	"fmt"

	run "google.golang.org/api/run/v2"
	"google.golang.org/api/option"
)

// Client represents a GCP Cloud Run client
type Client struct {
	runService *run.ProjectsLocationsServicesService
	project    string
	region     string
}

// NewClient creates a new GCP Cloud Run client
func NewClient(ctx context.Context, project, region string) (*Client, error) {
	runClient, err := run.NewService(ctx, option.WithScopes(run.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run client: %v", err)
	}

	return &Client{
		runService: runClient.Projects.Locations.Services,
		project:    project,
		region:     region,
	}, nil
}

// Project returns the configured project ID
func (c *Client) Project() string {
	return c.project
}

// Region returns the configured region
func (c *Client) Region() string {
	return c.region
}

// ListServices returns all Cloud Run services in the configured project and region
func (c *Client) ListServices(ctx context.Context) ([]*run.GoogleCloudRunV2Service, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", c.project, c.region)
	
	resp, err := c.runService.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %v", err)
	}

	return resp.Services, nil
}

// GetService retrieves a specific Cloud Run service
func (c *Client) GetService(ctx context.Context, name string) (*run.GoogleCloudRunV2Service, error) {
	fullName := fmt.Sprintf("projects/%s/locations/%s/services/%s", c.project, c.region, name)
	
	service, err := c.runService.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %v", err)
	}

	return service, nil
}
