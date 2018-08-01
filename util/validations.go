package util

// IsValidPort verifies if the port is valid or not.
func IsValidPort(port int) bool {
	if port < 0 || port > 65553 {
		return false
	} else {
		return true
	}
}
