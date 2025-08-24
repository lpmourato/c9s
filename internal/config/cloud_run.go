// Package config provides configuration structures for different parts of the application.
// It centralizes all configuration-related types and utilities, making it easier to manage
// application settings and ensure consistency across different components.
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
