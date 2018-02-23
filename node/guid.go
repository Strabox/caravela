package node

import (
	"fmt"
	"math/rand"
)

/*
Guid - Represents the global unique identifier of each node
*/
type Guid struct {
	key []byte
}

const GUID_BYTES_SIZE = 20 // 160-bits for now to maintain compatibility with chord implementation

func NewGuid(bytesSize uint) *Guid {
	var guid *Guid = &Guid{}
	guid.key = make([]byte, bytesSize, bytesSize)

	//Randomly generate the bytes
	rand.Read(guid.key)

	return guid
}

func (guid *Guid) GetKey() []byte {
	return guid.key
}

func (guid *Guid) PrintDecimal() {
	fmt.Println(guid.key)
}
