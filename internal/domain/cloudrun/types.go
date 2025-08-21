package cloudrun

import (
	"time"

	"github.com/lpmourato/c9s/internal/model"
)

// CloudRunService represents a Cloud Run service
type CloudRunService struct {
	Name       string
	Region     string
	URL        string
	Status     string
	LastDeploy time.Time
	Traffic    string
}

// Ensure CloudRunService implements model.Service
var _ model.Service = (*CloudRunService)(nil)

// Implementation of model.Service interface
func (s *CloudRunService) GetName() string          { return s.Name }
func (s *CloudRunService) GetRegion() string        { return s.Region }
func (s *CloudRunService) GetURL() string           { return s.URL }
func (s *CloudRunService) GetStatus() string        { return s.Status }
func (s *CloudRunService) GetLastDeploy() time.Time { return s.LastDeploy }
func (s *CloudRunService) GetTraffic() string       { return s.Traffic }

// CloudRunProvider defines the interface for getting Cloud Run services
type CloudRunProvider interface {
	GetServices() ([]model.Service, error)
	GetServicesByRegion(region string) ([]model.Service, error)
}
