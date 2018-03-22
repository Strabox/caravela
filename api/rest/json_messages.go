package rest

/*
Offer struct/JSON used in REST API when offering resources to be used by the system
*/
type OfferJSON struct {
	TraderDestGUID string `json:"TraderDestGUID"` // GUID of the destination trader
	SupplierIP     string `json:"SupplierIP"`     // IP address of the supplier node responsible for the offer
	SupplierGUID   string `json:"SupplierGUID"`   // GUID of the supplier responsible for the offer
	OfferID        int    `json:"OfferID"`        // Local ID of the offer (unique inside the supplier)
	Amount         int    `json:"Amount"`         // Amount of slots of this offer
}

/*
AckOffer struct/JSON used in REST API
*/
type AckOfferJSON struct {
	Guid string `json:"Guid"`
	IP   string `json:"IP"`
}

/*
Error struct/JSON used in REST API
*/
type ErrorJSON struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
