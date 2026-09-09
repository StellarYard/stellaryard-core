package models

import "time"

// ContainerStatus represents the state of a managed Docker container.
type ContainerStatus struct {
	Name    string    `json:"name"`   // "horizon" | "soroban-rpc"
	State   string    `json:"state"`  // "running" | "stopped" | "error"
	Health  string    `json:"health"` // "healthy" | "unhealthy" | "starting"
	Started time.Time `json:"started"`
}
