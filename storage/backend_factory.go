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
func RegisterBackend(strategyName string, factory BackendFactory) {
	if factory == nil {
		log.Panic("nil offers factory registering")
	}
	_, exist := backends[strategyName]
	if exist {
		log.Warnf("offers strategy %s is being overridden", strategyName)
	}
	backends[strategyName] = factory
}

// CreateBackend is used to obtain a storage backend based on the configurations.
func CreateBackend(config *configuration.Configuration) Backend {
	configuredBackend := config.ImagesStorageBackend()

	strategyFactory, exist := backends[configuredBackend]
	if !exist {
		existingBackends := make([]string, len(backends))
		for backendName := range backends {
			existingBackends = append(existingBackends, backendName)
		}
		err := errors.New(fmt.Sprintf("Invalid %s storage backend. Backends available: %s",
			configuredBackend, strings.Join(existingBackends, ", ")))
		log.Panic(err)
	}

	offersStrategy, err := strategyFactory()
	if err != nil {
		log.Panic(err)
	}

	return offersStrategy
}
