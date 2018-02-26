//Node represents a chord node and is the manager of the node
package node

import ()

type Node struct {
	guid      *Guid
	trader    *Trader
	resources *Resources
}
