package rest

/*
Offer struct/JSON used in REST API when offering resources to be used by the system
*/
type OfferJSON struct {
	FromSupplierIP   string `json:"FromSupplierIP"`   // IP address of the supplier node responsible for the offer
	FromSupplierGUID string `json:"FromSupplierGUID"` // GUID of the supplier responsible for the offer
	ToTraderGUID     string `json:"ToTraderGUID"`     // GUID of the destination trader
	OfferID          int64  `json:"OfferID"`          // Local ID of the offer (unique inside the supplier)
	Amount           int    `json:"Amount"`           // Amount of slots of this offer
	CPUs             int    `json:"CPUs"`             // Amount of CPUs in the offer
	RAM              int    `json:"RAM"`              // Amount of RAM in the offer
}

/*
Refresh offer struct/JSON used in remote REST API when traders refresh offers
*/
type OfferRefreshJSON struct {
	FromTraderGUID string `json:"FromTraderGUID"` // Trader's GUID that is refreshing the offer
	OfferID        int64  `json:"OfferID"`        // Offer's ID in the supplier
}

/*
Remove offer struct/JSON used in remote REST API when a supplier remove its offer from a trader
*/
type OfferRemoveJSON struct {
	FromSupplierIP   string `json:"FromSupplierIP"`   // IP address of the supplier node responsible for the offer
	FromSupplierGUID string `json:"FromSupplierGUID"` // GUID of the supplier responsible for the offer
	ToTraderGUID     string `json:"ToTraderGUID"`     // GUID of the destination trader
	OfferID          int64  `json:"OfferID"`          // Local ID of the offer (unique inside the supplier)
}

/*
Run container struct/JSON used in local REST API when a user submit a container o run
*/
type RunContainerJSON struct {
	ContainerImage string   `json:"ContainerImage"` // Container's image key
	Arguments      []string `json:"Arguments"`      // Arguments for container run
	CPUs           int      `json:"CPUs"`           // Amount of CPUs necessary
	RAM            int      `json:"RAM"`            // Amount of RAM necessary
}
