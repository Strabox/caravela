package node

<<<<<<< Upstream, based on origin/master
import (

)

/*
Supplier handles all the logic of offering the own resources and receiving requests to deploy containers
*/
type Supplier struct {
	node				*Node				// Node of the supplier 
	resources   		*Resources			// The maximum resources that the node can offer
	resourcesAvailable 	*Resources 			// The current resources that the node have available
	containerManager 	*ContainersManager	//
}

=======
import ()

type Supplier struct {
}
>>>>>>> 5c6e03b Refactoring, preparing the base for the CARAVELA
