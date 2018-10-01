package factory

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/overlay/chord"
	"strings"
)

// BackendFactory represents a method that creates a new overlay object.
type Factory func(config *configuration.Configuration) (overlay.Overlay, error)

// overlayFactories holds all the registered overlays available.
var overlayFactories = make(map[string]Factory)

// init initializes our predefined overlays.
func init() {
	Register("chord", chord.New)
}

// Register can be used to register a new overlay in order to be available.
func Register(overlayName string, factory Factory) {
	if factory == nil {
		log.Panic("nil overlay factory registering")
	}
	_, exist := overlayFactories[overlayName]
	if exist {
		log.Warnf("overlay %s is being overridden", overlayName)
	}
	overlayFactories[overlayName] = factory
}

// Create is used to create an overlay based on the configurations.
func Create(config *configuration.Configuration) overlay.Overlay {
	configuredOverlay := config.OverlayName()

	overlayFactory, exist := overlayFactories[configuredOverlay]
	if !exist {
		existingOverlays := make([]string, len(overlayFactories))
		for backendName := range overlayFactories {
			existingOverlays = append(existingOverlays, backendName)
		}
		err := errors.New(fmt.Sprintf("Invalid %s overlay. Overlays available: %s",
			configuredOverlay, strings.Join(existingOverlays, ", ")))
		log.Panic(err)
	}

	overlay, err := overlayFactory(config)
	if err != nil {
		log.Panic(err)
	}

	return overlay
}
