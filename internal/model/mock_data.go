package model

import "time"

// GetDefaultMockData returns the default set of mock services for testing
func GetDefaultMockData() []Service {
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
