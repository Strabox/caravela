package supplier

import (
	"context"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/node/external"
	overlayTypes "github.com/strabox/caravela/overlay/types"
	"github.com/strabox/caravela/util"
	"math/rand"
	"sync"
	"time"
)

var randomGenerator = rand.New(util.NewSourceSafe(rand.NewSource(time.Now().Unix())))

type SystemResourcePartitions struct {
	partitionsState sync.Map
}

func NewSystemResourcePartitions() *SystemResourcePartitions {
	return &SystemResourcePartitions{
		partitionsState: sync.Map{},
	}
}

func (s *SystemResourcePartitions) Try(targetResPartition resources.Resources) bool {
	if partition, exist := s.partitionsState.Load(targetResPartition); exist {
		if partitionState, ok := partition.(*ResourcePartitionState); ok {
			return partitionState.Try()
		}
	} else {
		newPartitionState := NewResourcePartitionState(15)
		s.partitionsState.Store(targetResPartition, newPartitionState)
		return newPartitionState.Try()
	}
	return true
}

func (s *SystemResourcePartitions) Hit(resPartition resources.Resources) {
	if partition, exist := s.partitionsState.Load(resPartition); exist {
		if partitionState, ok := partition.(*ResourcePartitionState); ok {
			partitionState.Hit()
		}
	}
}

func (s *SystemResourcePartitions) Miss(resPartition resources.Resources) {
	if partition, exist := s.partitionsState.Load(resPartition); exist {
		if partitionState, ok := partition.(*ResourcePartitionState); ok {
			partitionState.Miss()
		}
	}
}

type ResourcePartitionState struct {
	totalTries int
	hits       int
	mutex      sync.RWMutex
}

func NewResourcePartitionState(totalStat int) *ResourcePartitionState {
	return &ResourcePartitionState{
		totalTries: totalStat,
		hits:       totalStat,
		mutex:      sync.RWMutex{},
	}
}

func (rps *ResourcePartitionState) Try() bool {
	rps.mutex.RLock()
	defer rps.mutex.RUnlock()

	hitProbability := int((float64(rps.hits) / float64(rps.totalTries)) * 100)
	randTry := randomGenerator.Intn(100)
	if randTry < hitProbability {
		return true
	}
	secondChance := randomGenerator.Intn(100)
	if secondChance <= 10 {
		return true
	}
	return false
}

func (rps *ResourcePartitionState) Hit() {
	if rps.hits < rps.totalTries {
		rps.hits++
	}
}

func (rps *ResourcePartitionState) Miss() {
	if rps.hits > 0 {
		rps.hits--
	}
}

type SmartChordOffersManager struct {
	resourcesPartitions *SystemResourcePartitions
	configs             *configuration.Configuration
	resourcesMapping    *resources.Mapping
	overlay             external.Overlay
	remoteClient        external.Caravela
}

func newSmartChordManageOffers(config *configuration.Configuration) (OffersManager, error) {
	return &SmartChordOffersManager{
		resourcesPartitions: NewSystemResourcePartitions(),
		configs:             config,
	}, nil
}

func (man *SmartChordOffersManager) Init(resourcesMap *resources.Mapping, overlay external.Overlay, remoteClient external.Caravela) {
	man.resourcesMapping = resourcesMap
	man.overlay = overlay
	man.remoteClient = remoteClient
}

