package cli

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"net"
)

// ================== Defaults values for CLI flags and requests ====================

const defaultConfigurationFile = configuration.DefaultFilePath
const defaultLogLevel = "fatal"
const defaultCaravelaInstanceIP = "127.0.0.1" // Target the local's node,
const defaultHostIP = ""

const defaultContainerName = ""
const defaultCPUClass = types.LowCPUClassStr
const defaultCPUs = 0
const defaultMemory = 0
const defaultContainerGroupPolicy = types.SpreadGroupPolicyStr

var defaultContainerArgs = make([]string, 0)
var defaultPortMappingsArgs = make([]string, 0)

// getOutboundIP get preferred outbound IP of this machine.
// Returns the machine's IP that can connects to the internet.
func getOutboundIP() string {
	const googleDNSAddress = "8.8.8.8:80"
	const transportProtocol = "udp"

	conn, err := net.Dial(transportProtocol, googleDNSAddress)
	if err != nil {
		fatalPrintln("Please turn on the network")
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
