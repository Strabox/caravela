package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNodeComponent(t *testing.T) {
	comp := &NodeComponent{}

	assert.False(t, comp.Working(), "Component shouldn't be working")
}

func TestNodeComponent_Started_NotWorking(t *testing.T) {
	comp := &NodeComponent{}
	startFunctionExecution := false
	startFunction := func() {
		startFunctionExecution = true
	}

	comp.Started(false, startFunction)

	assert.True(t, startFunctionExecution, "Start function didn't executed")
	assert.True(t, comp.Working(), "Component should be working")
}

func TestNodeComponent_Started_Working(t *testing.T) {
	comp := &NodeComponent{working: true}
	startFunctionExecution := false
	startFunction := func() {
		startFunctionExecution = true
	}

	comp.Started(false, startFunction)

	assert.False(t, startFunctionExecution, "Start function executed")
	assert.True(t, comp.Working(), "Component should be working")
}

func TestNodeComponent_Stopped_NotWorking(t *testing.T) {
	comp := &NodeComponent{}
	stopFunctionExecution := false
	stopFunction := func() {
		stopFunctionExecution = true
	}

	comp.Stopped(stopFunction)

	assert.False(t, stopFunctionExecution, "Stop function executed")
	assert.False(t, comp.Working(), "Component shouldn't be working")
}

func TestNodeComponent_Stopped_Working(t *testing.T) {
	comp := &NodeComponent{working: true}
	stopFunctionExecution := false
	stopFunction := func() {
		stopFunctionExecution = true
	}

	comp.Stopped(stopFunction)

	assert.True(t, stopFunctionExecution, "Stop function didn't executed")
	assert.False(t, comp.Working(), "Component shouldn't be working")
}

func TestNodeComponent_UnderSimulation(t *testing.T) {
	comp := &NodeComponent{}

	comp.Started(true, func() { /* Do Nothing */ })

	assert.True(t, comp.Working(), "Component should be working (under simulation)")
}
