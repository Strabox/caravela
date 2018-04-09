package docker

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"strings"
)

/*
Client interfaces with docker daemon using the Docker SDK
*/
type Client struct {
	docker *dockerClient.Client
}

/*
Creates a new docker remote to interact with the local Docker Engine.
*/
func NewDockerClient() *Client {
	res := &Client{}
	res.docker = nil
	return res
}

const ClientNotInitializedError = "[Docker] Please turn on the Docker Engine"

/*
Initialize a Docker remote with a corresponding docker daemon API version.
*/
func (client *Client) Initialize(runningDockerVersion string) {
	var err error
	client.docker, err = dockerClient.NewClientWithOpts(dockerClient.WithVersion(runningDockerVersion))
	if err != nil {
		log.Fatalf("[Docker] Initialize error: %s", err.Error())
	}
}

/*
Get CPU and RAM dedicated to Docker engine (Decided by the user in Docker configuration).
*/
func (client *Client) GetDockerCPUAndRAM() (int, int) {
	if client.docker != nil {
		ctx := context.Background()
		info, _ := client.docker.Info(ctx)
		cpu := info.NCPU
		ram := info.MemTotal / 1000000 //Return in MB (MegaBytes)
		return cpu, int(ram)
	} else {
		panic(fmt.Errorf(ClientNotInitializedError))
	}
}

/*
Launches a container from an image in the local Docker Engine.
*/
func (client *Client) RunContainer(imageKey string, args []string, machineCpus string, ram int) {
	if client.docker != nil {
		imageKeyTokens := strings.Split(imageKey, "/")
		ctx := context.Background()

		_, err := client.docker.ImagePull(ctx, imageKey, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		resp, err := client.docker.ContainerCreate(ctx, &container.Config{
			Image: imageKeyTokens[len(imageKeyTokens)-1], // Image key name
			Cmd:   args,                                  // Command arguments to the container
			Tty:   true,
		}, &container.HostConfig{
			Resources: container.Resources{
				Memory:     int64(ram) * 1000000,
				CpusetCpus: machineCpus,
			},
		}, nil, "")
		if err != nil {
			panic(err)
		}

		if err := client.docker.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}

		statusCh, errCh := client.docker.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				panic(err)
			}
		case <-statusCh:
			// Container is running!!!
		}
	} else {
		panic(fmt.Errorf(ClientNotInitializedError))
	}
}
