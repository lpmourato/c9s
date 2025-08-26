package model

import "context"

// CloudProvider defines the interface for getting Cloud services
type CloudProvider interface {
	GetServices() ([]Service, error)
	GetServicesByRegion(region string) ([]Service, error)
	NewLogStreamer(serviceName, region string) (LogStreamer, error)
	GetServiceDetails(ctx context.Context, serviceName, region string) (*ServiceDetails, error)
}
