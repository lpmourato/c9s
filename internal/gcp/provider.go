package gcp

import (
	"github.com/lpmourato/c9s/internal/domain/cloudrun"
	"github.com/lpmourato/c9s/internal/model"
)

type serviceProvider struct {
	// GCP client fields will go here
}

// NewServiceProvider creates a new GCP service provider
func NewServiceProvider() (cloudrun.CloudRunProvider, error) {
	// TODO: Initialize GCP client
	return &serviceProvider{}, nil
}

func (p *serviceProvider) GetServices() ([]model.Service, error) {
	// TODO: Implement
	return nil, nil
}

func (p *serviceProvider) GetServicesByRegion(region string) ([]model.Service, error) {
	// TODO: Implement
	return nil, nil
}
