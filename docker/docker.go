package docker

import (
	"fmt"
	"github.com/docker/docker/client"
)

/*
InitializeDockerClient creates and initialize a docker client
*/
func InitializeDockerClient(runningDockerVersion string) *client.Client {
	cli, err := client.NewEnvClient()

	if err != nil {
		fmt.Println("[Creating Docker SDK Client] ", err)
		panic(err)
	}

	cli, err = client.NewClientWithOpts(client.WithVersion(runningDockerVersion))
	if err != nil {
		fmt.Println("[Init Docker SDK Client] ", err)
		panic(err)
	}

	return cli
}

/*
GetDockerCPUandRAM get CPU and RAM dedicated to Docker engine (BY the user)
*/
func GetDockerCPUandRAM(client *client.Client) (uint, uint) {
	ctx := context.Background()
	info, _ := client.Info(ctx)
	cpu := uint(info.NCPU)
	ram := uint(info.MemTotal / 1000000) //Return in MB (MegaBytes)
	return cpu, ram
}
