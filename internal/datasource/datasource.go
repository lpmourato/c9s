package datasource

import (
	"fmt"

	"github.com/lpmourato/c9s/internal/domain/cloudrun"
	"github.com/lpmourato/c9s/internal/model"
)

// Type represents the type of data source
type Type string

const (
	// Mock represents a mock data source
	Mock Type = "mock"
	// GCP represents Google Cloud Platform data source
	GCP Type = "gcp"
)

// Config holds configuration for any data source
type Config struct {
	Type       Type
	ProjectID  string
	Region     string
	MockedData []model.Service
}

// DataSource defines the interface for getting data
type DataSource interface {
	// GetServices returns all services
	GetServices() ([]model.Service, error)
	// GetServicesByRegion returns services filtered by region
	GetServicesByRegion(region string) ([]model.Service, error)
	// GetProvider returns the cloud run provider
	GetProvider() cloudrun.CloudRunProvider
}

// Factory creates and returns a DataSource based on config
func Factory(cfg *Config) (DataSource, error) {
	switch cfg.Type {
	case Mock:
		return newMockDataSource(cfg.MockedData), nil
	case GCP:
		return newGCPDataSource(cfg.ProjectID)
	default:
		return nil, fmt.Errorf("unsupported data source type: %s", cfg.Type)
	}
}
