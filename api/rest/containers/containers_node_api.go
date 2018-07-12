package containers

// Containers API necessary to forward the REST calls
type Containers interface {
	StopLocalContainer(containerID string) error
}
