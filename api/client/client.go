package client

import (
	"fmt"
	"github.com/strabox/caravela/api/rest"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
Client that can be used as (Golang SDK) to interact with the CARAVELA daemon.
It is used in the CARAVELA's CLI application.
*/
type Caravela struct {
	// HTTP client to send requests into Caravela's REST daemon
	httpClient *http.Client
	// Configuration parameters for the CARAVELA's client
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

func NewCaravelaTimeoutIP(caravelaHostIP string, requestTimeout time.Duration) *Caravela {
	res := &Caravela{}
	res.config = DefaultConfiguration(caravelaHostIP)
	res.httpClient = &http.Client{
		Timeout: requestTimeout,
	}
	return res
}

func (client *Caravela) RunContainer(containerImageKey string, portMappings []string, arguments []string,
	cpus int, ram int) *Error {

	portMappingsList := make([]rest.PortMapping, 0)
	for _, portMap := range portMappings {
		portMapping := strings.Split(portMap, ":")
		resultPortMap := rest.PortMapping{}
		resultPortMap.HostPort, _ = strconv.Atoi(portMapping[0])
		resultPortMap.ContainerPort, _ = strconv.Atoi(portMapping[1])
		portMappingsList = append(portMappingsList, resultPortMap)
	}

	runContainerMessage := rest.RunContainerMessage{
		ContainerImageKey: containerImageKey,
		Arguments:         arguments,
		PortMappings:      portMappingsList,
		CPUs:              cpus,
		RAM:               ram,
	}

	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserContainerBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, runContainerMessage, nil)
	if err != nil {
		return NewClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return NewClientError(fmt.Errorf("impossible deploy the container"))
	}
}

func (client *Caravela) StopContainers(containersIDs []string) *Error {
	stopContainersMessage := rest.StopContainersMessage{ContainersIDs: containersIDs}

	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserContainerBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodDelete, stopContainersMessage, nil)
	if err != nil {
		return NewClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return NewClientError(fmt.Errorf("error stopping the containers"))
	}
}

func (client *Caravela) ListContainers() (*rest.ContainersList, *Error) {
	var containersList rest.ContainersList

	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserContainerBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, nil, &containersList)
	if err != nil {
		return nil, NewClientError(err)
	}

	if httpCode == http.StatusOK {
		return &containersList, nil
	} else {
		return nil, NewClientError(fmt.Errorf("error checking the container"))
	}
}

func (client *Caravela) Exit() *Error {
	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserExitEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, nil, nil)
	if err != nil {
		return NewClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return NewClientError(fmt.Errorf("error exiting from the system"))
	}
}
