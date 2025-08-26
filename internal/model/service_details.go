package model

import "time"

// ServiceDetails contains detailed information about a Cloud Run service
type ServiceDetails struct {
	Name           string
	Region         string
	URL            string
	LastUpdated    time.Time
	ContainerImage string
	CPU            string
	Memory         string
	Port           int32
	EnvVars        map[string]string
	MinInstances   int32
	MaxInstances   int32
	Ready          bool
	ActiveRevision string
	Traffic        []RevisionTraffic
}

// RevisionTraffic represents traffic allocation for a revision
type RevisionTraffic struct {
	RevisionName string
	Percent      int32
	Tag          string
	Latest       bool
}
