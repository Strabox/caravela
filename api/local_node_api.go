package api

import (
	"github.com/strabox/caravela/api/rest/configuration"
	"github.com/strabox/caravela/api/rest/containers"
	"github.com/strabox/caravela/api/rest/discovery"
	"github.com/strabox/caravela/api/rest/scheduling"
	"github.com/strabox/caravela/api/rest/user"
)

// LocalNode exposes all the necessary functionality, of the local node, for the REST API web server.
type LocalNode interface {
	configuration.Configurations
	containers.Containers
	discovery.Discovery
	scheduling.Scheduling
	user.User
}
