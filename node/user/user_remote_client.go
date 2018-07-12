package user

// Interface that provides the necessary methods to talk with other nodes.
type userRemoteClient interface {
	StopLocalContainer(toSupplierIP string, containerID string) error
}
