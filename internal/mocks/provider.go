package mocks

import (
	"time"

	"github.com/lpmourato/c9s/internal/domain/cloudrun"
	"github.com/lpmourato/c9s/internal/model"
)

type mockProvider struct{}

// NewMockProvider creates a new mock service provider
func NewMockProvider() cloudrun.CloudRunProvider {
	return &mockProvider{}
}

func (p *mockProvider) GetServices() ([]model.Service, error) {
	services := getMockServices()
	result := make([]model.Service, len(services))
	for i, svc := range services {
		result[i] = &svc
	}
	return result, nil
}

func (p *mockProvider) GetServicesByRegion(region string) ([]model.Service, error) {
	services := getMockServices()
	if region == "" {
		result := make([]model.Service, len(services))
		for i, svc := range services {
			result[i] = &svc
		}
		return result, nil
	}

	var filtered []model.Service
	for _, svc := range services {
		if svc.Region == region {
			filtered = append(filtered, &svc)
		}
	}
	return filtered, nil
}

func getMockServices() []cloudrun.CloudRunService {
	now := time.Now()
	return []cloudrun.CloudRunService{
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
