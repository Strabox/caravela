package docker

const Running = 0
const Finished = 1
const Unknown = 2

/*
Simple execution status of a docker container.
*/
type ContainerStatus struct {
	statusCode int
}

func NewContainerStatus(statusCode int) ContainerStatus {
	return ContainerStatus{statusCode: statusCode}
}

func (cs ContainerStatus) IsRunning() bool {
	return cs.statusCode == Running
}
