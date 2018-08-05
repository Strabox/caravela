package containers

import "context"

// Containers API necessary to forward the REST calls
type Containers interface {
	StopLocalContainer(ctx context.Context, containerID string) error
}
