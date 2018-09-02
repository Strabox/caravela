package swarm

type weightedNode struct {
	*node
	weight int
}

type weightedNodeList []*weightedNode

func (n weightedNodeList) Len() int {
	return len(n)
}

func (n weightedNodeList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n weightedNodeList) Less(i, j int) bool {
	var (
		ip = n[i]
		jp = n[j]
	)

	// If the nodes have the same weight sort them out by number of containers.
	if ip.weight == jp.weight {
		return ip.totalContainersRunning() < jp.totalContainersRunning()
	}
	return ip.weight < jp.weight
}
