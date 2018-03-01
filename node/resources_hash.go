package node

import (
	"hash"
)

/*
ResourcesHash - It is used to pass to the Chord implementation
*/
type ResourcesHash struct {
	hash []byte
}

func GetHash() hash.Hash {
	return NewResourcesHash(GUID_BITS_SIZE / 8)
}

func NewResourcesHash(bytesSize uint) *ResourcesHash {
	hash := &ResourcesHash{}
	hash.hash = make([]byte, bytesSize, bytesSize)
	return hash
}

// ##################### Hash Interface #######################

func (rh *ResourcesHash) Write(p []byte) (n int, err error) {
	for index, value := range p {
		rh.hash[index] = value
	}
	return 0, nil
}

func (rh *ResourcesHash) Sum(b []byte) []byte {
	// Only return the hash we wrote
	return rh.hash
}

func (rh *ResourcesHash) Reset() {
	for index, _ := range rh.hash {
		rh.hash[index] = 0
	}
}

func (rh *ResourcesHash) Size() int {
	return 0
}

func (rh *ResourcesHash) BlockSize() int {
	return 0
}
