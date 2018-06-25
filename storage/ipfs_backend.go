package storage

import dockerClient "github.com/docker/docker/client"

type IPFSBackend struct {
	BaseBackend
}

func NewIPFSBackend(dockerCli *dockerClient.Client) *IPFSBackend {
	return &IPFSBackend{
		BaseBackend: BaseBackend{
			docker: dockerCli,
		},
	}
}

func (ipfs *IPFSBackend) Load(imageKey string) (string, error) {
	// TODO: Implement the IPFS backend
	return "", nil
}
