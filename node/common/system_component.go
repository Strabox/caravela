package common

import "sync"

// Component provides a common interface for all the internal independent components of a node.
// Providing the common interface of a component.
type Component interface {
	// Start starts the component.
	Start()
	// Stop stops the component.
	Stop()
	// IsWorking verifies if the component is working.
	IsWorking() bool
}

// NodeComponent is the base object for all system's internal components.
type NodeComponent struct {
	mutex      sync.Mutex
	working    bool
	simulation bool
}

func (n *NodeComponent) Started(simulation bool, startFunction func()) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.simulation = simulation

	if !n.working {
		n.working = true
		if startFunction != nil {
			startFunction()
		}
	}
}

func (n *NodeComponent) Stopped(stopFunction func()) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if n.working {
		n.working = false
		if stopFunction != nil {
			stopFunction()
		}
	}
}

func (n *NodeComponent) Working() bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	return n.working || n.simulation
}
