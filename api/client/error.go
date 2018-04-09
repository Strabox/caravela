package client

import "strings"

const Unknown = 1
const CaravelaInstanceUnavailable = 2

type Error struct {
	Code int
}

func NewClientError(err error) *Error {
	res := &Error{}
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
		return "Caravela instance unavailable"
	default:
		return "Unknown error"
	}
}
