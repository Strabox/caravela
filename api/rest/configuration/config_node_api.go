package configuration

import (
	systemConfiguration "github.com/strabox/caravela/configuration"
)

type Configurations interface {
	Configuration() *systemConfiguration.Configuration
}
