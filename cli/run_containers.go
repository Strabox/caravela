package cli

import (
	"context"
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/util"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func runContainers(c *cli.Context) {
	if c.NArg() < 1 {
		fatalPrintln("Please provide a container image to launch or a valid request .yml file path")
	}

	var containersConfigs []types.ContainerConfig = nil

	if strings.HasSuffix(c.Args().First(), ".yml") { // Deploy request using a .yml file.
		fileContent, err := ioutil.ReadFile(c.Args().First())
		if err != nil {
			fatalPrintf("Impossible read request file %s. %s\n", c.Args().First(), err)
		}

		var services map[string]containerRequest
		if err := yaml.Unmarshal(fileContent, &services); err != nil {
			fatalPrintf("Problem parsing request file %s. %s\n", c.Args().First(), err)
		}

		containersConfigs = make([]types.ContainerConfig, len(services))
		i := 0
		for serviceName, service := range services {
			portMappings, err := validatePortMappings(service.PortMappings)
			if err != nil {
				fatalPrintf("Service %s. %s\n", serviceName, err)
			}

			var cpuClass types.CPUClass
			err = cpuClass.ValueOf(service.CPUPower)
			if err != nil {
				fatalPrintf("Service %s. %s\n", serviceName, err)
			}

			var groupPolicy types.GroupPolicy
			err = groupPolicy.ValueOf(service.GroupPolicy)
			if err != nil {
				fatalPrintf("Service %s. %s\n", serviceName, err)
			}

			containersConfigs[i] = types.ContainerConfig{
				Name:         serviceName,
				ImageKey:     service.ImageKey,
				Args:         service.Args,
				PortMappings: portMappings,
				Resources: types.Resources{
					CPUClass: cpuClass,
					CPUs:     service.CPUs,
					RAM:      service.RAM,
				},
				GroupPolicy: groupPolicy,
			}
			i++
		}
	} else { // Deploy request using the command line arguments and flags.
		portMappings, err := validatePortMappings(c.StringSlice("p"))
		if err != nil {
			fatalPrintln(err)
		}

		// Obtains all the arguments provided to the container launch, from the command line.
		var containerArgs []string = nil
		if c.NArg() > 1 {
			containerArgs = make([]string, c.NArg()-1)
			for i := 1; i < c.NArg(); i++ {
				containerArgs[i-1] = c.Args().Get(i)
			}
		}

		var cpuClass types.CPUClass
		err = cpuClass.ValueOf(c.String("cp"))
		if err != nil {
			fatalPrintln(err)
		}

		containersConfigs = make([]types.ContainerConfig, 1)
		containersConfigs[0] = types.ContainerConfig{
			Name:         c.String("name"),
			ImageKey:     c.Args().First(),
			Args:         containerArgs,
			PortMappings: portMappings,
			Resources: types.Resources{
				CPUClass: cpuClass,
				CPUs:     int(c.Uint("cpus")),
				RAM:      int(c.Uint("ram")),
			},
		}
	}

	// Create a user client of the CARAVELA system
	caravelaClient := client.NewCaravelaTimeoutIP(c.GlobalString("ip"), 3600*time.Second) // TODO: Timeout hack to handler the submit container request
	err := caravelaClient.SubmitContainers(context.Background(), containersConfigs)
	if err != nil {
		fatalPrintln(err)
	}
}

// validatePortMappings validates a list of port mappings given by the user.
func validatePortMappings(inputPortMappings []string) ([]types.PortMapping, error) {
	resPortMappings := make([]types.PortMapping, 0)
	for _, portMapStr := range inputPortMappings {
		var err error = nil
		resultPortMap := types.PortMapping{}

		regex := regexp.MustCompile("(^[0-9]+$)|(^[0-9]+/(tcp|udp)$)|(^[0-9]+:[0-9]+$)|(^[0-9]+:[0-9]+/(tcp|udp)$)")
		if !regex.MatchString(portMapStr) {
			return nil, fmt.Errorf("invalid syntax for port mapping expression")
		}

		portMapping := strings.Split(portMapStr, ":")

		if len(portMapping) == 1 { // Publish the container port into a random one
			containerPort := strings.Split(portMapping[0], "/")[0]
			if resultPortMap.ContainerPort, err = strconv.Atoi(containerPort); err != nil {
				return nil, fmt.Errorf("Port should be a positive integer. Error: %s.\n", containerPort)
			}
			if !util.IsValidPort(resultPortMap.ContainerPort) {
				return nil, fmt.Errorf("Invalid container port number in: %s\n", portMapStr)
			}

			resultPortMap.HostPort = 0
		} else if len(portMapping) == 2 { // Publish the container port into a user defined port
			hostPort := portMapping[0]
			containerPort := strings.Split(portMapping[1], "/")[0]
			if resultPortMap.HostPort, err = strconv.Atoi(hostPort); err != nil {
				return nil, fmt.Errorf("Port should be a positive integer. Error: %s.\n", hostPort)
			}
			if !util.IsValidPort(resultPortMap.HostPort) {
				return nil, fmt.Errorf("Invalid host port number in: %s\n", portMapStr)
			}

			if resultPortMap.ContainerPort, err = strconv.Atoi(containerPort); err != nil {
				return nil, fmt.Errorf("Port should be a positive integer. Error: %s.\n", containerPort)
			}
			if !util.IsValidPort(resultPortMap.ContainerPort) {
				return nil, fmt.Errorf("Invalid container port number in: %s\n", portMapStr)
			}
		}

		if match, _ := regexp.MatchString("tcp", portMapStr); match {
			resultPortMap.Protocol = "tcp"
		} else if match, _ := regexp.MatchString("udp", portMapStr); match {
			resultPortMap.Protocol = "udp"
		} else {
			resultPortMap.Protocol = "tcp"
		}

		resPortMappings = append(resPortMappings, resultPortMap)
	}
	return resPortMappings, nil
}

// containerRequest holds the YAML file content for a container deployment request.
type containerRequest struct {
	Name         string   `yaml:"name"`
	ImageKey     string   `yaml:"image"`
	Args         []string `yaml:"args"`
	PortMappings []string `yaml:"ports"`
	CPUPower     string   `yaml:"cpu_power"`
	CPUs         int      `yaml:"cpus"`
	RAM          int      `yaml:"ram"`
	GroupPolicy  string   `yaml:"group_policy"`
}

func (s *containerRequest) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawContainerRequest containerRequest
	defaultValues := rawContainerRequest{
		Args:         defaultContainerArgs,
		PortMappings: defaultPortMappingsArgs,
		CPUPower:     defaultCPUPower,
		CPUs:         defaultCPUs,
		RAM:          defaultRAM,
		GroupPolicy:  defaultContainerGroupPolicy,
	} // Default values for a container configuration
	if err := unmarshal(&defaultValues); err != nil {
		return err
	}

	*s = containerRequest(defaultValues)
	return nil
}
