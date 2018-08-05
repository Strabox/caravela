package configuration

import (
	"context"
	systemConfiguration "github.com/strabox/caravela/configuration"
)

type Configurations interface {
	Configuration(ctx context.Context) *systemConfiguration.Configuration
}
