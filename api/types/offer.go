package types

import "github.com/pkg/errors"

type Resources struct {
	CPUPower CPUPower `json:"CPUPower"`
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

// ======================= CPU Power ========================

type CPUPower uint

const (
	LowCPUPower CPUPower = iota
	MediumCPUPower
	HighCPUPower
)

var cpuPowers = []string{"low", "medium", "high"}

func (cp CPUPower) name() string {
	return cpuPowers[cp]
}

func (cp CPUPower) ordinal() int {
	return int(cp)
}

func (cp CPUPower) String() string {
	return cpuPowers[cp]
}

func (cp CPUPower) values() *[]string {
	return &cpuPowers
}

func (cp *CPUPower) ValueOf(arg string) error {
	for i, name := range cpuPowers {
		if name == arg {
			*cp = CPUPower(i)
			return nil
		}
	}
	return errors.New("invalid enum value")
}
