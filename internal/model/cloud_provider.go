package model

// CloudProvider defines the interface for getting Cloud services
type CloudProvider interface {
	GetServices() ([]Service, error)
	GetServicesByRegion(region string) ([]Service, error)
	NewLogStreamer(serviceName, region string) (LogStreamer, error)
}
