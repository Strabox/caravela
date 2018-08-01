package client

import (
	"time"
)

const defaultCaravelaInstancePrt = 8001
const defaultHTTPRequestTimeoutSecs = 5

// Configuration holds the configuration parameters for the CARAVELA's client.
type Configuration struct {
	caravelaInstanceIP   string // IP address of the CARAVELA's Daemon that will receive the request.
	caravelaInstancePort int    // Port of the CARAVELA's Daemon that will receive the request.

	httpRequestTimeout time.Duration // HTTP requests timeout.
}

// DefaultConfig creates a new configuration structure with the default values.
func DefaultConfig(caravelaInstanceIP string) *Configuration {
	return &Configuration{
		caravelaInstanceIP:   caravelaInstanceIP,
		caravelaInstancePort: defaultCaravelaInstancePrt,

		httpRequestTimeout: defaultHTTPRequestTimeoutSecs * time.Second,
	}
}

// CaravelaInstanceIP returns the IP address to where send the API requests.
func (config *Configuration) CaravelaInstanceIP() string {
	return config.caravelaInstanceIP
}

// CaravelaInstancePort returns the port to where send the API requests.
func (config *Configuration) CaravelaInstancePort() int {
	return config.caravelaInstancePort
}

// SetCaravelaInstancePort sets the port to where send the API requests.
func (config *Configuration) SetCaravelaInstancePort(newPort int) {
	config.caravelaInstancePort = newPort
}

// RequestTimeout returns the timeout for the API requests.
func (config *Configuration) RequestTimeout() time.Duration {
	return config.httpRequestTimeout
}

// SetRequestTimeout sets the timeout for the API requests.
func (config *Configuration) SetRequestTimeout(newReqTimeout time.Duration) {
	config.httpRequestTimeout = newReqTimeout
}
