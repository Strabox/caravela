package client

import (
	"github.com/strabox/caravela/configuration"
	"time"
)

/*
Caravela client configuration struct.
*/
type Configuration struct {
	caravelaInstanceIP   string
	caravelaInstancePort int

	// Http client configuration
	httpContentType    string
	httpRequestTimeout time.Duration
}

func DefaultConfiguration(caravelaInstanceIP string) *Configuration {
	res := &Configuration{}
	res.caravelaInstanceIP = caravelaInstanceIP
	res.caravelaInstancePort = configuration.APIPort

	res.httpContentType = "application/json"
	res.httpRequestTimeout = 3 * time.Second
	return res
}

func (config *Configuration) CaravelaInstanceIP() string {
	return config.caravelaInstanceIP
}

func (config *Configuration) CaravelaInstancePort() int {
	return config.caravelaInstancePort
}

func (config *Configuration) HttpContentType() string {
	return config.httpContentType
}

func (config *Configuration) HttpRequestTimeout() time.Duration {
	return config.httpRequestTimeout
}
