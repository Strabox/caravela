package util

import (
	"strconv"
	"strings"
)

func ObtainIpPort(hostname string) (string, int) {
	nodeIP := strings.Split(hostname, ":")[0]
	nodePort, _ := strconv.Atoi(strings.Split(hostname, ":")[1])
	return nodeIP, nodePort
}
