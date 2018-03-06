package local

import (
	"github.com/strabox/caravela/node/resources"
)

type LocalNode interface {
	ResourcesMap() *resources.ResourcesMap
}