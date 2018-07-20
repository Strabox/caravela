package types

type Offer struct {
	ID        int64     `json:"ID"`
	Amount    int       `json:"Amount"`
	Resources Resources `json:"Resources"`
}

type AvailableOffer struct {
	Offer      `json:"Offer"`
	SupplierIP string `json:"SupplierIP"`
}

type Resources struct {
	CPUs int `json:"CPUs"`
	RAM  int `json:"RAM"`
}
