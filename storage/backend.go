package storage

import dockerClient "github.com/docker/docker/client"

// Backend interface for container's images storage backends.
type Backend interface {
	// Init initialize the storage backend with the docker client (SDK).
	Init(dockerClient *dockerClient.Client)

	// Load loads the image inside the docker engine and returns the image key assigned inside the docker engine.
	LoadImage(imageKey string) (string, error)
}

// BaseBackend is a base for all the storage backends.
type BaseBackend struct {
	docker *dockerClient.Client
}
