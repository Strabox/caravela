package chord

import (
	"github.com/bluele/go-chord"
	nodeAPI "github.com/strabox/caravela/node/api"
	"log"
	"math/big"
)

type Listener struct {
	thisNode nodeAPI.LocalNode
}

func (cl *Listener) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	log.Println("[Chord Overlay] New Predecessor!!")
	if local != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(local.Id)
		cl.thisNode.AddTrader(local.Id)
		log.Printf("Local Node: [ID:%s IP:%s]\n", idToPrint.String(), local.Host)
	}
	if remoteNew != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(remoteNew.Id)
		log.Printf("Remote Node: [ID:%s IP:%s]\n", idToPrint.String(), remoteNew.Host)
	}
	if remotePrev != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(remotePrev.Id)
		log.Printf("Previous Remote Node: [ID:%s IP:%s]\n", idToPrint.String(), remotePrev.Host)
	}
}

func (cl *Listener) Leaving(local, pred, succ *chord.Vnode) {
	log.Println("[Chord Overlay] I am leaving!!")
}

func (cl *Listener) PredecessorLeaving(local, remote *chord.Vnode) {
	log.Println("[Chord Overlay] Current predecessor is leaving!!")
}

func (cl *Listener) SuccessorLeaving(local, remote *chord.Vnode) {
	log.Println("[Chord Overlay] A successor is leaving!!")
}

func (cl *Listener) Shutdown() {
	log.Println("[Chord Overlay] Shutting Down!!")
}
