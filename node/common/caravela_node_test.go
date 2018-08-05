package common

import (
	"github.com/strabox/caravela/node/common/guid"
	"github.com/stretchr/testify/assert"
	"testing"
)

const ipAddressTest = "127.0.0.1"
const guidTest = 777

func TestNewRemoteNode(t *testing.T) {
	guid := *guid.NewGUIDInteger(guidTest)
	node := NewRemoteNode(ipAddressTest, guid)

	assert.Equal(t, ipAddressTest, node.IP(), "Node's IP is incorrect!")
	assert.Equal(t, guid, *node.GUID(), "Node's GUID is incorrect!")
}
