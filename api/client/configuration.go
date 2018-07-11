package client

import (
	"time"
)

// CARAVELA's client configuration
type Configuration struct {
	caravelaInstanceIP   string // IP of the CARAVELA's Daemon that will receive the request
	caravelaInstancePort int    // Port of the CARAVELA's Daemon that will receive the request

	httpRequestTimeout time.Duration // HTTP requests timeout
}

func DefaultConfiguration(caravelaInstanceIP string) *Configuration {
	return &Configuration{
		caravelaInstanceIP:   caravelaInstanceIP,
		caravelaInstancePort: 8001,

		httpRequestTimeout: 3 * time.Second,
	}
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
