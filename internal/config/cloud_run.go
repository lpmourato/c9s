package config

// CloudRunConfig holds configuration for Cloud Run view
type CloudRunConfig struct {
	ProjectID string
	Region    string
}

// NewCloudRunConfig creates a new configuration with default values
func NewCloudRunConfig() *CloudRunConfig {
	return &CloudRunConfig{
		ProjectID: "", // Empty by default
		Region:    "", // Empty by default
	}
}
