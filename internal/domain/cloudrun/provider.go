package cloudrun

import (
	"context"

	"github.com/lpmourato/c9s/internal/model"
)

// Provider implements CloudRunProvider for GCP
type Provider struct {
	delegate model.CloudRunProvider
}

// NewProvider creates a new GCP cloud run provider
func NewProvider(delegate model.CloudRunProvider) *Provider {
	return &Provider{
		delegate: delegate,
	}
}

// GetServices implements CloudRunProvider
func (p *Provider) GetServices() ([]model.Service, error) {
	return p.delegate.GetServices()
}

// GetServicesByRegion implements CloudRunProvider
func (p *Provider) GetServicesByRegion(region string) ([]model.Service, error) {
	return p.delegate.GetServicesByRegion(region)
}

// NewLogStreamer implements CloudRunProvider
func (p *Provider) NewLogStreamer(serviceName, region string) (model.LogStreamer, error) {
	return p.delegate.NewLogStreamer(serviceName, region)
}

// GetServiceDetails implements CloudRunProvider
func (p *Provider) GetServiceDetails(ctx context.Context, serviceName, region string) (*model.ServiceDetails, error) {
	return p.delegate.GetServiceDetails(ctx, serviceName, region)
}
