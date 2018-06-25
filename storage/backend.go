package storage

import dockerClient "github.com/docker/docker/client"

/*
Interface for container's images storage Backends
*/
type Backend interface {
	// Returns the image key assigned inside the docker engine.
	Load(imageKey string) (string, error)
}

/*
BaseBackend structure
*/
type BaseBackend struct {
	docker *dockerClient.Client
}

func (bb *BaseBackend) DockerClient() *dockerClient.Client {
	return bb.docker
}
