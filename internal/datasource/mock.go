package datasource

import (
	"context"
	"time"

	"github.com/lpmourato/c9s/internal/mock"
	"github.com/lpmourato/c9s/internal/model"
)

type mockDataSource struct {
	data     []model.Service
	provider model.CloudRunProvider
}

func newMockDataSource(data []model.Service) DataSource {
	provider := &mockProvider{serviceName: "mock-service"}
	return &mockDataSource{
		data:     data,
		provider: provider,
	}
}

func (ds *mockDataSource) GetServices() ([]model.Service, error) {
	return ds.data, nil
}

func (ds *mockDataSource) GetServicesByRegion(region string) ([]model.Service, error) {
	if region == "" {
		return ds.data, nil
	}

	var filtered []model.Service
	for _, svc := range ds.data {
		if svc.GetRegion() == region {
			filtered = append(filtered, svc)
		}
	}
	return filtered, nil
}

func (ds *mockDataSource) GetProvider() model.CloudRunProvider {
	return ds.provider
}

// mockProvider implements model.CloudRunProvider for testing
type mockProvider struct {
	serviceName string
}

func (p *mockProvider) GetServices() ([]model.Service, error) {
	return nil, nil
}

func (p *mockProvider) GetServicesByRegion(region string) ([]model.Service, error) {
	return nil, nil
}

func (p *mockProvider) NewLogStreamer(serviceName, region string) (model.LogStreamer, error) {
	return mock.NewLogStreamer(serviceName), nil
}

func (p *mockProvider) GetServiceDetails(ctx context.Context, serviceName, region string) (*model.ServiceDetails, error) {
	return &model.ServiceDetails{
		Name:           serviceName,
		Region:         region,
		URL:            "https://" + serviceName + ".run.app",
		LastUpdated:    time.Now(),
		ContainerImage: "gcr.io/mock/image:latest",
		CPU:            "1000m",
		Memory:         "512Mi",
		Port:           8080,
		EnvVars: map[string]string{
			"ENV":     "mock",
			"VERSION": "1.0.0",
		},
		MinInstances:   0,
		MaxInstances:   10,
		Ready:          true,
		ActiveRevision: serviceName + "-00001",
		Traffic: []model.RevisionTraffic{
			{
				RevisionName: serviceName + "-00001",
				Percent:      100,
				Tag:          "",
				Latest:       true,
			},
		},
	}, nil
}
