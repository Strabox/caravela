package storage

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/configuration"
	"strings"
)

// BackendFactory represents a method that creates a new image storage backend.
type BackendFactory func() (Backend, error)

// backends holds all the registered image storage backends available.
var backends = make(map[string]BackendFactory)

// init initializes our predefined storage backends.
func init() {
	RegisterBackend("DockerHub", newDockerHubBackend)
}

// RegisterBackend can be used to register a new storage backend in order to be available.
func RegisterBackend(backendName string, factory BackendFactory) {
	if factory == nil {
		log.Panic("nil storage backend factory registering")
	}
	_, exist := backends[backendName]
	if exist {
		log.Warnf("storage backend %s is being overridden", backendName)
	}
	backends[backendName] = factory
}

// CreateBackend is used to obtain a storage backend based on the configurations.
func CreateBackend(config *configuration.Configuration) Backend {
	configuredBackend := config.ImagesStorageBackend()

	backendFactory, exist := backends[configuredBackend]
	if !exist {
		existingBackends := make([]string, len(backends))
		for backendName := range backends {
			existingBackends = append(existingBackends, backendName)
		}
		err := errors.New(fmt.Sprintf("Invalid %s storage backend. Backends available: %s",
			configuredBackend, strings.Join(existingBackends, ", ")))
		log.Panic(err)
	}

	backend, err := backendFactory()
	if err != nil {
		log.Panic(err)
	}

	return backend
}
