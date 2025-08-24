package mock

import (
	"time"

	"github.com/lpmourato/c9s/internal/model"
)

// GetDefaultServices returns the default set of mock services for testing
func GetDefaultServices() []model.Service {
	now := time.Now()
	return []model.Service{
		&model.CloudRunService{
			Name:       "frontend-service",
			Region:     "us-central1",
			URL:        "https://frontend-service-hash.run.app",
			Status:     "Ready",
			LastDeploy: now.Add(-24 * time.Hour),
			Traffic:    "100%",
		},
		&model.CloudRunService{
			Name:       "backend-api",
			Region:     "us-central1",
			URL:        "https://backend-api-hash.run.app",
			Status:     "Ready",
			LastDeploy: now.Add(-48 * time.Hour),
			Traffic:    "100%",
		},
		&model.CloudRunService{
			Name:       "auth-service",
			Region:     "us-central1",
			URL:        "https://auth-service-hash.run.app",
			Status:     "Failed",
			LastDeploy: now.Add(-12 * time.Hour),
			Traffic:    "No traffic (failed)",
		},
		&model.CloudRunService{
			Name:       "worker-service",
			Region:     "us-east1",
			URL:        "https://worker-service-hash.run.app",
			Status:     "Updating",
			LastDeploy: now.Add(-1 * time.Hour),
			Traffic:    "v1 (90%), v2 (10%)",
		},
	}
}
