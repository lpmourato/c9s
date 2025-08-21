package datasource

import (
	"encoding/json"
	"os"

	"github.com/lpmourato/c9s/internal/domain/cloudrun"
	"github.com/lpmourato/c9s/internal/model"
)

type jsonDataSource struct {
	filePath string
}

func newJSONDataSource(filePath string) DataSource {
	return &jsonDataSource{filePath: filePath}
}

func (ds *jsonDataSource) GetServices() ([]model.Service, error) {
	data, err := os.ReadFile(ds.filePath)
	if err != nil {
		return nil, err
	}

	var services []cloudrun.CloudRunService
	if err := json.Unmarshal(data, &services); err != nil {
		return nil, err
	}

	// Convert to model.Service interface
	result := make([]model.Service, len(services))
	for i := range services {
		result[i] = &services[i]
	}
	return result, nil
}

func (ds *jsonDataSource) GetServicesByRegion(region string) ([]model.Service, error) {
	services, err := ds.GetServices()
	if err != nil {
		return nil, err
	}

	if region == "" {
		return services, nil
	}

	var filtered []model.Service
	for _, svc := range services {
		if svc.GetRegion() == region {
			filtered = append(filtered, svc)
		}
	}
	return filtered, nil
}
