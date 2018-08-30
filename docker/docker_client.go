package docker

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	caravelaTypes "github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	myContainer "github.com/strabox/caravela/docker/container"
	"github.com/strabox/caravela/docker/events"
	"github.com/strabox/caravela/storage"
	"github.com/strabox/caravela/util"
	"strconv"
)

// Client interfaces with docker Golang's SDK/Client.
type Client struct {
	docker        *dockerClient.Client
	imagesBackend storage.Backend
	configs       *configuration.Configuration
}

// NewDockerClient creates a new docker client to interact with the local Docker Engine.
func NewDockerClient(config *configuration.Configuration) *Client {
	dockClient, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(config.DockerAPIVersion()))
	if err != nil {
		log.Fatalf(util.LogTag("DOCKER")+"Init error: %s", err)
	}

	imagesBackend := storage.CreateBackend(config)
	imagesBackend.Init(dockClient)

	return &Client{
		docker:        dockClient,
		imagesBackend: imagesBackend,
		configs:       config,
	}
}

func (c *Client) Start() <-chan *events.Event {
	caravelaEventChan := make(chan *events.Event, 15)
	go func() {
		eventsToListen := filters.NewArgs(filters.Arg("event", "die"))
		eventChan, errChan := c.docker.Events(
			context.Background(),
			types.EventsOptions{
				Filters: eventsToListen,
			})
		for {
			select {
			case newDockerEvent := <-eventChan:
				caravelaEventChan <- &events.Event{Type: events.ContainerDied, Value: newDockerEvent.ID}
			case newDockerErrEvent := <-errChan:
				log.Fatalf(util.LogTag("DOCKER")+"Error receiving events, error: %s", newDockerErrEvent)
				eventChan, errChan = c.docker.Events(
					context.Background(),
					types.EventsOptions{
						Filters: eventsToListen,
					})
			}
		}
	}()
	return caravelaEventChan
}

// isInit verify if the Docker client is initialized or not.
func (c *Client) isInit() {
	if c.docker != nil {
		if _, err := c.docker.Ping(context.Background()); err != nil {
			// TODO: Shutdown node gracefully in each place where docker calls can fail!!
			log.Fatalf(util.LogTag("DOCKER") + "Please turn on the Docker Engine")
		}
	} else {
		log.Fatalf(util.LogTag("DOCKER") + "Please initialize the Docker client")
	}
}

// Get CPUs and RAM dedicated to Docker engine (Decided by the user in Docker configuration).
func (c *Client) GetDockerCPUAndRAM() (int, int) {
	c.isInit()

	ctx := context.Background()
	info, err := c.docker.Info(ctx)
	if err != nil {
		log.Errorf(util.LogTag("DOCKER")+"Get Docker Info error: %s", err)
	}

	cpu := info.NCPU
	ram := info.MemTotal / 1000000 // Return in MB (MegaBytes)
	return cpu, int(ram)
}

// CheckContainerStatus checks the container status (running, stopped, etc)
func (c *Client) CheckContainerStatus(containerID string) (myContainer.Status, error) {
	c.isInit()

	ctx := context.Background()
	status, err := c.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return myContainer.NewContainerStatus(myContainer.Unknown), err
	}

	if status.State.Running {
		return myContainer.NewContainerStatus(myContainer.Running), nil
	} else {
		return myContainer.NewContainerStatus(myContainer.Finished), nil
	}
}

// RunContainer launches a container from an image in the local Docker Engine.
func (c *Client) RunContainer(contConfig caravelaTypes.ContainerConfig) (*caravelaTypes.ContainerStatus, error) {
	c.isInit()

	dockerImageKey, err := c.imagesBackend.LoadImage(contConfig.ImageKey)
	if err != nil {
		log.Errorf(util.LogTag("DOCKER")+"Loading image error", err)
		return nil, err
	}

	// Port mappings creation
	containerPortSet := nat.PortSet{}
	hostPortMap := nat.PortMap{}
	for _, portMap := range contConfig.PortMappings {
		containerPort := strconv.Itoa(portMap.ContainerPort)
		hostPort := strconv.Itoa(portMap.HostPort)
		port, _ := nat.NewPort(portMap.Protocol, containerPort)
		hostPortMap[port] = []nat.PortBinding{{
			HostIP:   "0.0.0.0", // Publish to the local host!
			HostPort: hostPort,
		}}
		containerPortSet[port] = struct{}{}
	}

	resp, err := c.docker.ContainerCreate(context.Background(),
		&container.Config{
			Image:        dockerImageKey,  // Image key name
			Cmd:          contConfig.Args, // Command arguments to the container
			Tty:          true,
			ExposedPorts: containerPortSet, // Container's exposed ports
		}, &container.HostConfig{
			Resources: container.Resources{
				CPUPeriod: 100000,
				CPUQuota:  int64(float64(100000) * (float64(contConfig.Resources.CPUs) / float64(c.configs.CPUSlices()))), // CPU maximum quota for the container using the Linux CFS scheduler.
				Memory:    int64(contConfig.Resources.RAM) * 1000000,                                                      // Maximum memory available to the container.
			},
			PortBindings: hostPortMap, // Port mappings between container's port and host's port
		}, nil, contConfig.Name)
	if err != nil { // Error creating the container
		c.docker.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{}) // Remove the container (avoid filling space)
		log.Errorf(util.LogTag("DOCKER")+"Creating container error: %s", err)
		return nil, err
	}

	if err := c.docker.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		c.docker.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{}) // Remove the container (avoid filling space)
		log.Errorf(util.LogTag("DOCKER")+"Starting container error: %s", err)
		return nil, err // Error starting the container
	}

	log.Infof(util.LogTag("DOCKER")+"Container RUNNING, Img: %s, Args: %v, Res: <%d,%d>",
		contConfig.ImageKey, contConfig.Args, contConfig.Resources.CPUs, contConfig.Resources.RAM)

	// Update the container's information with Docker's engine generated information, e.g. random name, random port etc
	contDockerInfo, _ := c.docker.ContainerInspect(context.Background(), resp.ID)
	contConfig.Name = contDockerInfo.Name[1:]
	for ui, userPortMap := range contConfig.PortMappings {
		portKey, _ := nat.NewPort(userPortMap.Protocol, strconv.Itoa(userPortMap.ContainerPort))
		dockerBindings := contDockerInfo.NetworkSettings.Ports[portKey]
		for _, binding := range dockerBindings {
			contConfig.PortMappings[ui].HostPort, _ = strconv.Atoi(binding.HostPort)
		}
	}

	return &caravelaTypes.ContainerStatus{
		ContainerConfig: contConfig,
		ContainerID:     resp.ID,
		Status:          "Running",
	}, nil
}

// RemoveContainer removes a container from the Docker engine (to avoid filling space in the node).
func (c *Client) RemoveContainer(containerID string) error {
	c.isInit()

	err := c.docker.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("problem stopping/removing container error: %s", err)
	}
	return nil
}
