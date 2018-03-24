package rest

/*
Offer struct/JSON used in REST API when offering resources to be used by the system
*/
type OfferJSON struct {
	FromSupplierIP   string `json:"FromSupplierIP"`   // IP address of the supplier node responsible for the offer
	FromSupplierGUID string `json:"FromSupplierGUID"` // GUID of the supplier responsible for the offer
	ToTraderGUID     string `json:"ToTraderGUID"`     // GUID of the destination trader
	OfferID          int    `json:"OfferID"`          // Local ID of the offer (unique inside the supplier)
	Amount           int    `json:"Amount"`           // Amount of slots of this offer
	CPUs             int    `json:"CPUs"`             // Amount of CPUs in the offer
	RAM              int    `json:"RAM"`              // Amount of RAM in the offer
}

/*
Refresh offer struct/JSON used in REST API
*/
type OfferRefreshJSON struct {
	FromTraderGUID string `json:"FromTraderGUID"` // Trader's GUID that is refreshing the offer
	OfferID        int    `json:"OfferID"`        // Offer's ID in the supplier
}

/*
Error struct/JSON used in REST API
*/
type ErrorJSON struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
