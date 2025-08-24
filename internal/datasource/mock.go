package datasource

import (
	"github.com/lpmourato/c9s/internal/domain/cloudrun"
	"github.com/lpmourato/c9s/internal/mock"
	"github.com/lpmourato/c9s/internal/model"
)

type mockDataSource struct {
	data     []model.Service
	provider cloudrun.CloudRunProvider
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

func (ds *mockDataSource) GetProvider() cloudrun.CloudRunProvider {
	return ds.provider
}

// mockProvider implements cloudrun.CloudRunProvider for testing
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
