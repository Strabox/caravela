package docker

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	myContainer "github.com/strabox/caravela/docker/container"
	"github.com/strabox/caravela/storage"
	"github.com/strabox/caravela/util"
	"strconv"
)

// DefaultClient that interfaces with docker SDK.
type DefaultClient struct {
	docker        *dockerClient.Client
	imagesBackend storage.Backend
}

// Creates a new docker client to interact with the local Docker Engine.
func NewDockerClient(config *configuration.Configuration) *DefaultClient {
	var err error
	res := &DefaultClient{}

	res.docker, err = dockerClient.NewClientWithOpts(dockerClient.WithVersion(config.DockerAPIVersion()))
	if err != nil {
		log.Fatalf(util.LogTag("Docker")+"Initialize error: %s", err)
	}

	switch imagesBackend := config.ImagesStorageBackend(); imagesBackend {
	case configuration.ImagesStorageDockerHub:
		res.imagesBackend = storage.NewDockerHubBackend(res.docker)
	case configuration.ImagesStorageIPFS:
		res.imagesBackend = storage.NewIPFSBackend(res.docker)
	}
	return res
}

// Verify if the Docker client is initialized or not.
func (client *DefaultClient) isInit() {
	if client.docker != nil {
		if _, err := client.docker.Ping(context.Background()); err != nil {
			// TODO: Shutdown node gracefully in each place where docker calls can fail!!
			log.Fatalf(util.LogTag("Docker") + "Please turn on the Docker Engine")
		}
	} else {
		log.Fatalf(util.LogTag("Docker") + "Please initialize the Docker client")
	}
}

// Get CPUs and RAM dedicated to Docker engine (Decided by the user in Docker configuration).
func (client *DefaultClient) GetDockerCPUAndRAM() (int, int) {
	client.isInit()

	ctx := context.Background()
	info, err := client.docker.Info(ctx)
	if err != nil {
		log.Errorf(util.LogTag("Docker")+"Get Docker Info error: %s", err)
	}

	cpu := info.NCPU
	ram := info.MemTotal / 1000000 // Return in MB (MegaBytes)
	return cpu, int(ram)
}

// Check the container status (running, stopped, etc)
func (client *DefaultClient) CheckContainerStatus(containerID string) (myContainer.ContainerStatus, error) {
	client.isInit()

	ctx := context.Background()
	status, err := client.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return myContainer.NewContainerStatus(myContainer.Unknown), err
	}

	if status.State.Running {
		return myContainer.NewContainerStatus(myContainer.Running), nil
	} else {
		return myContainer.NewContainerStatus(myContainer.Finished), nil
	}
}

// Launches a container from an image in the local Docker Engine.
func (client *DefaultClient) RunContainer(imageKey string, portMappings []rest.PortMapping,
	args []string, cpus int64, ram int) (string, error) {

	client.isInit()

	dockerImageKey, err := client.imagesBackend.Load(imageKey)
	if err != nil {
		log.Errorf(util.LogTag("Docker")+"Loading image error", err)
		return "", err
	}

	ctx := context.Background()

	// Port mappings creation
	containerPortSet := nat.PortSet{}
	hostPortMap := nat.PortMap{}
	for _, portMap := range portMappings {
		containerPort := strconv.Itoa(portMap.ContainerPort)
		hostPort := strconv.Itoa(portMap.HostPort)
		port, _ := nat.NewPort(fmt.Sprintf("tcp"), containerPort)
		hostPortMap[port] = []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: hostPort,
		}}
		containerPortSet[port] = struct{}{}
	}

	resp, err := client.docker.ContainerCreate(ctx, &container.Config{
		Image:        dockerImageKey, // Image key name
		Cmd:          args,           // Command arguments to the container
		Tty:          true,
		ExposedPorts: containerPortSet, // Container's exposed ports
	}, &container.HostConfig{
		Resources: container.Resources{
			Memory:   int64(ram) * 1000000, // Maximum memory available to container
			CPUCount: cpus,                 // Number of CPUs available to the container
		},
		PortBindings: hostPortMap, // Port mappings between container' port and host ports
	}, nil, "")
	if err != nil { // Error creating the container
		client.docker.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}) // Remove the container (avoid filling space)
		log.Errorf(util.LogTag("Docker")+"Creating container error: %s", err)
		return "", err
	}

	if err := client.docker.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		client.docker.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}) // Remove the container (avoid filling space)
		log.Errorf(util.LogTag("Docker")+"Starting container error: %s", err)
		return "", err // Error starting the container
	}

	log.Infof(util.LogTag("Docker")+"Container RUNNING, Image: %s, Args: %v, Resources: <%d,%d>",
		imageKey, args, cpus, ram)

	return resp.ID, nil
}

// Remove a container from the Docker engine (to avoid filling space in the node).
func (client *DefaultClient) RemoveContainer(containerID string) error {
	client.isInit()

	err := client.docker.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})
	return fmt.Errorf("problem stopping and removing container. Error: %s", err)
}
