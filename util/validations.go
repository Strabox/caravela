package util

// Function to validate a port
func IsValidPort(port int) bool {
	if port < 0 || port > 65553 {
		return false
	} else {
		return true
	}
}
