package rest

import (

)

/*
Decode/Encode from/to Json strings in HTTP body.
Used only to transmit information in HTTP body.
*/
type OfferJSON struct {
	DestGuid	string
	SuppIP		string
	OfferID 	int
	Amount 		int
}

/*

*/
type AckOfferJSON struct {
	Guid	string
	Ip		string
}
