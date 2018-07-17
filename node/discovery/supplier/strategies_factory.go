package supplier

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/configuration"
	"strings"
)

type ManageOffersFactory func(config *configuration.Configuration) (OffersManager, error)

func newDefaultChordManageOffers(config *configuration.Configuration) (OffersManager, error) {
	return &DefaultChordOffersManager{
		configs: config,
	}, nil
}

var manageOffersStrategies = make(map[string]ManageOffersFactory)

func RegisterOffersStrategy(strategyName string, factory ManageOffersFactory) {
	if factory == nil {
		log.Panic("nil offers factory registering")
	}
	_, exist := manageOffersStrategies[strategyName]
	if exist {
		log.Warnf("offers strategy %s is being overridden", strategyName)
	}
	manageOffersStrategies[strategyName] = factory
}

func initOffersFactory() {
	RegisterOffersStrategy("chordDefault", newDefaultChordManageOffers)
}

func CreateOffersStrategy(config *configuration.Configuration) OffersManager {
	configuredStrategy := config.OffersStrategy()
	strategyFactory, exist := manageOffersStrategies[configuredStrategy]
	if !exist {
		existingStrategies := make([]string, len(manageOffersStrategies))
		for strategyName := range manageOffersStrategies {
			existingStrategies = append(existingStrategies, strategyName)
		}
		err := errors.New(fmt.Sprintf("Invalid %s offer strategy. Strategies available: %s",
			configuredStrategy, strings.Join(existingStrategies, ", ")))
		log.Errorf(err.Error())
		panic(err)
	}

	offersStrategy, _ := strategyFactory(config)
	return offersStrategy
}
