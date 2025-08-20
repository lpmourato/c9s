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

// GetMockServices returns a list of mock services
func GetMockServices() []Service {
	now := time.Now()
	return []Service{
		&CloudRunService{
			Name:       "frontend-service",
			Region:     "us-central1",
			URL:        "https://frontend-service-hash.run.app",
			Status:     "Ready",
			LastDeploy: now.Add(-24 * time.Hour),
			Traffic:    "100%",
		},
		&CloudRunService{
			Name:       "backend-api",
			Region:     "us-central1",
			URL:        "https://backend-api-hash.run.app",
			Status:     "Ready",
			LastDeploy: now.Add(-48 * time.Hour),
			Traffic:    "100%",
		},
		&CloudRunService{
			Name:       "auth-service",
			Region:     "us-central1",
			URL:        "https://auth-service-hash.run.app",
			Status:     "Failed",
			LastDeploy: now.Add(-12 * time.Hour),
			Traffic:    "No traffic (failed)",
		},
		&CloudRunService{
			Name:       "worker-service",
			Region:     "us-east1",
			URL:        "https://worker-service-hash.run.app",
			Status:     "Updating",
			LastDeploy: now.Add(-1 * time.Hour),
			Traffic:    "v1 (90%), v2 (10%)",
		},
	}
}
