package client

import (
	"github.com/strabox/caravela/api/rest"
	"net/http"
)

type Caravela struct {
	httpClient *http.Client   // Http client to send requests into Caravela's daemon
	config     *Configuration // Configuration parameters for the CARAVELA client
}

func NewCaravela(caravelaHostIP string) *Caravela {
	res := &Caravela{}
	res.config = DefaultConfiguration(caravelaHostIP)
	res.httpClient = &http.Client{
		Timeout: res.config.HttpRequestTimeout(),
	}
	return res
}

func (client *Caravela) Run(containerImage string, arguments []string, cpus int, ram int) *Error {
	runContainerJSON := rest.RunContainerJSON{containerImage, arguments, cpus, ram}

	url := rest.BuildHttpURL(false, client.config.CaravelaInstanceIP(), client.config.CaravelaInstancePort(),
		rest.UserContainerBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, runContainerJSON, nil)
	if err == nil {
		return nil
	} else {
		return NewClientError(err)
	}
}
