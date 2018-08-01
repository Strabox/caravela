package client

import "strings"

// CARAVELA's client error codes.
const (
	// UnknownError deals with unexpected errors.
	UnknownError = iota
	// CaravelaInstanceUnavailableError is obtained when the client tries to send a request to a CARAVELA's
	// instance that is not working e.g. because it is turned off.
	CaravelaInstanceUnavailableError
)

// Error represents the errors returned by the CARAVELA's client.
// It implements the error interface.
type Error struct {
	err error
	// Error's code.
	Code int
}

// newClientError creates a new CARAVELA's client error based on internal errors.
func newClientError(err error) *Error {
	res := &Error{
		err: err,
	}
	if strings.Contains(err.Error(), "No connection") {
		res.Code = CaravelaInstanceUnavailableError
	} else {
		res.Code = UnknownError
	}
	return res
}

func (ce *Error) Error() string {
	switch ce.Code {
	case CaravelaInstanceUnavailableError:
		return "Caravela instance unavailable"
	case UnknownError:
		return "Unknown error"
	default:
		return ce.err.Error()
	}
}
