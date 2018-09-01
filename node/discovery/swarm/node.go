package swarm

import "github.com/strabox/caravela/node/common/resources"

type node struct {
	availableResources resources.Resources
	ip                 string
}
