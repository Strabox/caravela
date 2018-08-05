package storage

import dockerClient "github.com/docker/docker/client"

// IPFSBackend is a backend implemented on top of the highly distributed IPFS file system.
type IPFSBackend struct {
	BaseBackend
}

// newIPFSBackend creates a new IPFSBackend storage backend.
func newIPFSBackend() (Backend, error) {
	return &IPFSBackend{
		BaseBackend: BaseBackend{},
	}, nil
}

func (ipfs *IPFSBackend) Init(dockerClient *dockerClient.Client) {
	ipfs.docker = dockerClient
}

func (ipfs *IPFSBackend) LoadImage(imageKey string) (string, error) {
	// To be implemented if there is time.
	return "", nil
}
