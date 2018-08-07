package discovery

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/backend"
	"github.com/strabox/caravela/node/discovery/random"
	"github.com/strabox/caravela/node/discovery/smart"
	"github.com/strabox/caravela/node/external"
	"strings"
)

// DiscoveryBackendFactory represents a method that creates a new discovery backend.
type BackendFactory func(config *configuration.Configuration, overlay external.Overlay,
	client external.Caravela, resourcesMap *resources.Mapping, maxResources resources.Resources) (backend.Discovery, error)

// manageOffers holds all the registered discovery backends available.
var discoveryBackends = make(map[string]BackendFactory)

// init initializes our predefined offers managers.
func init() {
	RegisterDiscoveryBackend("chord-smart", smart.NewSmartDiscovery)
	RegisterDiscoveryBackend("chord-random", random.NewRandomDiscovery)
}

// RegisterOffersStrategy can be used to register a discovery backend in order to be available.
func RegisterDiscoveryBackend(discBackendName string, factory BackendFactory) {
	if factory == nil {
		log.Panic("nil offers factory registering")
	}
	_, exist := discoveryBackends[discBackendName]
	if exist {
		log.Warnf("offers strategy %s is being overridden", discBackendName)
	}
	discoveryBackends[discBackendName] = factory
}

// CreateDiscoveryBackend is used to obtain a discovery backend based on the configurations.
func CreateDiscoveryBackend(config *configuration.Configuration, overlay external.Overlay,
	client external.Caravela, resourcesMap *resources.Mapping, maxResources resources.Resources) backend.Discovery {
	configuredDiscoveryBackend := config.DiscoveryBackend()

	discoveryFactory, exist := discoveryBackends[configuredDiscoveryBackend]
	if !exist {
		existingBackends := make([]string, len(discoveryBackends))
		for backendsName := range discoveryBackends {
			existingBackends = append(existingBackends, backendsName)
		}
		err := errors.New(fmt.Sprintf("Invalid %s discovery backend. Backend available: %s",
			configuredDiscoveryBackend, strings.Join(existingBackends, ", ")))
		log.Panic(err)
	}

	discoveryBackends, err := discoveryFactory(config, overlay, client, resourcesMap, maxResources)
	if err != nil {
		log.Panic(err)
	}

	return discoveryBackends
}
