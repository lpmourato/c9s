package datasource

import (
	"github.com/lpmourato/c9s/internal/domain/cloudrun"
	"github.com/lpmourato/c9s/internal/model"
)

type gcpDataSource struct {
	projectID string
	client    *cloudrun.CloudRunProvider
}

func newGCPDataSource(projectID string) DataSource {
	return &gcpDataSource{projectID: projectID}
}

func (ds *gcpDataSource) GetServices() ([]model.Service, error) {
	// TODO: Implement GCP client
	return nil, nil
}

func (ds *gcpDataSource) GetServicesByRegion(region string) ([]model.Service, error) {
	// TODO: Implement GCP client
	return nil, nil
}
