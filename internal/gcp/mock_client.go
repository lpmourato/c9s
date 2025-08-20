package gcp

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	run "google.golang.org/api/run/v2"
)

// MockClient implements a mock GCP client for testing
type MockClient struct {
	services map[string]*run.GoogleCloudRunV2Service
	mu       sync.RWMutex
}

// NewMockClient creates a new mock client with sample data
func NewMockClient() *MockClient {
	mock := &MockClient{
		services: make(map[string]*run.GoogleCloudRunV2Service),
	}

	// Add some sample services
	sampleServices := []struct {
		name      string
		region    string
		revisions []string
	}{
		{"web-frontend", "us-central1", []string{"web-frontend-00001", "web-frontend-00002"}},
		{"api-backend", "us-central1", []string{"api-backend-00001"}},
		{"auth-service", "us-east1", []string{"auth-service-00001", "auth-service-00002", "auth-service-00003"}},
		{"notification-service", "us-west1", []string{"notification-service-00001"}},
		{"analytics-service", "us-central1", []string{"analytics-service-00001"}},
	}

	for _, svc := range sampleServices {
		mock.services[svc.name] = createMockService(svc.name, svc.region, svc.revisions)
	}

	// Start service state simulation
	go mock.simulateServiceChanges()

	return mock
}

func createMockService(name, region string, revisions []string) *run.GoogleCloudRunV2Service {
	now := time.Now().Format(time.RFC3339)
	fullName := fmt.Sprintf("projects/mock-project/locations/%s/services/%s", region, name)

	traffic := make([]*run.GoogleCloudRunV2TrafficTarget, len(revisions))
	remainingPercent := 100
	for i, rev := range revisions {
		var percent int64
		if i == len(revisions)-1 {
			percent = int64(remainingPercent)
		} else {
			percent = int64(rand.Intn(remainingPercent))
			remainingPercent -= int(percent)
		}
		traffic[i] = &run.GoogleCloudRunV2TrafficTarget{
			Percent: percent,
			Revision: fmt.Sprintf("projects/mock-project/locations/%s/services/%s/revisions/%s",
				region, name, rev),
		}
	}

	conditions := []*run.GoogleCloudRunV2Condition{
		{
			Type:    "Ready",
			State:   "CONDITION_MET",
			Message: "Service is ready",
		},
		{
			Type:    "Serving",
			State:   "CONDITION_MET",
			Message: "Service is serving",
		},
	}

	return &run.GoogleCloudRunV2Service{
		Name:       fullName,
		Uri:        fmt.Sprintf("https://%s-%s.run.app", name, region),
		UpdateTime: now,
		Traffic:    traffic,
		Conditions: conditions,
	}
}

func (m *MockClient) simulateServiceChanges() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		// Randomly select a service to update
		for name, svc := range m.services {
			if rand.Float32() < 0.3 { // 30% chance of state change
				switch rand.Intn(4) {
				case 0: // Update traffic split
					m.updateTrafficSplit(svc)
				case 1: // Simulate deployment
					m.simulateDeployment(name, svc)
				case 2: // Toggle service state
					m.toggleServiceState(svc)
				case 3: // Update revision
					m.updateRevision(name, svc)
				}
			}
		}
		m.mu.Unlock()
	}
}

func (m *MockClient) updateTrafficSplit(svc *run.GoogleCloudRunV2Service) {
	if len(svc.Traffic) < 2 {
		return
	}
	// Randomly adjust traffic between revisions
	remainingPercent := int64(100)
	for i := 0; i < len(svc.Traffic)-1; i++ {
		percent := rand.Int63n(remainingPercent)
		svc.Traffic[i].Percent = percent
		remainingPercent -= percent
	}
	svc.Traffic[len(svc.Traffic)-1].Percent = remainingPercent
}

func (m *MockClient) simulateDeployment(name string, svc *run.GoogleCloudRunV2Service) {
	svc.Conditions = []*run.GoogleCloudRunV2Condition{
		{
			Type:    "Ready",
			State:   "CONDITION_PENDING",
			Message: "Deploying new revision",
		},
		{
			Type:    "Serving",
			State:   "CONDITION_PENDING",
			Message: "Deploying",
		},
	}
	// Add a new revision
	newRevision := fmt.Sprintf("%s-%05d", name, len(svc.Traffic)+1)
	svc.Traffic = append(svc.Traffic, &run.GoogleCloudRunV2TrafficTarget{
		Revision: newRevision,
		Percent:  0,
	})
}

func (m *MockClient) toggleServiceState(svc *run.GoogleCloudRunV2Service) {
	states := []string{"CONDITION_MET", "CONDITION_FAILED", "CONDITION_PENDING"}
	newState := states[rand.Intn(len(states))]
	messages := map[string]string{
		"CONDITION_MET":     "Service is ready",
		"CONDITION_FAILED":  "Service failed to start",
		"CONDITION_PENDING": "Service is updating",
	}

	for _, cond := range svc.Conditions {
		cond.State = newState
		cond.Message = messages[newState]
	}
}

func (m *MockClient) updateRevision(name string, svc *run.GoogleCloudRunV2Service) {
	svc.UpdateTime = time.Now().Format(time.RFC3339)
}

// ListServices implements the Client interface
func (m *MockClient) ListServices(ctx context.Context) ([]*run.GoogleCloudRunV2Service, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]*run.GoogleCloudRunV2Service, 0, len(m.services))
	for _, svc := range m.services {
		services = append(services, svc)
	}
	return services, nil
}

// GetService implements the Client interface
func (m *MockClient) GetService(ctx context.Context, name string) (*run.GoogleCloudRunV2Service, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Extract service name from full path
	parts := strings.Split(name, "/")
	serviceName := parts[len(parts)-1]

	if svc, ok := m.services[serviceName]; ok {
		return svc, nil
	}
	return nil, fmt.Errorf("service %s not found", name)
}
