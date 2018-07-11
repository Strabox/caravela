package remote

import (
	"strings"
)

const Unknown = 1
const CaravelaInstanceUnavailable = 2

// Error of CARAVELA's remote client.
type Error struct {
	Code int
	err  error
}

func NewRemoteClientError(err error) *Error {
	res := &Error{
		err: err,
	}
	if strings.Contains(err.Error(), "No connection") {
		res.Code = CaravelaInstanceUnavailable
	} else {
		res.Code = Unknown
	}
	return res
}

func (ce *Error) Error() string {
	switch ce.Code {
	case CaravelaInstanceUnavailable:
		return "Instance unavailable"
	default:
		return ce.err.Error()
	}
}
