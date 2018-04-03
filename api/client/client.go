package client

import (
	"bytes"
	"encoding/json"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	"net/http"
	"time"
)

// Our HTTP body is always a JSON
const HTTPContentType = "application/json"
const HTTPRequestTimeout = 2 * time.Second // TODO: Put in configuration struct?

type Caravela struct {
	httpClient *http.Client
}

func NewCaravela() *Caravela {
	res := &Caravela{}

	res.httpClient = &http.Client{
		Timeout: HTTPRequestTimeout,
	}

	return res
}

func (client *Caravela) Run(containerImage string, arguments []string, cpus int, ram int) {
	var runContainer rest.RunContainerJSON

	runContainer.ContainerImage = containerImage
	runContainer.Arguments = arguments
	runContainer.CPUs = cpus
	runContainer.RAM = ram

	url := rest.BuildHttpURL(false, "localhost", configuration.APIPort, rest.UserBaseEndpoint+
		rest.UserRunContainerEndpoint)

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(runContainer)

	_, err := client.httpClient.Post(url, HTTPContentType, buffer)
	if err == nil {
		//return nil
	} else {
		//return NewClientError(UNKNOWN)
	}

}
