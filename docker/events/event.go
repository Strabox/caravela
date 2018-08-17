package events

const ContainerDied = "die"

type Event struct {
	Type  string
	Value string
}
