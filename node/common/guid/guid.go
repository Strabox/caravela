package guid

import (
	"github.com/strabox/caravela/util"
	"math/big"
	"math/rand"
	"time"
)

const guidShortStringSize = 12

// randomGenerator random source to generate random GUIDs
var randomGenerator = rand.New(util.NewSourceSafe(rand.NewSource(time.Now().Unix())))

// Used to allow only one initialization of the GUID module
var isGUIDInitialized = false

// 160-bits default (To maintain compatibility with used chord overlay implementation)
var guidSizeBits = 160

// GUID represents a Global Unique Identifier (GUID) for a system's node
type GUID struct {
	id *big.Int
}

// Init initializes the GUID package with the size of the GUID.
func Init(guidBitsSize int) {
	if !isGUIDInitialized {
		guidSizeBits = guidBitsSize
		isGUIDInitialized = true
	}
}

// SizeBits returns the size of the GUID (in bits).
func SizeBits() int {
	return guidSizeBits
}

// SizeBytes returns the size of the GUID (in bytes).
func SizeBytes() int {
	return guidSizeBits / 8
}

// MaximumGUID creates the maximum available for the current defined number of bits.
func MaximumGUID() *GUID {
	maxId := big.NewInt(0)
	maxId.Exp(big.NewInt(2), big.NewInt(int64(guidSizeBits)), nil)
	maxId = maxId.Sub(maxId, big.NewInt(1))
	return newGUIDBigInt(maxId)
}

// NewZero creates the 0 GUID.
func NewZero() *GUID {
	return &GUID{
		id: big.NewInt(0),
	}
}

// NewGUIDRandom creates a random GUID in the range [0,MaxGUID).
func NewGUIDRandom() *GUID {
	guid := &GUID{}

	guid.id = big.NewInt(0)
	guid.id.Rand(randomGenerator, MaximumGUID().id)

	return guid
}

// NewGUIDString creates a new GUID based on a string representation (in base 10) of the identifier.
func NewGUIDString(stringID string) *GUID {
	guid := &GUID{}

	guid.id = big.NewInt(0)
	guid.id.SetString(stringID, 10)

	return guid
}

// NewGUIDInteger creates a new GUID based on an integer64 representation of the identifier.
func NewGUIDInteger(intId int64) *GUID {
	guid := &GUID{}

	guid.id = big.NewInt(0)
	guid.id.SetInt64(intId)

	return guid
}

// NewGUIDBytes creates a new GUID based on an array of bytes representation of the identifier.
// Array of bytes is a representation of the number using the minimum number of bits.
func NewGUIDBytes(bytesID []byte) *GUID {
	guid := &GUID{}

	guid.id = big.NewInt(0)
	guid.id.SetBytes(bytesID)
	return guid
}

// newGUIDBigInt creates a new GUID based on Golang big.Int representation.
func newGUIDBigInt(bytesID *big.Int) *GUID {
	guid := &GUID{}

	guid.id = big.NewInt(0)
	guid.id.Set(bytesID)
	return guid
}

func scaleWindow() *big.Int {
	maxGUID := MaximumGUID()

	tempBigInt := big.NewInt(0)
	tempBigInt.Div(maxGUID.id, big.NewInt(10000))
	tempBigInt.Mul(tempBigInt, big.NewInt(5))

	return tempBigInt
}

// GenerateInnerRandomGUIDV2 ...
func (g *GUID) GenerateInnerRandomGUIDV2(topGUID GUID) (*GUID, error) {
	dif := big.NewInt(0)

	dif.Sub(topGUID.id, g.id)

	dif.Div(dif, scaleWindow())

	dif.Rand(randomGenerator, dif)

	dif.Mul(dif, scaleWindow())

	dif.Add(g.id, dif)

	return NewGUIDString(dif.String()), nil
}

// GenerateInnerRandomGUID generates a random GUID that belongs to the interval [this, topGUID).
func (g *GUID) GenerateInnerRandomGUID(topGUID GUID) (*GUID, error) {
	dif := big.NewInt(0)
	randOffset := big.NewInt(0)
	res := big.NewInt(0)

	dif.Sub(topGUID.id, g.id)

	randOffset.Rand(randomGenerator, dif)

	res.Add(g.id, randOffset)

	return NewGUIDString(res.String()), nil
}

// PercentageOffset returns the number of ids (as a string with an integer in base 10) using % offset to higher GUID.
func (g *GUID) PercentageOffset(offsetPercentage int, nextGuid GUID) string {
	offset := big.NewInt(int64(offsetPercentage))
	dif := big.NewInt(0)
	dif.Sub(nextGuid.id, g.id) // Dif between nextGuid and receiver.

	offset.Mul(offset, dif)
	offset.Div(offset, big.NewInt(100))
	return offset.String()
}

// AddOffset adds an offset (as a string in base 10) of ids to the GUID.
func (g *GUID) AddOffset(offset string) {
	toAdd := big.NewInt(0)
	toAdd.SetString(offset, 10)

	g.id.Add(g.id, toAdd)
}

// Cmp used to check what if the guid is higher, lower or equal than the given guid.
func (g *GUID) Cmp(guid2 GUID) int {
	return g.id.Cmp(guid2.id)
}

// Higher returns true if guid is higher than the given guid and false otherwise.
func (g *GUID) Higher(guid2 GUID) bool {
	return g.id.Cmp(guid2.id) > 0
}

// Greater returns true if guid is lower than the given guid and false otherwise.
func (g *GUID) Lower(guid2 GUID) bool {
	return g.id.Cmp(guid2.id) < 0
}

// Compare if two GUIDs are equal or not.
func (g *GUID) Equals(guid2 GUID) bool {
	return g.id.Cmp(guid2.id) == 0
}

// Bytes returns an array of bytes (with size of guidSizeBits) with the value of the GUID.
func (g *GUID) Bytes() []byte {
	numOfBytes := guidSizeBits / 8
	res := make([]byte, numOfBytes)
	idBytes := g.id.Bytes()
	index := 0
	for ; index < numOfBytes-cap(idBytes); index++ { // Padding the higher bytes with 0s.
		res[index] = 0
	}
	for k := 0; index < numOfBytes; k++ {
		res[index] = idBytes[k]
		index++
	}
	return res
}

// Int64 returns an int64 that represents the GUID.
func (g *GUID) Int64() int64 {
	return g.id.Int64()
}

// Copy creates a copy of the GUID object.
func (g *GUID) Copy() *GUID {
	return NewGUIDString(g.String())
}

// String returns the value of the GUID in a string representation (as an integer in base 10).
func (g *GUID) String() string {
	return g.id.String()
}

// Short returns the first digits of the GUID in a string representation.
func (g *GUID) Short() string {
	return g.id.String()[0:guidShortStringSize]
}
