package chordoverlay

import (
	"fmt"
	"github.com/bluele/go-chord"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/local"
)

type ChordListner struct {
	thisNode local.LocalNode
}

func (cl *ChordListner) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	fmt.Println("[Chord Overlay] New Predecessor!!")
	if local != nil {
		guid := guid.NewGuidBytes(local.Id)
		resources ,_ := cl.thisNode.ResourcesMap().ResourcesByGuid(*guid)
		fmt.Printf("Local Node: [ID:%s IP:%s Resources:%s]\n", guid.ToString(), local.Host, resources.ToString())
	}
	if remoteNew != nil {
		fmt.Printf("Remote Node: [ID:%s IP:%s]\n", guid.NewGuidBytes(remoteNew.Id).ToString(), remoteNew.Host)
	}
	if remotePrev != nil {
		fmt.Printf("Previous Remote Node: [ID:%s IP:%s]\n", guid.NewGuidBytes(remotePrev.Id).ToString(), remotePrev.Host)
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
