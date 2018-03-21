package client

import ()

const UNKNOWN = 1
const TIMEOUT = 2

type ClientError struct {
	Code int
}

func NewClientError(code int) *ClientError {
	res := &ClientError{}
	res.Code = code
	return res
}

func (ce *ClientError) Error() string {
	switch ce.Code {
	case TIMEOUT:
		return "Request Timeout!!"
	default:
		return "Unknown error. THIS SHOULDN'T HAPPEN"
	}
}
