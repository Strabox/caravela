package storage

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/strabox/caravela/util"
	"io/ioutil"
)

type DockerHubBackend struct {
	BaseBackend
}

func NewDockerHubBackend(dockerCli *dockerClient.Client) *DockerHubBackend {
	return &DockerHubBackend{
		BaseBackend: BaseBackend{
			docker: dockerCli,
		},
	}
}

func (dockerHub *DockerHubBackend) Load(imageKey string) (string, error) {
	ctx := context.Background()

	out, err := dockerHub.DockerClient().ImagePull(ctx, imageKey, types.ImagePullOptions{})
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
