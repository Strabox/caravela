package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"log"
)

/*
Client interfaces with docker daemon using the Docker SDK
*/
type Client struct {
	docker *client.Client
}

/*
Creates a new docker client
*/
func NewDockerClient() *Client {
	res := &Client{}
	res.docker = nil
	return res
}

/*
Initialize a Docker client.
*/
func (dockerClient *Client) Initialize(runningDockerVersion string) {
	var err error
	dockerClient.docker, err = client.NewClientWithOpts(client.WithVersion(runningDockerVersion))
	if err != nil {
		log.Println("[Docker] Initialize: ", err)
		panic(err)
	}

}

/*
Get CPU and RAM dedicated to Docker engine (Decided by the user in Docker configuration)
*/
func (dockerClient *Client) GetDockerCPUandRAM() (int, int) {
	if dockerClient.docker != nil {
		ctx := context.Background()
		info, _ := dockerClient.docker.Info(ctx)
		cpu := info.NCPU
		ram := info.MemTotal / 1000000 //Return in MB (MegaBytes)
		return cpu, int(ram)
	}
	panic(fmt.Errorf("[Docker] Docker client not initialized"))
}
