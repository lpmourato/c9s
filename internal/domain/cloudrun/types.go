package cloudrun

import (
	"github.com/lpmourato/c9s/internal/model"
)

// CloudRunProvider defines the interface for getting Cloud Run services
type CloudRunProvider interface {
	GetServices() ([]model.Service, error)
	GetServicesByRegion(region string) ([]model.Service, error)
	NewLogStreamer(serviceName, region string) (model.LogStreamer, error)
}
