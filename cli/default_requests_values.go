package cli

import (
	"github.com/strabox/caravela/configuration"
	"net"
)

// ================== Defaults values for CLI flags and requests ====================

const defaultConfigurationFile = configuration.DefaultFilePath
const defaultLogLevel = "fatal"
const defaultCaravelaInstanceIP = "127.0.0.1"
const defaultHostIP = ""

const defaultContainerName = ""
const defaultCPUPower = "low"
const defaultCPUs = 0
const defaultRAM = 0
const defaultContainerGroupPolicy = "spread"

var defaultContainerArgs = make([]string, 0)
var defaultPortMappingsArgs = make([]string, 0)

// getOutboundIP get preferred outbound IP of this machine.
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
