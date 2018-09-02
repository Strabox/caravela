package types

import "github.com/pkg/errors"

type Resources struct {
	CPUClass CPUClass `json:"CPUClass"`
	CPUs     int      `json:"CPUs"`
	RAM      int      `json:"RAM"`
}

type Offer struct {
	ID        int64     `json:"ID"`
	Amount    int       `json:"Amount"`
	Resources Resources `json:"Resources"`
}

type AvailableOffer struct {
	Offer      `json:"Offer"`
	SupplierIP string `json:"SupplierIP"`
}

type AvailableOffers []AvailableOffer

func (ao AvailableOffers) Len() int {
	return len(ao)
}
func (ao AvailableOffers) Swap(i, j int) {
	ao[i], ao[j] = ao[j], ao[i]
}
func (ao AvailableOffers) Less(i, j int) bool {
	if ao[i].Resources.CPUs < ao[j].Resources.CPUs {
		return true
	} else if ao[i].Resources.CPUs == ao[j].Resources.CPUs {
		if ao[i].Resources.RAM <= ao[j].Resources.RAM {
			return true
		}
	}
	return false
}

// ======================= CPU Class ========================

type CPUClass uint

const (
	LowCPUPClass CPUClass = iota
	HighCPUClass
)

var cpuClasses = []string{"low", "high"}

func (cp CPUClass) name() string {
	return cpuClasses[cp]
}

func (cp CPUClass) ordinal() int {
	return int(cp)
}

func (cp CPUClass) String() string {
	return cpuClasses[cp]
}

func (cp CPUClass) values() *[]string {
	return &cpuClasses
}

func (cp *CPUClass) ValueOf(arg string) error {
	for i, name := range cpuClasses {
		if name == arg {
			*cp = CPUClass(i)
			return nil
		}
	}
	return errors.New("invalid enum value")
}
