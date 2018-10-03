package types

type PartitionState struct {
	PartitionResources Resources `json:"PR"`
	Hits               int       `json:"H"`
}
