package client

import (
	"time"
)

/*
Caravela client configuration struct.
*/
type Configuration struct {
	// IP of the Caravela's Daemon that will receive the request
	caravelaInstanceIP string
	// Port of the Caravela's Daemon that will receive the request
	caravelaInstancePort int

	// HTTP requests timeout
	httpRequestTimeout time.Duration
}

func DefaultConfiguration(caravelaInstanceIP string) *Configuration {
	res := &Configuration{}
	res.caravelaInstanceIP = caravelaInstanceIP
	res.caravelaInstancePort = 8001

	res.httpRequestTimeout = 3 * time.Second
	return res
}

func (config *Configuration) CaravelaInstanceIP() string {
	return config.caravelaInstanceIP
}

func (config *Configuration) CaravelaInstancePort() int {
	return config.caravelaInstancePort
}

func (config *Configuration) HttpRequestTimeout() time.Duration {
	return config.httpRequestTimeout
}
