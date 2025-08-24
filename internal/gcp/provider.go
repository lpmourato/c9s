package gcp

import (
	"github.com/lpmourato/c9s/internal/domain/cloudrun"
	"github.com/lpmourato/c9s/internal/logging"
	"github.com/lpmourato/c9s/internal/model"
)

type serviceProvider struct {
	projectID string
}

// NewServiceProvider creates a new GCP service provider
func NewServiceProvider(projectID string) (cloudrun.CloudRunProvider, error) {
	return &serviceProvider{
		projectID: projectID,
	}, nil
}

func (p *serviceProvider) GetServices() ([]model.Service, error) {
	// TODO: Implement
	return nil, nil
}

func (p *serviceProvider) GetServicesByRegion(region string) ([]model.Service, error) {
	// TODO: Implement
	return nil, nil
}

// NewLogStreamer creates a log streamer for a Cloud Run service
func (p *serviceProvider) NewLogStreamer(serviceName, region string) (model.LogStreamer, error) {
	provider, err := logging.NewGCPLogService(p.projectID, serviceName, region)
	if err != nil {
		return nil, err
	}

	opts := model.CloudProviderOptions{
		ProjectID:   p.projectID,
		ServiceName: serviceName,
		Region:      region,
	}

	return logging.NewLogService(provider, opts), nil
}
