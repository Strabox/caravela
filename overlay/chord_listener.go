package overlay

import (
	"fmt"
	"github.com/strabox/caravela/node"
	"github.com/bluele/go-chord"
)

type ChordListner struct {
}

func (*ChordListner) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	fmt.Println("[CHORD] New Predecessor")
	if local != nil {
		fmt.Printf("Local Node: [ID:%s IP: %s]\n", node.NewGuidBytes(local.Id).ToString(), local.Host)
	}
	if remoteNew != nil {
		fmt.Printf("Remote Node: [ID:%s IP: %s]\n", node.NewGuidBytes(remoteNew.Id).ToString(), remoteNew.Host)
	}
	if remotePrev != nil {
		fmt.Printf("Previous Remote Node: [ID:%s IP: %s]\n", node.NewGuidBytes(remotePrev.Id).ToString(), remotePrev.Host)
	}
}

func (*ChordListner) Leaving(local, pred, succ *chord.Vnode) {
	fmt.Println("[CHORD-Leaving]\n\n")
}

func (*ChordListner) PredecessorLeaving(local, remote *chord.Vnode) {
	fmt.Println("[CHORD-Predecessor Leaving]\n\n")
}

func (*ChordListner) SuccessorLeaving(local, remote *chord.Vnode) {
	fmt.Println("[CHORD-Successor Leaving]\n\n")
}

func (*ChordListner) Shutdown() {
	fmt.Println("[CHORD-Shutdown]\n\n")
}
