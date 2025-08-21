package cloudrun_test

import (
	"time"

	"github.com/lpmourato/c9s/internal/domain/cloudrun"
)

// MockServiceProvider provides mock data for testing
type MockServiceProvider struct {
	services []cloudrun.Service
}

// NewMockServiceProvider creates a new mock provider with test data
func NewMockServiceProvider() *MockServiceProvider {
	return &MockServiceProvider{
		services: []cloudrun.Service{
			{
				Name:       "frontend-service",
				Region:     "us-central1",
				URL:        "https://frontend-service-hash.run.app",
				Status:     "Ready",
				LastDeploy: time.Now().Add(-24 * time.Hour),
				Traffic:    "100%",
			},
			{
				Name:       "backend-api",
				Region:     "us-central1",
				URL:        "https://backend-api-hash.run.app",
				Status:     "Ready",
				LastDeploy: time.Now().Add(-48 * time.Hour),
				Traffic:    "100%",
			},
			{
				Name:       "auth-service",
				Region:     "us-central1",
				URL:        "https://auth-service-hash.run.app",
				Status:     "Failed",
				LastDeploy: time.Now(),
				Traffic:    "No traffic (failed)",
			},
			{
				Name:       "worker-service",
				Region:     "us-east1",
				URL:        "https://worker-service-hash.run.app",
				Status:     "Updating",
				LastDeploy: time.Now().Add(-1 * time.Hour),
				Traffic:    "v1 (90%), v2 (10%)",
			},
		},
	}
}

// GetServices implements ServiceProvider interface
func (m *MockServiceProvider) GetServices() ([]cloudrun.Service, error) {
	return m.services, nil
}

// GetServicesByRegion implements ServiceProvider interface
func (m *MockServiceProvider) GetServicesByRegion(region string) ([]cloudrun.Service, error) {
	var filtered []cloudrun.Service
	for _, svc := range m.services {
		if svc.Region == region {
			filtered = append(filtered, svc)
		}
	}
	return filtered, nil
}
