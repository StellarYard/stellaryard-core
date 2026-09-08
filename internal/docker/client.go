package docker

import (
	"github.com/StellarYard/stellaryard-core/internal/models"
)

// Client wraps the Docker SDK for container orchestration.
type Client struct {
	// TODO: add docker client field
}

// NewClient creates a new Docker client connected to the local Docker daemon.
func NewClient() (*Client, error) {
	// TODO: implement - connect to Docker daemon via SDK
	panic("not implemented")
}

// Start starts a named container ("horizon" or "soroban-rpc").
func (c *Client) Start(name string) error {
	// TODO: implement
	panic("not implemented")
}

// Stop stops a named container.
func (c *Client) Stop(name string) error {
	// TODO: implement
	panic("not implemented")
}

// ListStatus returns the status of all managed containers.
func (c *Client) ListStatus() ([]models.ContainerStatus, error) {
	// TODO: implement
	panic("not implemented")
}

// Logs returns a channel of log lines for the named container.
// The caller is responsible for closing the channel when done.
func (c *Client) Logs(name string) (<-chan string, error) {
	// TODO: implement - stream container logs
	panic("not implemented")
}

// HealthCheck verifies a container is responsive.
func (c *Client) HealthCheck(name string) (string, error) {
	// TODO: implement - check container health
	// NOTE: Horizon HTTP 200 does NOT mean synced.
	// Verify actual Stellar API responsiveness, not just port availability.
	panic("not implemented")
}
