package docker

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/storage"
	"github.com/strabox/caravela/util"
)

/*
DefaultClient that interfaces with docker SDK.
*/
type DefaultClient struct {
	docker        *dockerClient.Client
	imagesBackend storage.Backend
}

/*
Creates a new docker client to interact with the local Docker Engine.
*/
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

/*
Verify if the Docker client is initialized or not.
*/
func (client *DefaultClient) verifyInitialization() {
	if client.docker != nil {
		if _, err := client.docker.Ping(context.Background()); err != nil {
			// TODO: Shutdown node gracefully in each place where docker calls can fail!!
			log.Fatalf(util.LogTag("Docker") + "Please turn on the Docker Engine")
		}
	} else {
		log.Fatalf(util.LogTag("Docker") + "Please initialize the Docker client")
	}
}

/*
Get CPUs and RAM dedicated to Docker engine (Decided by the user in Docker configuration).
*/
func (client *DefaultClient) GetDockerCPUAndRAM() (int, int) {
	client.verifyInitialization()

	ctx := context.Background()
	info, err := client.docker.Info(ctx)
	if err != nil {
		log.Errorf(util.LogTag("Docker")+"Get Info: %s", err)
	}
	cpu := info.NCPU
	ram := info.MemTotal / 1000000 // Return in MB (MegaBytes)
	return cpu, int(ram)
}

/*
Check the container status (running, stopped, etc)
*/
func (client *DefaultClient) CheckContainerStatus(containerID string) (ContainerStatus, error) {
	client.verifyInitialization()

	ctx := context.Background()
	status, err := client.docker.ContainerInspect(ctx, containerID)
	if err == nil {
		if status.State.Running {
			return NewContainerStatus(Running), nil
		} else {
			return NewContainerStatus(Finished), nil
		}
	} else {
		return NewContainerStatus(Unknown), err
	}
}

/*
Launches a container from an image in the local Docker Engine.
*/
func (client *DefaultClient) RunContainer(imageKey string, args []string, machineCpus string, ram int) (string, error) {
	client.verifyInitialization()

	dockerImageKey, err := client.imagesBackend.Load(imageKey)
	if err != nil {
		log.Errorf(util.LogTag("Docker")+"Loading image error", err)
		return "", err
	}

	ctx := context.Background()

	resp, err := client.docker.ContainerCreate(ctx, &container.Config{
		Image: dockerImageKey, // Image key name
		Cmd:   args,           // Command arguments to the container
		Tty:   true,
	}, &container.HostConfig{
		Resources: container.Resources{
			Memory:     int64(ram) * 1000000, // Maximum memory available to container
			CpusetCpus: machineCpus,          // Number of CPUs available to the container
		},
	}, nil, "")
	if err != nil { // Error creating the container from the given
		log.Errorf(util.LogTag("Docker")+"Creating container error: %s", err)
		return "", err
	}

	if err := client.docker.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Errorf(util.LogTag("Docker")+"Starting container error: %s", err)
		return "", err // Error starting the container
	}

	statusCh, errCh := client.docker.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Errorf(util.LogTag("Docker")+"Waiting for container to exit error: %s", err)
			return "", err
		}
	case <-statusCh:
		// Container is finally running!!!
		log.Infof(util.LogTag("Docker")+"Container running, Image: %s, Args: %v, Resources: <%d,%d>",
			imageKey, args, machineCpus, ram)
	}
	return resp.ID, nil
}

/*
Remove a container from the Docker image (to avoid filling space in the node).
*/
func (client *DefaultClient) RemoveContainer(containerID string) {
	client.verifyInitialization()

	ctx := context.Background()
	client.docker.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
}
