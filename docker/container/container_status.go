package container

const Running = 0
const Finished = 1
const Unknown = 2

// Simple execution status of a docker container.
type Status struct {
	statusCode int
}

func NewContainerStatus(statusCode int) Status {
	return Status{statusCode: statusCode}
}

func (cs Status) IsRunning() bool {
	return cs.statusCode == Running
}
