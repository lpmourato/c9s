package cloudrun

import "time"

// Service represents a Cloud Run service
type Service struct {
	Name       string
	Region     string
	URL        string
	Status     string
	LastDeploy time.Time
	Traffic    string
}

// ServiceProvider defines the interface for getting Cloud Run services
type ServiceProvider interface {
	GetServices() ([]Service, error)
	GetServicesByRegion(region string) ([]Service, error)
}
