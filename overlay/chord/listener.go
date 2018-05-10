package chord

import (
	log "github.com/Sirupsen/logrus"
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/util"
	"github.com/strabox/go-chord"
	"math/big"
)

type Listener struct {
	thisNode nodeAPI.OverlayMembership // Caravela Node on top of the chord overlay
}

func (cl *Listener) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	if local != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(local.Id)
		cl.thisNode.AddTrader(local.Id)
		log.Debugf(util.LogTag("[Chord]")+"Local Node: ID: %s IP: %s", idToPrint.String(), local.Host)
	}
	if remoteNew != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(remoteNew.Id)
		log.Debugf(util.LogTag("[Chord]")+"Remote Node: ID: %s IP: %s", idToPrint.String(), remoteNew.Host)
	}
	if remotePrev != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(remotePrev.Id)
		log.Debugf(util.LogTag("[Chord]")+"Previous Remote Node: ID: %s IP: %s", idToPrint.String(), remotePrev.Host)
	}
}

func (cl *Listener) Leaving(local, predecessor, successor *chord.Vnode) {
	log.Debug(util.LogTag("[Chord]") + "I am leaving!!")
}

func (cl *Listener) PredecessorLeaving(local, remote *chord.Vnode) {
	log.Debug(util.LogTag("[Chord]") + "Current predecessor is leaving!!")
}

func (cl *Listener) SuccessorLeaving(local, remote *chord.Vnode) {
	log.Debug(util.LogTag("[Chord]") + "A successor is leaving!!")
}

func (cl *Listener) Shutdown() {
	log.Debug(util.LogTag("[Chord]") + "Shutting Down!!")
}
