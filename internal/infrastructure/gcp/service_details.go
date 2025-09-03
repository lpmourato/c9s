package gcp

import (
	"context"
	"fmt"
	"time"

	"github.com/lpmourato/c9s/internal/model"
	run "google.golang.org/api/run/v1"
)

// GetServiceDetails fetches detailed information about a Cloud Run service
func (p *serviceProvider) GetServiceDetails(ctx context.Context, serviceName, region string) (*model.ServiceDetails, error) {
	// Initialize Cloud Run client if not already initialized
	runClient, err := run.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run client: %v", err)
	}

	// Build the full service name
	name := fmt.Sprintf("projects/%s/locations/%s/services/%s", p.projectID, region, serviceName)

	// Get service details
	service, err := runClient.Projects.Locations.Services.Get(name).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get service details: %v", err)
	}

	// Parse creation timestamp
	creationTime, _ := time.Parse(time.RFC3339, service.Metadata.CreationTimestamp)
	lastUpdated := creationTime
	if service.Metadata.Annotations["serving.knative.dev/lastModifier"] != "" {
		// Use creation time as fallback for last updated
		lastUpdated = creationTime
	}

	// Extract labels and annotations
	labels := make(map[string]string)
	if service.Metadata.Labels != nil {
		for k, v := range service.Metadata.Labels {
			labels[k] = v
		}
	}

	annotations := make(map[string]string)
	if service.Metadata.Annotations != nil {
		for k, v := range service.Metadata.Annotations {
			annotations[k] = v
		}
	}

	// Extract environment variables and secrets
	envVars := make(map[string]string)
	var secrets []model.SecretMount
	var volumes []model.VolumeMount
	var containerName string
	var containerImage string
	var imageDigest string
	var cpu, memory string
	var port int32
	var concurrency int32 = 80 // Default value
	var timeout int32 = 300    // Default value
	var livenessProbe, readinessProbe, startupProbe *model.HealthProbe

	if service.Spec.Template != nil && len(service.Spec.Template.Spec.Containers) > 0 {
		container := service.Spec.Template.Spec.Containers[0]
		containerName = container.Name
		containerImage = container.Image

		// Extract environment variables
		for _, env := range container.Env {
			if env.Value != "" {
				envVars[env.Name] = env.Value
			} else if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				envVars[env.Name] = fmt.Sprintf("Secret: %s/%s", env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key)
			}
		}

		// Extract resource limits
		if container.Resources != nil && container.Resources.Limits != nil {
			if cpuVal, ok := container.Resources.Limits["cpu"]; ok {
				cpu = cpuVal
			}
			if memVal, ok := container.Resources.Limits["memory"]; ok {
				memory = memVal
			}
		}

		// Extract port
		if len(container.Ports) > 0 {
			port = int32(container.Ports[0].ContainerPort)
		}

		// Extract volume mounts
		for _, vm := range container.VolumeMounts {
			volumes = append(volumes, model.VolumeMount{
				Name:      vm.Name,
				MountPath: vm.MountPath,
				ReadOnly:  vm.ReadOnly,
			})
		}

		// Extract health probes
		if container.LivenessProbe != nil && container.LivenessProbe.HttpGet != nil {
			livenessProbe = &model.HealthProbe{
				HTTPGet: &model.HTTPGetAction{
					Path: container.LivenessProbe.HttpGet.Path,
					Port: int32(container.LivenessProbe.HttpGet.Port),
				},
				InitialDelaySeconds: int32(container.LivenessProbe.InitialDelaySeconds),
				PeriodSeconds:       int32(container.LivenessProbe.PeriodSeconds),
				TimeoutSeconds:      int32(container.LivenessProbe.TimeoutSeconds),
				FailureThreshold:    int32(container.LivenessProbe.FailureThreshold),
			}
		}

		if container.ReadinessProbe != nil && container.ReadinessProbe.HttpGet != nil {
			readinessProbe = &model.HealthProbe{
				HTTPGet: &model.HTTPGetAction{
					Path: container.ReadinessProbe.HttpGet.Path,
					Port: int32(container.ReadinessProbe.HttpGet.Port),
				},
				InitialDelaySeconds: int32(container.ReadinessProbe.InitialDelaySeconds),
				PeriodSeconds:       int32(container.ReadinessProbe.PeriodSeconds),
				TimeoutSeconds:      int32(container.ReadinessProbe.TimeoutSeconds),
				FailureThreshold:    int32(container.ReadinessProbe.FailureThreshold),
			}
		}

		if container.StartupProbe != nil && container.StartupProbe.HttpGet != nil {
			startupProbe = &model.HealthProbe{
				HTTPGet: &model.HTTPGetAction{
					Path: container.StartupProbe.HttpGet.Path,
					Port: int32(container.StartupProbe.HttpGet.Port),
				},
				InitialDelaySeconds: int32(container.StartupProbe.InitialDelaySeconds),
				PeriodSeconds:       int32(container.StartupProbe.PeriodSeconds),
				TimeoutSeconds:      int32(container.StartupProbe.TimeoutSeconds),
				FailureThreshold:    int32(container.StartupProbe.FailureThreshold),
			}
		}
	}

	// Extract template-level configuration
	if service.Spec.Template != nil {
		if service.Spec.Template.Spec.ContainerConcurrency > 0 {
			concurrency = int32(service.Spec.Template.Spec.ContainerConcurrency)
		}
		if service.Spec.Template.Spec.TimeoutSeconds > 0 {
			timeout = int32(service.Spec.Template.Spec.TimeoutSeconds)
		}

		// Extract volumes and secrets
		for _, vol := range service.Spec.Template.Spec.Volumes {
			if vol.Secret != nil {
				secret := model.SecretMount{
					Name:      vol.Name,
					MountPath: "", // Will be filled from volume mounts
				}
				for _, item := range vol.Secret.Items {
					secret.Items = append(secret.Items, model.SecretItem{
						Key:  item.Key,
						Path: item.Path,
					})
				}
				secrets = append(secrets, secret)
			}
		}
	}

	// Extract scaling configuration from annotations
	var minInstances, maxInstances int32
	if annotations["autoscaling.knative.dev/minScale"] != "" {
		fmt.Sscanf(annotations["autoscaling.knative.dev/minScale"], "%d", &minInstances)
	}
	if annotations["autoscaling.knative.dev/maxScale"] != "" {
		fmt.Sscanf(annotations["autoscaling.knative.dev/maxScale"], "%d", &maxInstances)
	}

	// Extract network and security settings
	serviceAccount := service.Spec.Template.Spec.ServiceAccountName
	vpcConnector := annotations["run.googleapis.com/vpc-access-connector"]
	vpcEgress := annotations["run.googleapis.com/vpc-access-egress"]
	ingressSettings := annotations["run.googleapis.com/ingress"]
	executionEnv := annotations["run.googleapis.com/execution-environment"]
	cpuThrottling := annotations["run.googleapis.com/cpu-throttling"] == "true"

	// Build traffic information
	var traffic []model.RevisionTraffic
	for _, t := range service.Status.Traffic {
		traffic = append(traffic, model.RevisionTraffic{
			RevisionName: t.RevisionName,
			Percent:      int32(t.Percent),
			Tag:          t.Tag,
			Latest:       t.RevisionName == service.Status.LatestReadyRevisionName,
		})
	}

	// Build conditions
	var conditions []model.Condition
	for _, cond := range service.Status.Conditions {
		condTime, _ := time.Parse(time.RFC3339, cond.LastTransitionTime)
		conditions = append(conditions, model.Condition{
			Type:               cond.Type,
			Status:             cond.Status,
			LastTransitionTime: condTime,
			Reason:             cond.Reason,
			Message:            cond.Message,
		})
	}

	// Extract revision creation time
	revisionCreationTime := creationTime
	if service.Status.LatestCreatedRevisionName != "" {
		// For now, use service creation time as fallback
		revisionCreationTime = creationTime
	}

	// Build log URL
	logURL := fmt.Sprintf("https://console.cloud.google.com/logs/viewer?project=%s&resource=cloud_run_revision/service_name/%s", p.projectID, serviceName)

	// Create service details with all the new fields
	details := &model.ServiceDetails{
		// Basic Service Information
		Name:           serviceName,
		Region:         region,
		URL:            service.Status.Url,
		LastUpdated:    lastUpdated,
		Ready:          service.Status.ObservedGeneration > 0,
		ActiveRevision: service.Status.LatestReadyRevisionName,
		Traffic:        traffic,

		// Service Metadata
		UID:          service.Metadata.Uid,
		Generation:   service.Metadata.Generation,
		CreationTime: creationTime,
		Creator:      annotations["serving.knative.dev/creator"],
		LastModifier: annotations["serving.knative.dev/lastModifier"],
		Labels:       labels,
		Annotations:  annotations,

		// Container Configuration
		ContainerImage:       containerImage,
		ImageDigest:          imageDigest,
		CPU:                  cpu,
		Memory:               memory,
		Port:                 port,
		ContainerName:        containerName,
		ContainerConcurrency: concurrency,
		TimeoutSeconds:       timeout,

		// Environment & Secrets
		EnvVars: envVars,
		Secrets: secrets,
		Volumes: volumes,

		// Scaling Configuration
		MinInstances: minInstances,
		MaxInstances: maxInstances,

		// Network & Security
		ServiceAccount:  serviceAccount,
		VPCConnector:    vpcConnector,
		VPCEgress:       vpcEgress,
		IngressSettings: ingressSettings,
		ExecutionEnv:    executionEnv,
		CPUThrottling:   cpuThrottling,

		// Health Checks
		LivenessProbe:  livenessProbe,
		ReadinessProbe: readinessProbe,
		StartupProbe:   startupProbe,

		// Additional Metadata
		LaunchStage: annotations["run.googleapis.com/launch-stage"],
		OperationID: annotations["run.googleapis.com/operation-id"],
		LogURL:      logURL,
		SelfLink:    service.Metadata.SelfLink,

		// Revision Details
		LatestRevision:       service.Status.LatestCreatedRevisionName,
		LatestReadyRevision:  service.Status.LatestReadyRevisionName,
		RevisionCreationTime: revisionCreationTime,
		ContainerStatuses:    []model.ContainerStatus{}, // TODO: Get from revision details
		RevisionConditions:   conditions,
	}

	return details, nil
}
