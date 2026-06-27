package health

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status represents the health status.
type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

// HealthCheck represents the health check response.
type HealthCheck struct {
	Status    Status            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Components map[string]Component `json:"components"`
}

// Component represents a component's health status.
type Component struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}

// Checker performs health checks.
type Checker struct {
	mu         sync.RWMutex
	checks     map[string]CheckFunc
	port       string
}

// CheckFunc is a function that checks a component's health.
type CheckFunc func() Component

// NewChecker creates a new health checker.
func NewChecker(port string) *Checker {
	return &Checker{
		checks: make(map[string]CheckFunc),
		port:   port,
	}
}

// Register registers a health check.
func (c *Checker) Register(name string, check CheckFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// Start starts the health check HTTP server.
func (c *Checker) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", c.handleHealth)
	mux.HandleFunc("/health/live", c.handleLiveness)
	mux.HandleFunc("/health/ready", c.handleReadiness)

	server := &http.Server{
		Addr:    ":" + c.port,
		Handler: mux,
	}

	return server.ListenAndServe()
}

// handleHealth handles the /health endpoint.
func (c *Checker) handleHealth(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	health := HealthCheck{
		Status:     StatusUp,
		Timestamp:  time.Now(),
		Components: make(map[string]Component),
	}

	for name, check := range c.checks {
		component := check()
		health.Components[name] = component
		if component.Status == StatusDown {
			health.Status = StatusDown
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if health.Status == StatusDown {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(health)
}

// handleLiveness handles the /health/live endpoint.
func (c *Checker) handleLiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

// handleReadiness handles the /health/ready endpoint.
func (c *Checker) handleReadiness(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ready := true
	for _, check := range c.checks {
		component := check()
		if component.Status == StatusDown {
			ready = false
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if ready {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready"})
	}
}
