package chord

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/util"
	"github.com/strabox/go-chord"
	"math/big"
)

/*
Used to handle events fired by the chord overlay.
The listener let the important events bubble up into Node layer using a provided interface
called OverlayMembership.
*/
type Listener struct {
	chordOverlay *Overlay // Chord's overlay
}

/*
Fired when the a new predecessor of the local node appears in the overlay.
*/
func (cl *Listener) NewPredecessor(local, newPredecessor, previousPredecessor *chord.Vnode) {
	if local != nil && newPredecessor != nil && previousPredecessor == nil {
		// First time a virtual node is entering in the ring
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(local.Id)
		predecessorNode := overlay.NewNode(newPredecessor.Host, newPredecessor.Id)
		cl.chordOverlay.newLocalVirtualNode(local.Id, predecessorNode)
		log.Debugf(util.LogTag("Chord")+"New LVNode: ID: %s IP: %s", idToPrint.String(), local.Host)
	} else if local != nil && newPredecessor != nil && previousPredecessor != nil {
		// New predecessor for a existing node
		predecessorNode := overlay.NewNode(newPredecessor.Host, newPredecessor.Id)
		cl.chordOverlay.predecessorNodeChanged(local.Id, predecessorNode)
	}
}

/*
Fired when the local node is leaving the chord overlay.
*/
func (cl *Listener) Leaving(local, predecessor, successor *chord.Vnode) {
	// DO NOTHING FOR NOW
	log.Debug(util.LogTag("Chord") + "I am leaving!!")
}

/*
Fired when the current predecessor of the local node is leaving the chord overlay.
*/
func (cl *Listener) PredecessorLeaving(local, remote *chord.Vnode) {
	// DO NOTHING FOR NOW
	log.Debug(util.LogTag("Chord") + "Current predecessor is leaving!!")
}

/*
Fired when a current successor of the local node is leaving the chord overlay.
*/
func (cl *Listener) SuccessorLeaving(local, remote *chord.Vnode) {
	// DO NOTHING FOR NOW
	log.Debug(util.LogTag("Chord") + "A successor is leaving!!")
}

/*
Fired when when one node decided to shutdown the chord ring system.
Do the shutdown message propagates to all the nodes ??
*/
func (cl *Listener) Shutdown() {
	// DO NOTHING FOR NOW
	log.Debug(util.LogTag("Chord") + "Shutting Down!!")
}
