package mock

import (
	"time"

	"github.com/lpmourato/c9s/internal/model"
)

// MockServiceDetails returns a ServiceDetails struct with sample data for testing
func MockServiceDetails() *model.ServiceDetails {
	return &model.ServiceDetails{
		Name:           "sample-service",
		Region:         "us-central1",
		URL:            "https://sample-service-uc.a.run.app",
		LastUpdated:    time.Now().Add(-2 * time.Hour),
		Ready:          true,
		ActiveRevision: "sample-service-00002-abc",
		Traffic: []model.RevisionTraffic{
			{RevisionName: "sample-service-00002-abc", Percent: 80, Tag: "prod", Latest: true},
			{RevisionName: "sample-service-00001-def", Percent: 20, Tag: "canary", Latest: false},
		},
		UID:                  "123e4567-e89b-12d3-a456-426614174000",
		Generation:           2,
		CreationTime:         time.Now().Add(-24 * time.Hour),
		Creator:              "alice@example.com",
		LastModifier:         "bob@example.com",
		Labels:               map[string]string{"env": "production", "team": "devops"},
		Annotations:          map[string]string{"autoscaling.knative.dev/maxScale": "10"},
		ContainerImage:       "gcr.io/project/sample-image:v1.2.3",
		ImageDigest:          "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		CPU:                  "1",
		Memory:               "512Mi",
		Port:                 8080,
		ContainerName:        "app-container",
		ContainerConcurrency: 80,
		TimeoutSeconds:       300,
		EnvVars:              map[string]string{"API_KEY": "123456", "PASSWORD": "secret"},
		Secrets: []model.SecretMount{
			{
				Name:      "db-credentials",
				MountPath: "/secrets/db",
				Items: []model.SecretItem{
					{Key: "username", Path: "user.txt"},
					{Key: "password", Path: "pass.txt"},
				},
			},
		},
		Volumes: []model.VolumeMount{
			{Name: "config-vol", MountPath: "/config", ReadOnly: true, VolumeType: "configMap"},
		},
		MinInstances:    1,
		MaxInstances:    10,
		ServiceAccount:  "service-account@project.iam.gserviceaccount.com",
		VPCConnector:    "projects/project/locations/us-central1/connectors/my-vpc",
		VPCEgress:       "all-traffic",
		IngressSettings: "all",
		ExecutionEnv:    "gen2",
		CPUThrottling:   true,
		LivenessProbe: &model.HealthProbe{
			HTTPGet:             &model.HTTPGetAction{Path: "/healthz", Port: 8080, Scheme: "HTTP"},
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			TimeoutSeconds:      2,
			FailureThreshold:    3,
			SuccessThreshold:    1,
		},
		ReadinessProbe: &model.HealthProbe{
			HTTPGet:             &model.HTTPGetAction{Path: "/ready", Port: 8080, Scheme: "HTTP"},
			InitialDelaySeconds: 5,
			PeriodSeconds:       3,
			TimeoutSeconds:      1,
			FailureThreshold:    2,
			SuccessThreshold:    1,
		},
		StartupProbe: &model.HealthProbe{
			HTTPGet:             &model.HTTPGetAction{Path: "/startup", Port: 8080, Scheme: "HTTP"},
			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			TimeoutSeconds:      3,
			FailureThreshold:    5,
			SuccessThreshold:    1,
		},
		LaunchStage:          "GA",
		OperationID:          "op-987654321",
		LogURL:               "https://console.cloud.google.com/logs/viewer?project=project",
		SelfLink:             "projects/project/locations/us-central1/services/sample-service",
		LatestRevision:       "sample-service-00002-abc",
		LatestReadyRevision:  "sample-service-00002-abc",
		RevisionCreationTime: time.Now().Add(-1 * time.Hour),
		ContainerStatuses: []model.ContainerStatus{
			{Name: "app-container", ImageDigest: "sha256:abcdef...", Ready: true, RestartCount: 0},
			{Name: "sidecar", ImageDigest: "sha256:123456...", Ready: false, RestartCount: 2},
		},
		RevisionConditions: []model.Condition{
			{Type: "Ready", Status: "True", Reason: "ServiceReady", Message: "Service is ready."},
			{Type: "ResourcesAvailable", Status: "True", Reason: "ResourcesAvailable", Message: "All resources available."},
		},
	}
}
