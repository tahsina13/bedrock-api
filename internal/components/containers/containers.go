package containers

import "time"

// ContainerConfig holds the parameters needed to create a container.
type ContainerConfig struct {
	// Container name.
	Name string
	// Container image and tag, e.g. "ubuntu:latest".
	Image string
	// Environment variables in "KEY=VALUE" format.
	Env []string
	// Command to run in the container. If empty, the image's default CMD is used.
	Cmd []string
	// Volumes to mount, mapping host paths to container paths.
	Volumes map[string]string
	// Flags to control container behavior (e.g. privileged, network mode).
	Flags map[string]any
}

// ContainerInfo describes a container's current state.
type ContainerInfo struct {
	// Unique identifier of the container.
	ID string
	// Human-readable name of the container.
	Name string
	// Image the container was created from, e.g. "ubuntu:latest".
	Image string
	// Command the container was started with, e.g. "/bin/sh -c 'echo hello'".
	Command string
	// Current status of the container, e.g. "running", "exited".
	Status string
	// Exited indicates whether the container has finished execution.
	Exited bool
	// Exit code if the container has finished.
	ExitCode int
	// Creation timestamp of the container.
	CreatedAt time.Time
}
