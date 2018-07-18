package client

import (
	"fmt"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/api/types"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HTTPClient can be used as a Golang SDK to interact with a CARAVELA daemon.
// It is used in the CARAVELA's CLI package (github.com/strabox/caravela/cli).
type HttpClient struct {
	// HTTP client to send requests into CARAVELA's REST daemon
	httpClient *http.Client
	// Configuration parameters for the CARAVELA's client
	config *Configuration
}

func NewCaravelaIP(caravelaHostIP string) *HttpClient {
	config := DefaultConfiguration(caravelaHostIP)

	return &HttpClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HttpRequestTimeout(),
		},
	}

}

func NewCaravelaTimeoutIP(caravelaHostIP string, requestTimeout time.Duration) *HttpClient {
	return &HttpClient{
		config: DefaultConfiguration(caravelaHostIP),
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (client *HttpClient) SubmitContainers(containerImageKey string, portMappings []string, arguments []string,
	cpus int, ram int) *Error {

	portMappingsList := make([]types.PortMapping, 0)
	for _, portMap := range portMappings {
		portMapping := strings.Split(portMap, ":")
		resultPortMap := types.PortMapping{}
		resultPortMap.HostPort, _ = strconv.Atoi(portMapping[0])
		resultPortMap.ContainerPort, _ = strconv.Atoi(portMapping[1])
		portMappingsList = append(portMappingsList, resultPortMap)
	}

	runContainerMessage := rest.RunContainerMsg{
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

func (client *HttpClient) StopContainers(containersIDs []string) *Error {
	stopContainersMessage := rest.StopContainersMsg{ContainersIDs: containersIDs}

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

func (client *HttpClient) ListContainers() ([]types.ContainerStatus, *Error) {
	var containersList rest.ContainersStatusMsg

	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserContainerBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, nil, &containersList)
	if err != nil {
		return nil, NewClientError(err)
	}

	if httpCode == http.StatusOK {
		return containersList.ContainersStatus, nil
	} else {
		return nil, NewClientError(fmt.Errorf("error checking the container"))
	}
}

func (client *HttpClient) Exit() *Error {
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
