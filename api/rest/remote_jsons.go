package rest

/* =================================================================================
									Request Messages
   ================================================================================= */

/*
Create offer struct/JSON used in REST APIs when a supplier offer resources to be used by others
*/
type CreateOfferJSON struct {
	FromSupplierIP   string `json:"FromSupplierIP"`   // IP address of the supplier node responsible for the offer
	FromSupplierGUID string `json:"FromSupplierGUID"` // GUID of the supplier responsible for the offer
	ToTraderGUID     string `json:"ToTraderGUID"`     // GUID of the destination trader
	OfferID          int64  `json:"OfferID"`          // Local ID of the offer (unique inside the supplier)
	Amount           int    `json:"Amount"`           // Amount of quantity slots of this offer
	CPUs             int    `json:"CPUs"`             // Amount of CPUs in the offer
	RAM              int    `json:"RAM"`              // Amount of RAM in the offer
}

/*
Refresh offer struct/JSON used in remote REST APIs when a trader refresh an offer
*/
type RefreshOfferJSON struct {
	FromTraderGUID string `json:"FromTraderGUID"` // Trader's GUID that is refreshing the offer
	OfferID        int64  `json:"OfferID"`        // Offer's ID in the supplier
}

/*
Remove offer struct/JSON used in remote REST APIs when a supplier remove its offer from a trader
*/
type OfferRemoveJSON struct {
	FromSupplierIP   string `json:"FromSupplierIP"`   // IP address of the supplier node responsible for the offer
	FromSupplierGUID string `json:"FromSupplierGUID"` // GUID of the supplier responsible for the offer
	ToTraderGUID     string `json:"ToTraderGUID"`     // GUID of the destination trader
	OfferID          int64  `json:"OfferID"`          // Local ID of the offer (unique inside the supplier)
}

/*
Get offers struct/JSON used in the REST APIs
*/
type GetOffersJSON struct {
	ToTraderGUID string `json:"ToTraderGUID"` // GUID of the destination trader
}

/*
Launch container struct/JSON used in the REST APIs
*/
type LaunchContainerJSON struct {
	FromBuyerIP       string   `json:"FromBuyerIP"`
	OfferID           int64    `json:"OfferID"`
	ContainerImageKey string   `json:"ContainerImageKey"`
	ContainerArgs     []string `json:"ContainerArgs"`
	CPUs              int      `json:"CPUs"`
	RAM               int      `json:"RAM"`
}

/* =================================================================================
									Response Messages
   ================================================================================= */

/*
Offer struct/JSON used in the REST APIs
*/
type OfferJSON struct {
	ID         int64  `json:"ID"`         // Local ID of the offer (unique inside the supplier)
	SupplierIP string `json:"SupplierIP"` // IP address of the supplier node responsible for the offer
}

/*
List of offers struct/JSON used in the REST APIs
*/
type OffersListJSON struct {
	Offers []OfferJSON `json:"Offers"` // A list of offers that a trader has
}
