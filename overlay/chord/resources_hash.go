package chord

import (
	log "github.com/Sirupsen/logrus"
	"math/rand"
	"sync"
	"time"
)

// Used to generate random hashes => GUID for the joining nodes
var randomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
var randomGeneratorMutex = sync.Mutex{}

/*
ResourcesHash It is used to pass to the Chord implementation
*/
type ResourcesHash struct {
	sizeBytes        int
	hash             []byte
	hostname         string
	ignoreChordWrite bool
}

func NewResourcesHash(bytesSize int, hostname string) *ResourcesHash {
	hash := &ResourcesHash{}
	hash.hash = make([]byte, bytesSize, bytesSize)
	hash.sizeBytes = bytesSize
	hash.hostname = hostname
	hash.ignoreChordWrite = false
	return hash
}

// ============================== Hash Interface ================================

func (rh *ResourcesHash) Write(p []byte) (n int, err error) {
	if !rh.ignoreChordWrite {
		pString := string(p)
		if pString == rh.hostname { // Generate a random GUID id for a joining node
			rh.generateRandomHash(p)
			log.Debugf("[Hash] Hostname: %s, Random GUID: %v", rh.hostname, p)
			for index, value := range p {
				rh.hash[index] = value
			}
			rh.ignoreChordWrite = true
			return 0, nil
		} else { // Passing a GUID id that I have already randomly generated (depending on resources)
			for index, value := range p {
				rh.hash[index] = value
			}
			return 0, nil
		}
	}
	rh.ignoreChordWrite = false
	return 0, nil
}

func (rh *ResourcesHash) Sum(b []byte) []byte {
	// Only return the hash we wrote
	return rh.hash
}

func (rh *ResourcesHash) Reset() {
	for index := range rh.hash {
		rh.hash[index] = 0
	}
}

func (rh *ResourcesHash) Size() int {
	return rh.sizeBytes
}

func (rh *ResourcesHash) BlockSize() int {
	return 0
}

func (rh *ResourcesHash) generateRandomHash(hashToFill []byte) {
	randomGeneratorMutex.Lock()
	defer randomGeneratorMutex.Unlock()
	randomGenerator.Read(hashToFill)
}
