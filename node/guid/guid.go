package guid

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

/*
Guid Represents the global unique identifier of each node
*/
type Guid struct {
	id *big.Int
}

var randomSource rand.Source = rand.NewSource(time.Now().Unix())	// Random source to generate random Guids

var isGuidInitialized bool = false;	// Used to allow only one initialization of the Guid size

var guidSizeBits int = 160 			// 160-bits default (To maintain compatibility with chord implementation)


func InitializeGuid(guidBitsSize int) {
	if !isGuidInitialized {
		guidSizeBits = guidBitsSize
		isGuidInitialized = true
	}
}

func GuidSizeBits() int {
	return guidSizeBits
}

func GuidSizeBytes() int {
	return guidSizeBits / 8
}


func GetMaximumGuid() *Guid {
	maxId := big.NewInt(0)
	maxId.Exp(big.NewInt(2), big.NewInt(int64(guidSizeBits)), nil)
	maxId = maxId.Sub(maxId, big.NewInt(1))
	return newGuidBigInt(maxId)
}

func NewGuidRandom() *Guid {
	guid := &Guid{}

	guid.id = big.NewInt(0)
	guid.id.Rand(rand.New(randomSource), GetMaximumGuid().id)

	return guid
}

func NewGuidString(stringId string) *Guid {
	guid := &Guid{}

	guid.id = big.NewInt(0)
	guid.id.SetString(stringId, 10)

	return guid
}

func NewGuidInteger(intId int64) *Guid {
	guid := &Guid{}

	guid.id = big.NewInt(0)
	guid.id.SetInt64(intId)

	return guid
}

func NewGuidBytes(bytesId []byte) *Guid {
	guid := &Guid{}

	guid.id = big.NewInt(0)
	guid.id.SetBytes(bytesId)
	return guid
}

func newGuidBigInt(intId *big.Int) *Guid {
	guid := &Guid{}

	guid.id = big.NewInt(0)
	guid.id.Set(intId)
	return guid
}

func (guid *Guid) GenerateRandomBetween(nextGuid Guid) (*Guid, error) {
	dif := big.NewInt(0)
	randOffset := big.NewInt(0)
	res := big.NewInt(0)

	dif.Sub(nextGuid.id, guid.id)

	randOffset.Rand(rand.New(randomSource), dif)

	res.Add(guid.id, randOffset)

	return NewGuidString(res.String()), nil
}

/*
Returns the number of ids (as a string with an integer in base 10) using % offset to higher Guid
*/
func (guid *Guid) Partitionate(offsetPercentage int, nextGuid Guid) string {
	offset := big.NewInt(int64(offsetPercentage))
	dif := big.NewInt(0)
	dif.Sub(nextGuid.id, guid.id)

	offset.Mul(offset, dif)
	offset.Div(offset, big.NewInt(100))
	return offset.String()
}

/*
Adds an offset of ids to the Guid
*/
func (guid *Guid) AddOffset(offset string) {
	toAdd := big.NewInt(0)
	toAdd.SetString(offset, 10)

	guid.id.Add(guid.id, toAdd)
}

/*
Compare to see what is the biggest Guid
*/
func (guid *Guid) Cmp(guid2 Guid) int {
	return guid.id.Cmp(guid2.id)
}

/*
Returns an array of bytes (with size of guidSizeBits) with the value of the Guid
*/
func (guid *Guid) Bytes() []byte {
	numOfBytes := guidSizeBits / 8
	res := make([]byte, numOfBytes)
	idBytes := guid.id.Bytes()
	index := 0
	for ; index < numOfBytes-cap(idBytes); index++ {
		res[index] = 0
	}
	for k := 0; index < numOfBytes; k++ {
		res[index] = idBytes[k]
		index++
	}
	return res
}

/*
Creates a copy of the Guid
*/
func (guid *Guid) Copy() *Guid {
	return NewGuidString(guid.String())
}

/*
Returns the value of the id in a string as an integer in base 10
*/
func (guid *Guid) String() string {
	return guid.id.String()
}

func (guid *Guid) PrintDecimal() {
	fmt.Println(guid.id)
}
