package client

import (
	"fmt"
	"github.com/strabox/caravela/api/rest"
	"net/http"
	"strconv"
	"strings"
)

/*
Client that can be used as (Golang SDK) to interact with the CARAVELA daemon.
It is used in the CARAVELA's CLI application.
*/
type Caravela struct {
	// HTTP client to send requests into Caravela's REST daemon
	httpClient *http.Client
	// Configuration parameters for the CARAVELA client
	config *Configuration
}

func NewCaravelaLocal() *Caravela {
	return NewCaravelaIP("127.0.0.1")
}

func NewCaravelaIP(caravelaHostIP string) *Caravela {
	res := &Caravela{}
	res.config = DefaultConfiguration(caravelaHostIP)
	res.httpClient = &http.Client{
		Timeout: res.config.HttpRequestTimeout(),
	}
	return res
}

func (client *Caravela) Run(containerImageKey string, portMappings []string, arguments []string,
	cpus int, ram int) *Error {

	portMappingsJSON := make([]rest.PortMappingJSON, 0)
	for _, portMap := range portMappings {
		portMapping := strings.Split(portMap, ":")
		resultPortMap := rest.PortMappingJSON{}
		resultPortMap.HostPort, _ = strconv.Atoi(portMapping[0])
		resultPortMap.ContainerPort, _ = strconv.Atoi(portMapping[1])
		portMappingsJSON = append(portMappingsJSON, resultPortMap)
	}

	runContainerJSON := rest.RunContainerJSON{ContainerImageKey: containerImageKey, Arguments: arguments,
		PortMappings: portMappingsJSON, CPUs: cpus, RAM: ram}

	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserContainerBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, runContainerJSON, nil)
	if err == nil {
		if httpCode == http.StatusOK {
			return nil
		} else {
			return NewClientError(fmt.Errorf("impossible deploy the container"))
		}
	} else {
		return NewClientError(err)
	}
}
