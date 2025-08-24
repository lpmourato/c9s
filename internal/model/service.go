package model

import "time"

// Service represents a Cloud Run service
type Service interface {
	GetName() string
	GetRegion() string
	GetURL() string
	GetStatus() string
	GetLastDeploy() time.Time
	GetTraffic() string
}

// CloudRunService implements Service interface
type CloudRunService struct {
	Name       string
	Region     string
	URL        string
	Status     string
	LastDeploy time.Time
	Traffic    string
}

func (s *CloudRunService) GetName() string          { return s.Name }
func (s *CloudRunService) GetRegion() string        { return s.Region }
func (s *CloudRunService) GetURL() string           { return s.URL }
func (s *CloudRunService) GetStatus() string        { return s.Status }
func (s *CloudRunService) GetLastDeploy() time.Time { return s.LastDeploy }
func (s *CloudRunService) GetTraffic() string       { return s.Traffic }
