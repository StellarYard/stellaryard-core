package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// Managed container names.
const (
	ContainerHorizon    = "horizon"
	ContainerSorobanRPC = "soroban-rpc"
)

// validContainers lists the containers managed by core.
var validContainers = map[string]bool{
	ContainerHorizon:    true,
	ContainerSorobanRPC: true,
}

// ContainerStatus represents the status of a Docker container.
type ContainerStatus struct {
	Name      string    `json:"name"`
	State     string    `json:"state"`
	Health    string    `json:"health"`
	StartedAt time.Time `json:"started_at"`
}

// Client wraps the Docker SDK client.
type Client struct {
	docker *client.Client
}

// NewClient creates a new Docker client connected to the local daemon.
func NewClient() (*Client, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &Client{docker: c}, nil
}

// Close closes the Docker client connection.
func (c *Client) Close() error {
	return c.docker.Close()
}

// Start starts a named container.
func (c *Client) Start(ctx context.Context, name string) error {
	if !validContainers[name] {
		return fmt.Errorf("unknown container: %s", name)
	}

	// Find the container by name
	containerID, err := c.findContainer(ctx, name)
	if err != nil {
		return err
	}

	return c.docker.ContainerStart(ctx, containerID, container.StartOptions{})
}

// Stop stops a named container.
func (c *Client) Stop(ctx context.Context, name string) error {
	if !validContainers[name] {
		return fmt.Errorf("unknown container: %s", name)
	}

	containerID, err := c.findContainer(ctx, name)
	if err != nil {
		return err
	}

	timeout := 5 // seconds
	return c.docker.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

// Status returns the status of a named container.
func (c *Client) Status(ctx context.Context, name string) (*ContainerStatus, error) {
	if !validContainers[name] {
		return nil, fmt.Errorf("unknown container: %s", name)
	}

	containerID, err := c.findContainer(ctx, name)
	if err != nil {
		return nil, err
	}

	inspect, err := c.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", name, err)
	}

	status := &ContainerStatus{
		Name: name,
	}

	switch inspect.State.Status {
	case "running":
		status.State = "running"
	case "exited", "dead":
		status.State = "stopped"
	default:
		status.State = inspect.State.Status
	}

	if inspect.State.Health != nil {
		status.Health = inspect.State.Health.Status
	} else {
		status.Health = "no-healthcheck"
	}

	if inspect.State.StartedAt != "" {
		t, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
		if err == nil {
			status.StartedAt = t
		}
	}

	return status, nil
}

// ListStatus returns the status of all managed containers.
func (c *Client) ListStatus(ctx context.Context) ([]ContainerStatus, error) {
	var statuses []ContainerStatus
	for name := range validContainers {
		status, err := c.Status(ctx, name)
		if err != nil {
			// Container might not exist yet
			statuses = append(statuses, ContainerStatus{
				Name:  name,
				State: "not-found",
			})
			continue
		}
		statuses = append(statuses, *status)
	}
	return statuses, nil
}

// Logs returns a reader for container logs.
func (c *Client) Logs(ctx context.Context, name string) (io.ReadCloser, error) {
	if !validContainers[name] {
		return nil, fmt.Errorf("unknown container: %s", name)
	}

	containerID, err := c.findContainer(ctx, name)
	if err != nil {
		return nil, err
	}

	return c.docker.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "100",
	})
}

// SubscribeLogs subscribes to container log events.
func (c *Client) SubscribeLogs(ctx context.Context, name string) (<-chan events.Message, error) {
	if !validContainers[name] {
		return nil, fmt.Errorf("unknown container: %s", name)
	}

	containerID, err := c.findContainer(ctx, name)
	if err != nil {
		return nil, err
	}

	msgCh, _ := c.docker.Events(ctx, events.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("container", containerID),
		),
	})

	return msgCh, nil
}

// findContainer finds a container ID by name.
func (c *Client) findContainer(ctx context.Context, name string) (string, error) {
	containers, err := c.docker.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	for _, cont := range containers {
		for _, n := range cont.Names {
			// Docker prefixes names with /
			if n == "/"+name || n == name {
				return cont.ID, nil
			}
		}
	}

	return "", fmt.Errorf("container not found: %s", name)
}
