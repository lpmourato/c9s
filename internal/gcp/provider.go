package gcp

import (
	"github.com/lpmourato/c9s/internal/domain/cloudrun"
)

type serviceProvider struct {
	// GCP client fields will go here
}

// NewServiceProvider creates a new GCP service provider
func NewServiceProvider() (cloudrun.ServiceProvider, error) {
	// TODO: Initialize GCP client
	return &serviceProvider{}, nil
}

func (p *serviceProvider) GetServices() ([]cloudrun.Service, error) {
	// TODO: Implement
	return nil, nil
}

func (p *serviceProvider) GetServicesByRegion(region string) ([]cloudrun.Service, error) {
	// TODO: Implement
	return nil, nil
}
