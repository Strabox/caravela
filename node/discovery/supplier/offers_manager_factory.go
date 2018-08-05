package supplier

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/configuration"
	"strings"
)

// ManageOffersFactory represents a method that creates a new offers manager.
type ManageOffersFactory func(config *configuration.Configuration) (OffersManager, error)

// manageOffers holds all the registered offer managers available.
var manageOffers = make(map[string]ManageOffersFactory)

// init initializes our predefined offers managers.
func init() {
	RegisterOffersStrategy("chord-smart", newSmartChordManageOffers)
}

// RegisterOffersStrategy can be used to register a new strategy in order to be available.
func RegisterOffersStrategy(strategyName string, factory ManageOffersFactory) {
	if factory == nil {
		log.Panic("nil offers factory registering")
	}
	_, exist := manageOffers[strategyName]
	if exist {
		log.Warnf("offers strategy %s is being overridden", strategyName)
	}
	manageOffers[strategyName] = factory
}

// CreateOffersStrategy is used to obtain an offers manager based on the configurations.
func CreateOffersStrategy(config *configuration.Configuration) OffersManager {
	configuredStrategy := config.OffersStrategy()

	strategyFactory, exist := manageOffers[configuredStrategy]
	if !exist {
		existingStrategies := make([]string, len(manageOffers))
		for strategyName := range manageOffers {
			existingStrategies = append(existingStrategies, strategyName)
		}
		err := errors.New(fmt.Sprintf("Invalid %s offer strategy. Strategies available: %s",
			configuredStrategy, strings.Join(existingStrategies, ", ")))
		log.Panic(err)
	}

	offersStrategy, err := strategyFactory(config)
	if err != nil {
		log.Panic(err)
	}

	return offersStrategy
}
