package storage

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/strabox/caravela/util"
	"io/ioutil"
)

// DockerHubBackend represents a image storage backup that uses the Docker public registry to store
// and retrieve images to run in CARAVELA.
type DockerHubBackend struct {
	BaseBackend
}

// newDockerHubBackend creates a new DockerHubBackend storage backend.
func newDockerHubBackend() (Backend, error) {
	return &DockerHubBackend{
		BaseBackend: BaseBackend{},
	}, nil
}

func (dockerHub *DockerHubBackend) Init(dockerClient *dockerClient.Client) {
	dockerHub.docker = dockerClient
}

func (dockerHub *DockerHubBackend) LoadImage(imageKey string) (string, error) {
	out, err := dockerHub.docker.ImagePull(context.Background(), imageKey, types.ImagePullOptions{})
	if err != nil { // Error pulling the image from Docker
		log.Errorf(util.LogTag("DockerHub")+"Pulling container image error: %s", err)
		return imageKey, err
	}
	defer out.Close()

	if _, err := ioutil.ReadAll(out); err != nil {
		log.Errorf(util.LogTag("DockerHub")+"Reading container image error: %s", err)
		return imageKey, err
	}

	return imageKey, nil
}
