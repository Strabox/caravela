package chord

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/util"
	"math/rand"
	"sync"
	"time"
)

// Used to generate random hashes => GUID for the joining nodes.
var randomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
var randomGeneratorMutex = sync.Mutex{}

// It is used to pass to the Chord implementation and it does a pass-through hash
// because we generate our hash in the node layer depending on the resources for a request.
type ResourcesHash struct {
	// Hash size (in bytes).
	sizeBytes int
	// Current hash value. In our case this is the value generated in the node layer.
	hash []byte
	//(Hack) Hostname necessary to intercept hash of hostname strings when a join join request is issued in the local node.
	hostname string
	// (Hack) Used to ignore the second call to write after a call issued to generate the hash/GUID of a joining node ????
	ignoreChordWrite bool
}

// NewResourcesHash creates a new hash.
func NewResourcesHash(bytesSize int, hostname string) *ResourcesHash {
	hash := &ResourcesHash{}
	hash.hash = make([]byte, bytesSize, bytesSize)
	hash.sizeBytes = bytesSize
	hash.hostname = hostname
	hash.ignoreChordWrite = false
	return hash
}

// Generate a random hash. Used to generate random GUIDs for the joining nodes.
func (r *ResourcesHash) generateRandomHash(hashToFill []byte) {
	randomGeneratorMutex.Lock()
	defer randomGeneratorMutex.Unlock()
	randomGenerator.Read(hashToFill)
}

// ============================== Hash Interface ================================

func (r *ResourcesHash) Write(p []byte) (n int, err error) {
	if !r.ignoreChordWrite {
		pString := string(p)
		if pString == r.hostname { // Generate a random GUID id for a joining node
			r.generateRandomHash(p)
			log.Debugf(util.LogTag("Hash")+"Trader Hash/GUID: %v", p)
			for index, value := range p {
				r.hash[index] = value
			}
			r.ignoreChordWrite = true
			return 0, nil
		} else { // Passing a GUID id that I have already randomly generated in node layer (depending on resources)
			for index, value := range p {
				r.hash[index] = value
			}
			return 0, nil
		}
	}
	r.ignoreChordWrite = false
	return 0, nil
}

func (r *ResourcesHash) Sum(b []byte) []byte {
	// Only return the hash we wrote
	return r.hash
}

func (r *ResourcesHash) Reset() {
	for index := range r.hash {
		r.hash[index] = 0
	}
}

func (r *ResourcesHash) Size() int {
	return r.sizeBytes
}

func (r *ResourcesHash) BlockSize() int {
	return 0
}
