package chordoverlay

import (
	"fmt"
	"github.com/bluele/go-chord"
)

type ChordListner struct {
}

func (*ChordListner) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	fmt.Println("[Chord Overlay] New Predecessor!!")
	if local != nil {
		fmt.Printf("Local Node: [ID:%s IP:%s]\n", local.String(), local.Host)
	}
	if remoteNew != nil {
		fmt.Printf("Remote Node: [ID:%s IP:%s]\n", remoteNew.String(), remoteNew.Host)
	}
	if remotePrev != nil {
		fmt.Printf("Previous Remote Node: [ID:%s IP:%s]\n", remotePrev.String(), remotePrev.Host)
	}
}

func (*ChordListner) Leaving(local, pred, succ *chord.Vnode) {
	fmt.Println("[Chord Overlay] I am leaving!!")
}

func (*ChordListner) PredecessorLeaving(local, remote *chord.Vnode) {
	fmt.Println("[Chord Overlay] Current predecessor is leaving!!")
}

func (*ChordListner) SuccessorLeaving(local, remote *chord.Vnode) {
	fmt.Println("[Chord Overlay] A successor is leaving!!")
}

func (*ChordListner) Shutdown() {
	fmt.Println("[Chord Overlay] Shutting Down!!")
}
