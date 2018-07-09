package api

import (
	"github.com/strabox/caravela/configuration"
	nodeAPI "github.com/strabox/caravela/node/api"
)

type Server interface {
	Start(config *configuration.Configuration, thisNode nodeAPI.Node) error
	Stop()
}
