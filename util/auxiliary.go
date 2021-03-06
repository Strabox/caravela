package util

import (
	"strconv"
	"strings"
)

// Obtains IP port from a hostname string i.e. 178.673.111.33:9122.
func ObtainIpPort(hostname string) (string, int) {
	nodeIP := strings.Split(hostname, ":")[0]
	nodePort, _ := strconv.Atoi(strings.Split(hostname, ":")[1])
	return nodeIP, nodePort
}
