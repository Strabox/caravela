package types

type PartitionState struct {
	PartitionResources Resources `json:"PartitionResources"`
	Hits               int       `json:"Hits"`
}
