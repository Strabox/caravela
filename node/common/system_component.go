package common

import "sync"

// Provides a common interface for all the internal independent components of a node.
// Providing the common interface of a component.
type Component interface {
	// Starts the component
	Start(simulation bool)
	// Stops the component
	Stop()
	// Verifies if the component is working
	isWorking() bool
}

// Base object for all system's internal components.
type NodeComponent struct {
	mutex      sync.Mutex
	working    bool
	simulation bool
}

//func NewNodeComponent() *NodeComponent

func (comp *NodeComponent) Started(simulation bool, startFunction func()) {
	comp.mutex.Lock()
	defer comp.mutex.Unlock()

	comp.simulation = simulation

	if !comp.working {
		comp.working = true
		if startFunction != nil {
			startFunction()
		}
	}
}

func (comp *NodeComponent) Stopped(stopFunction func()) {
	comp.mutex.Lock()
	defer comp.mutex.Unlock()

	if comp.working {
		comp.working = false
		if stopFunction != nil {
			stopFunction()
		}
	}
}

func (comp *NodeComponent) Working() bool {
	comp.mutex.Lock()
	defer comp.mutex.Unlock()

	return comp.working || comp.simulation
}
