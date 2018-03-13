package rest

import ()

/*
Offer struct/JSON used in REST API
*/
type OfferJSON struct {
	DestGuid string `json:"DestGuid"`
	SuppIP   string `json:"SuppIP"`
	OfferID  int    `json:"OfferID"`
	Amount   int    `json:"Amount"`
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
