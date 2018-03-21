package rest

import ()

/*
Offer struct/JSON used in REST API when offering resources to be used by the system
*/
type OfferJSON struct {
	DestGuid string `json:"DestGuid"` // GUID of the destination trader
	SuppIP   string `json:"SuppIP"`   // IP address of the supplier node responsible for the offer
	SuppGUID string `json:"SuppGUID"` // GUID of the supplier responsible for the offer
	OfferID  int    `json:"OfferID"`  // Local ID of the offer (unique inside the supplier)
	Amount   int    `json:"Amount"`   // Amount of slots of this offer
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
	code    int    `json:"code"`
	message string `json:"message"`
}
