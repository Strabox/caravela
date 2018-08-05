/*
Types package includes all the structures shared between the caravela's server/daemon and its clients.
*/
package types

type Node struct {
	IP   string `json:"IP"`
	GUID string `json:"GUID"`
}
