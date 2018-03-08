package chordoverlay

import (
	"fmt"
	"github.com/bluele/go-chord"
	"github.com/strabox/caravela/node/local"
	"math/big"
)

type ChordListner struct {
	thisNode local.LocalNode
}

func (cl *ChordListner) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	fmt.Println("[Chord Overlay] New Predecessor!!")
	if local != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(local.Id)
		cl.thisNode.AddTrader(local.Id)
		fmt.Printf("Local Node: [ID:%s IP:%s]\n", idToPrint.String(), local.Host)
	}
	if remoteNew != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(remoteNew.Id)
		fmt.Printf("Remote Node: [ID:%s IP:%s]\n", idToPrint.String(), remoteNew.Host)
	}
	if remotePrev != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(remotePrev.Id)
		fmt.Printf("Previous Remote Node: [ID:%s IP:%s]\n", idToPrint.String(), remotePrev.Host)
	}
}

func (cl *ChordListner) Leaving(local, pred, succ *chord.Vnode) {
	fmt.Println("[Chord Overlay] I am leaving!!")
}

func (cl *ChordListner) PredecessorLeaving(local, remote *chord.Vnode) {
	fmt.Println("[Chord Overlay] Current predecessor is leaving!!")
}

func (cl *ChordListner) SuccessorLeaving(local, remote *chord.Vnode) {
	fmt.Println("[Chord Overlay] A successor is leaving!!")
}

func (cl *ChordListner) Shutdown() {
	fmt.Println("[Chord Overlay] Shutting Down!!")
}
