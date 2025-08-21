package datasource

import "github.com/lpmourato/c9s/internal/model"

type mockDataSource struct {
	data []model.Service
}

func newMockDataSource(data []model.Service) DataSource {
	return &mockDataSource{data: data}
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
