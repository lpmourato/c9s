package gcp

import (
	"context"
	"fmt"
	"time"

	"github.com/lpmourato/c9s/internal/model"
	run "google.golang.org/api/run/v1"
)

// GetServiceDetails fetches detailed information about a Cloud Run service
func (p *serviceProvider) GetServiceDetails(ctx context.Context, serviceName, region string) (*model.ServiceDetails, error) {
	// Initialize Cloud Run client if not already initialized
	runClient, err := run.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run client: %v", err)
	}

	// Build the full service name
	name := fmt.Sprintf("projects/%s/locations/%s/services/%s", p.projectID, region, serviceName)

	// Get service details
	service, err := runClient.Projects.Locations.Services.Get(name).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get service details: %v", err)
	}

	// Extract environment variables
	envVars := make(map[string]string)
	if service.Spec.Template != nil && len(service.Spec.Template.Spec.Containers) > 0 {
		for _, env := range service.Spec.Template.Spec.Containers[0].Env {
			envVars[env.Name] = env.Value
		}
	}

	// Extract container details
	var cpu, memory string
	var port int32
	if service.Spec.Template != nil && len(service.Spec.Template.Spec.Containers) > 0 {
		container := service.Spec.Template.Spec.Containers[0]
		if container.Resources != nil && container.Resources.Limits != nil {
			cpu = container.Resources.Limits["cpu"]
			memory = container.Resources.Limits["memory"]
		}
		if len(container.Ports) > 0 {
			port = int32(container.Ports[0].ContainerPort)
		}
	}

	// Build traffic information
	var traffic []model.RevisionTraffic
	for _, t := range service.Status.Traffic {
		traffic = append(traffic, model.RevisionTraffic{
			RevisionName: t.RevisionName,
			Percent:      int32(t.Percent),
			Tag:          t.Tag,
			Latest:       t.RevisionName == service.Status.LatestReadyRevisionName,
		})
	}

	// Create service details
	details := &model.ServiceDetails{
		Name:           serviceName,
		Region:         region,
		URL:            service.Status.Url,
		LastUpdated:    time.Now(), // TODO: Parse from service.Metadata.CreationTimestamp,
		ContainerImage: service.Spec.Template.Spec.Containers[0].Image,
		CPU:            cpu,
		Memory:         memory,
		Port:           int32(port),
		EnvVars:        envVars,
		MinInstances:   int32(service.Spec.Template.Spec.ContainerConcurrency),
		MaxInstances:   0, // TODO: Get this from annotations or other source
		Ready:          service.Status.ObservedGeneration > 0,
		ActiveRevision: service.Status.LatestReadyRevisionName,
		Traffic:        traffic,
	}

	return details, nil
}
