package remote

const UNKNOWN = 1
const TIMEOUT = 2

type Error struct {
	Code int
}

func NewClientError(code int) *Error {
	res := &Error{}
	res.Code = code
	return res
}

func (ce *Error) Error() string {
	switch ce.Code {
	case TIMEOUT:
		return "Request Timeout!!"
	default:
		return "Unknown error. THIS SHOULDN'T HAPPEN"
	}
}
