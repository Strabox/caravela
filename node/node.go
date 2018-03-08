//Node represents a chord node and is the manager of the node
package node

import (
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/node/discovery/supplier"
	"github.com/strabox/caravela/node/discovery/trader"
	"github.com/strabox/caravela/node/configuration"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/api/client"
	"fmt"
)

type Node struct {
	resourcesMap *resources.ResourcesMap
	supplier     *supplier.Supplier
	traders      []*trader.Trader
}

func NewNode(config *configuration.Configuration, overlay overlay.Overlay, client client.CaravelaClient, 
		rm *resources.ResourcesMap, maxNumTraders int, maxResourcesAvailable resources.Resources) *Node {
	res := &Node{}
	res.supplier = supplier.NewSupplier(config, overlay, client, rm, maxResourcesAvailable)
	res.resourcesMap = rm
	
	res.traders = make([]*trader.Trader, maxNumTraders)
	for index, _ := range res.traders{
		res.traders[index] = nil
	}
	return res
}

func (node *Node) AddTrader(guidBytes []byte)  {
	guidObj := guid.NewGuidBytes(guidBytes)
	traderResources,_ := node.resourcesMap.ResourcesByGuid(*guidObj)
	newTrader := trader.NewTrader(*guidObj, *traderResources)
	for index, value := range node.traders {
		if value == nil {
			fmt.Printf("[Node] New Trader: %s | Resources: %s\n", guidObj.String(), traderResources.ToString())
			node.traders[index] = newTrader
			break;
		}
	}
}
