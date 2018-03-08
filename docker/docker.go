package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
)

var dockerClient *client.Client = nil

/*
Creates and initialize a docker client.
Used to access the Docker engine.
*/
func Initialize(runningDockerVersion string) {
	if dockerClient == nil {
		cli, err := client.NewEnvClient()

		if err != nil {
			fmt.Println("[Docker] Init: ", err)
			panic(err)
		}

		cli, err = client.NewClientWithOpts(client.WithVersion(runningDockerVersion))
		if err != nil {
			fmt.Println("[Docker] Init: ", err)
			panic(err)
		}

		dockerClient = cli
	}
}

/*
Get CPU and RAM dedicated to Docker engine (Decided by the user in Docker configuration)
*/
func GetDockerCPUandRAM() (int, int) {
	if dockerClient != nil {
		ctx := context.Background()
		info, _ := dockerClient.Info(ctx)
		cpu := info.NCPU
		ram := info.MemTotal / 1000000 //Return in MB (MegaBytes)
		return cpu, int(ram)
	}
	panic(fmt.Errorf("[Docker] Client Not Initialized"))
}
