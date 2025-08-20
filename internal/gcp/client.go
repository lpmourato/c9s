package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/option"
	run "google.golang.org/api/run/v2"
)

// Client represents a GCP Cloud Run client
type Client struct {
	runService       *run.Service
	servicesService  *run.ProjectsLocationsServicesService
	locationsService *run.ProjectsLocationsService
	project          string
	region           string
	mock             *MockClient // For testing/demo purposes
}

// NewClient creates a new GCP Cloud Run client
func NewClient(ctx context.Context, project, region string) (*Client, error) {
	runClient, err := run.NewService(ctx, option.WithScopes(run.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run client: %v", err)
	}

	return &Client{
		runService:       runClient,
		servicesService:  runClient.Projects.Locations.Services,
		locationsService: runClient.Projects.Locations,
		project:          project,
		region:           region,
	}, nil
}

// NewTestClient creates a new client with mock data
func NewTestClient(project, region string) *Client {
	return &Client{
		project: project,
		region:  region,
		mock:    NewMockClient(),
	}
}

// Project returns the configured project ID
func (c *Client) Project() string {
	return c.project
}

// Region returns the configured region
func (c *Client) Region() string {
	return c.region
}

// ListServices returns all Cloud Run services across all regions in the configured project
func (c *Client) ListServices(ctx context.Context) ([]*run.GoogleCloudRunV2Service, error) {
	if c.mock != nil {
		return c.mock.ListServices(ctx)
	}

	// Use a fresh context with timeout for list operation
	freshCtx, cancel := context.WithTimeout(ctx, 10*time.Second) // Increased timeout for multi-region
	defer cancel()

	// For now, use a predefined list of regions where Cloud Run is typically available
	regions := []string{
		"us-central1",
		"us-east1",
		"us-east4",
		"us-west1",
		"us-west2",
		"us-west3",
		"us-west4",
		"europe-west1",
		"europe-west2",
		"europe-west3",
		"europe-west4",
		"europe-west6",
		"europe-central2",
		"asia-east1",
		"asia-east2",
		"asia-northeast1",
		"asia-northeast2",
		"asia-northeast3",
		"asia-southeast1",
		"asia-southeast2",
		"asia-south1",
		"australia-southeast1",
		"australia-southeast2",
		"northamerica-northeast1",
		"southamerica-east1",
	}

	parent := fmt.Sprintf("projects/%s", c.project)

	// Create channels for collecting results
	type regionResult struct {
		services []*run.GoogleCloudRunV2Service
		err      error
		region   string
	}
	results := make(chan regionResult, len(regions))

	// Query each region in parallel
	for _, region := range regions {
		go func(region string) {
			regionParent := fmt.Sprintf("%s/locations/%s", parent, region)
			req := c.servicesService.List(regionParent).Context(freshCtx)

			// Add headers to prevent caching
			req.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			req.Header().Set("Pragma", "no-cache")
			req.Header().Set("X-Goog-Api-Client", fmt.Sprintf("c9s/%d", time.Now().UnixNano()))

			// Get all fields
			req.Fields("*")

			resp, err := req.Do()
			if err != nil {
				results <- regionResult{err: fmt.Errorf("failed to list services in %s: %v", region, err), region: region}
				return
			}
			if resp.Services != nil {
				results <- regionResult{services: resp.Services, region: region}
			} else {
				results <- regionResult{services: []*run.GoogleCloudRunV2Service{}, region: region}
			}
		}(region)
	}

	// Collect all results
	var allServices []*run.GoogleCloudRunV2Service
	var errors []string

	for i := 0; i < len(regions); i++ {
		result := <-results
		if result.err != nil {
			errors = append(errors, result.err.Error())
			continue
		}
		allServices = append(allServices, result.services...)
	}

	// Return error if no services found anywhere
	if len(allServices) == 0 {
		if len(errors) > 0 {
			return nil, fmt.Errorf("errors listing services: %s", strings.Join(errors, "; "))
		}
		return nil, fmt.Errorf("no services found in project %s", c.project)
	}

	return allServices, nil
}

// getConditionState returns the state of a specific condition
func getConditionState(conditions []*run.GoogleCloudRunV2Condition, conditionType string) string {
	for _, c := range conditions {
		if c.Type == conditionType {
			return c.State
		}
	}
	return "UNKNOWN"
}

// GetService retrieves a specific Cloud Run service
func (c *Client) GetService(ctx context.Context, name string) (*run.GoogleCloudRunV2Service, error) {
	// If using mock client, delegate to mock implementation
	if c.mock != nil {
		return c.mock.GetService(ctx, name)
	}

	// If name is already a full name (projects/*/locations/*/services/*), use it directly
	var fullName string
	if strings.HasPrefix(name, "projects/") {
		fullName = name
	} else {
		fullName = fmt.Sprintf("projects/%s/locations/%s/services/%s", c.project, c.region, name)
	}

	// Call with fresh context and ensure we don't get cached responses
	freshCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	call := c.servicesService.Get(fullName).Context(freshCtx)

	// Add special headers to bypass caching at various levels
	call.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	call.Header().Set("Pragma", "no-cache")
	call.Header().Set("X-Goog-Api-Client", fmt.Sprintf("c9s/%d", time.Now().UnixNano()))
	call.Header().Set("X-Goog-Request-Params", fmt.Sprintf("name=%s&time=%d", fullName, time.Now().UnixNano()))

	// Add field selector to ensure we get all fields
	call.Fields("*")

	service, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %v", err)
	}

	return service, nil
}
