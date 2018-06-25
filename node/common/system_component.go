package common

import "sync"

/*
Provides a common interface for all the internal independent components of a node.
Providing the common interface of a component.
*/
type Component interface {
	Start()          // Starts the component
	Stop()           // Stops the component
	isWorking() bool // Verifies if the component is working
}

/*
Base object for all system's internal components.
*/
type SystemSubComponent struct {
	mutex   sync.Mutex
	working bool
}

func (comp *SystemSubComponent) Started(startFunction func()) {
	comp.mutex.Lock()
	defer comp.mutex.Unlock()

	if !comp.working {
		comp.working = true
		if startFunction != nil {
			startFunction()
		}
	}
}

func (comp *SystemSubComponent) Stopped(stopFunction func()) {
	comp.mutex.Lock()
	defer comp.mutex.Unlock()

	if comp.working {
		comp.working = false
		if stopFunction != nil {
			stopFunction()
		}
	}
}

func (comp *SystemSubComponent) Working() bool {
	comp.mutex.Lock()
	defer comp.mutex.Unlock()

	return comp.working
}