func (man *SmartChordOffersManager) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	var destinationGUID *guid.GUID = nil
	findPhase := 0
	availableOffers := make([]types.AvailableOffer, 0)
	for {
		var err error = nil

		if findPhase == 0 { // Random trader inside resources zone
			destinationGUID, err = man.resourcesMapping.RandGUIDSearch(targetResources)
			if err != nil { // System can't handle that many resources
				return availableOffers
			}
		} else { // Random trader in higher resources zone
			destinationGUID, err = man.resourcesMapping.HigherRandGUIDSearch(*destinationGUID, targetResources)
			if err != nil { // No more resource partitions to search
				return availableOffers
			}
		}

		targetResPartition := *man.resourcesMapping.ResourcesByGUID(*destinationGUID)
		log.Debugf(util.LogTag("SUPPLIER")+"FINDING OFFERS %s", targetResPartition)

		if man.resourcesPartitions.Try(targetResPartition) {
			overlayNodes, _ := man.overlay.Lookup(ctx, destinationGUID.Bytes())
			overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)

			for _, node := range overlayNodes {
				offers, err := man.remoteClient.GetOffers(
					ctx,
					&types.Node{},
					&types.Node{IP: node.IP(), GUID: guid.NewGUIDBytes(node.GUID()).String()},
					true)
				if err == nil && len(offers) != 0 {
					availableOffers = append(availableOffers, offers...)
					man.resourcesPartitions.Hit(targetResPartition)
					break
				} else if err == nil && len(offers) == 0 {
					man.resourcesPartitions.Miss(targetResPartition)
				}
			}

			if len(availableOffers) > 0 {
				return availableOffers
			}
		}

		findPhase++
	}
}

func (man *SmartChordOffersManager) CreateOffer(newOfferID int64, availableResources resources.Resources) (*supplierOffer, error) {
	var err error
	var overlayNodes []*overlayTypes.OverlayNode = nil

	destinationGUID, err := man.resourcesMapping.RandGUIDOffer(availableResources)
	if err != nil {
		return nil, errors.New("no nodes capable of handle this offer resources")
	}
	overlayNodes, _ = man.overlay.Lookup(context.Background(), destinationGUID.Bytes())
	overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)

	// .. try search nodes in the beginning of the original target resource range region
	if len(overlayNodes) == 0 {
		destinationGUID, err := man.resourcesMapping.FirstGUIDOffer(availableResources)
		if err != nil {
			return nil, err
		}
		overlayNodes, _ = man.overlay.Lookup(context.Background(), destinationGUID.Bytes())
		overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// ... try search for random nodes that handle less powerful resource combinations
	for len(overlayNodes) == 0 {
		destinationGUID, err = man.resourcesMapping.LowerRandGUIDOffer(*destinationGUID, availableResources)
		if err != nil {
			log.Errorf(util.LogTag("SUPPLIER")+"NO NODES to handle resources offer: %s. Error: %s",
				availableResources.String(), err)
			return nil, errors.New("no nodes available to accept offer") // Wait fot the next tick to try supply resources
		}
		overlayNodes, _ = man.overlay.Lookup(context.Background(), destinationGUID.Bytes())
		overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// Chose the first node returned by the overlay API
	chosenNode := overlayNodes[0]
	chosenNodeGUID := guid.NewGUIDBytes(chosenNode.GUID())

	err = man.remoteClient.CreateOffer(context.Background(),
		&types.Node{IP: man.configs.HostIP(), GUID: ""},
		&types.Node{IP: chosenNode.IP(), GUID: chosenNodeGUID.String()},
		&types.Offer{
			ID:        newOfferID,
			Amount:    1,
			Resources: types.Resources{CPUs: availableResources.CPUs(), RAM: availableResources.RAM()}})
	if err == nil {
		return newSupplierOffer(common.OfferID(newOfferID), 1, availableResources, chosenNode.IP(), *chosenNodeGUID), nil
	}

	return nil, errors.New("impossible advertise offer")
}

// Remove nodes that do not belong to that target GUID partition. (Probably because we were target a partition frontier node)
func (man *SmartChordOffersManager) removeNonTargetNodes(remoteNodes []*overlayTypes.OverlayNode,
	targetGuid guid.GUID) []*overlayTypes.OverlayNode {

	resultNodes := make([]*overlayTypes.OverlayNode, 0)
	targetGuidResources := man.resourcesMapping.ResourcesByGUID(targetGuid)
	for _, remoteNode := range remoteNodes {
		remoteNodeResources := man.resourcesMapping.ResourcesByGUID(*guid.NewGUIDBytes(remoteNode.GUID()))
		if targetGuidResources.Equals(*remoteNodeResources) {
			resultNodes = append(resultNodes, remoteNode)
		}
	}
	return resultNodes
}
