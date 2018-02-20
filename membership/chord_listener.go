package membership

import (
	"fmt"
	"github.com/strabox/go-chord"
)

type ChordListner struct {
}

func (*ChordListner) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	fmt.Println("[CHORD-New Predecessor]")
	if local != nil {
		fmt.Printf("Local Node: [IP:%s ID:%s]\n", local.String(), local.Host)
	}
	if remoteNew != nil {
		fmt.Printf("Remote Node: [IP:%s ID:%s]\n", remoteNew.String(), remoteNew.Host)
	}
	if remotePrev != nil {
		fmt.Printf("Previous Remote Node: [IP:%s ID:%s]\n", remotePrev.String(), remotePrev.Host)
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
