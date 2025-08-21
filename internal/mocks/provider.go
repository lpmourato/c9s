package mocks

import (
	"time"

	"github.com/lpmourato/c9s/internal/domain/cloudrun"
)

type mockProvider struct{}

// NewMockProvider creates a new mock service provider
func NewMockProvider() cloudrun.ServiceProvider {
	return &mockProvider{}
}

func (p *mockProvider) GetServices() ([]cloudrun.Service, error) {
	return getMockServices(), nil
}

func (p *mockProvider) GetServicesByRegion(region string) ([]cloudrun.Service, error) {
	services := getMockServices()
	if region == "" {
		return services, nil
	}

	var filtered []cloudrun.Service
	for _, svc := range services {
		if svc.Region == region {
			filtered = append(filtered, svc)
		}
	}
	return filtered, nil
}

func getMockServices() []cloudrun.Service {
	now := time.Now()
	return []cloudrun.Service{
		{
			Name:       "service-a",
			Region:     "us-central1",
			URL:        "https://service-a-xyz.run.app",
			Status:     "Running",
			LastDeploy: now.Add(-24 * time.Hour),
			Traffic:    "100%",
		},
		{
			Name:       "service-b",
			Region:     "us-east1",
			URL:        "https://service-b-xyz.run.app",
			Status:     "Running",
			LastDeploy: now.Add(-48 * time.Hour),
			Traffic:    "100%",
		},
		{
			Name:       "service-c",
			Region:     "us-west1",
			URL:        "https://service-c-xyz.run.app",
			Status:     "Failed",
			LastDeploy: now.Add(-12 * time.Hour),
			Traffic:    "0%",
		},
	}
}
