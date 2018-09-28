package containers

import "github.com/strabox/caravela/node/common/resources"

type supplierLocal interface {
	ObtainResources(offerID int64, resourcesNecessary resources.Resources, numContainersToRun int) bool
	ReturnResources(resources resources.Resources, numContainersStopped int)
}
