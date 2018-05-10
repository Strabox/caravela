package client

import (
	"github.com/strabox/caravela/api/rest"
	"net/http"
)

/*
Client that can be used as (Golang SDK) to interact with the Caravela Daemon.
It is used in the CLI client too.
*/
type Caravela struct {
	httpClient *http.Client   // Http client to send requests into Caravela's REST daemon
	config     *Configuration // Configuration parameters for the CARAVELA client
}

func NewCaravelaLocal() *Caravela {
	return NewCaravelaIP("localhost")
}

func NewCaravelaIP(caravelaHostIP string) *Caravela {
	res := &Caravela{}
	res.config = DefaultConfiguration(caravelaHostIP)
	res.httpClient = &http.Client{
		Timeout: res.config.HttpRequestTimeout(),
	}
	return res
}

func (client *Caravela) Run(containerImage string, arguments []string, cpus int, ram int) *Error {
	runContainerJSON := rest.RunContainerJSON{ContainerImage: containerImage, Arguments: arguments,
		CPUs: cpus, RAM: ram}

	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserContainerBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, runContainerJSON, nil)
	if err == nil {
		return nil
	} else {
		return NewClientError(err)
	}
}
