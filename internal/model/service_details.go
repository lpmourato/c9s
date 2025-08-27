package model

import "time"

// ServiceDetails contains detailed information about a Cloud Run service
type ServiceDetails struct {
	// Basic Service Information
	Name           string
	Region         string
	URL            string
	LastUpdated    time.Time
	Ready          bool
	ActiveRevision string
	Traffic        []RevisionTraffic

	// Service Metadata
	UID          string
	Generation   int64
	CreationTime time.Time
	Creator      string
	LastModifier string
	Labels       map[string]string
	Annotations  map[string]string

	// Container Configuration
	ContainerImage       string
	ImageDigest          string
	CPU                  string
	Memory               string
	Port                 int32
	ContainerName        string
	ContainerConcurrency int32
	TimeoutSeconds       int32

	// Environment & Secrets
	EnvVars map[string]string
	Secrets []SecretMount
	Volumes []VolumeMount

	// Scaling Configuration
	MinInstances int32
	MaxInstances int32

	// Network & Security
	ServiceAccount  string
	VPCConnector    string
	VPCEgress       string
	IngressSettings string
	ExecutionEnv    string
	CPUThrottling   bool

	// Health Checks
	LivenessProbe  *HealthProbe
	ReadinessProbe *HealthProbe
	StartupProbe   *HealthProbe

	// Additional Metadata
	LaunchStage string
	OperationID string
	LogURL      string
	SelfLink    string

	// Revision Details
	LatestRevision       string
	LatestReadyRevision  string
	RevisionCreationTime time.Time
	ContainerStatuses    []ContainerStatus
	RevisionConditions   []Condition
}

// RevisionTraffic represents traffic allocation for a revision
type RevisionTraffic struct {
	RevisionName string
	Percent      int32
	Tag          string
	Latest       bool
}

// SecretMount represents a mounted secret
type SecretMount struct {
	Name      string
	MountPath string
	Items     []SecretItem
}

// SecretItem represents an item in a secret
type SecretItem struct {
	Key  string
	Path string
}

// VolumeMount represents a mounted volume
type VolumeMount struct {
	Name       string
	MountPath  string
	ReadOnly   bool
	VolumeType string // secret, configMap, etc.
}

// HealthProbe represents health check configuration
type HealthProbe struct {
	HTTPGet             *HTTPGetAction
	InitialDelaySeconds int32
	PeriodSeconds       int32
	TimeoutSeconds      int32
	FailureThreshold    int32
	SuccessThreshold    int32
}

// HTTPGetAction represents HTTP GET health check
type HTTPGetAction struct {
	Path    string
	Port    int32
	Scheme  string
	Headers map[string]string
}

// ContainerStatus represents the status of a container
type ContainerStatus struct {
	Name         string
	ImageDigest  string
	Ready        bool
	RestartCount int32
}

// Condition represents a service or revision condition
type Condition struct {
	Type               string
	Status             string
	LastTransitionTime time.Time
	Reason             string
	Message            string
}
