package chord

import (
	log "github.com/Sirupsen/logrus"
	nodeAPI "github.com/strabox/caravela/node/api"
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
	// Caravela Node layer on top of the chord overlay.
	thisNode nodeAPI.OverlayMembership
}

/*
Fired when the a new predecessor of the local node appears in the overlay.
*/
func (cl *Listener) NewPredecessor(local, newPredecessor, previousPredecessor *chord.Vnode) {
	if local != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(local.Id)
		cl.thisNode.AddTrader(local.Id)
		log.Debugf(util.LogTag("[Chord]")+"Local Node: ID: %s IP: %s", idToPrint.String(), local.Host)
	}
	if newPredecessor != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(newPredecessor.Id)
		log.Debugf(util.LogTag("[Chord]")+"Remote Node: ID: %s IP: %s", idToPrint.String(), newPredecessor.Host)
	}
	if previousPredecessor != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(previousPredecessor.Id)
		log.Debugf(util.LogTag("[Chord]")+"Previous Remote Node: ID: %s IP: %s", idToPrint.String(), previousPredecessor.Host)
	}
}

/*
Fired when the local node is leaving the chord overlay.
*/
func (cl *Listener) Leaving(local, predecessor, successor *chord.Vnode) {
	// DO NOTHING FOR NOW
	log.Debug(util.LogTag("[Chord]") + "I am leaving!!")
}

/*
Fired when the current predecessor of the local node is leaving the chord overlay.
*/
func (cl *Listener) PredecessorLeaving(local, remote *chord.Vnode) {
	// DO NOTHING FOR NOW
	log.Debug(util.LogTag("[Chord]") + "Current predecessor is leaving!!")
}

/*
Fired when a current successor of the local node is leaving the chord overlay.
*/
func (cl *Listener) SuccessorLeaving(local, remote *chord.Vnode) {
	// DO NOTHING FOR NOW
	log.Debug(util.LogTag("[Chord]") + "A successor is leaving!!")
}

/*
Fired when when ?????
*/
func (cl *Listener) Shutdown() {
	// DO NOTHING FOR NOW
	log.Debug(util.LogTag("[Chord]") + "Shutting Down!!")
}
