package error_code

//go:generate stringer -type=ErrorCodeBase

type ErrorCodeBase int

func (c ErrorCodeBase) GetCode() string {
	return c.String()
}
func (c ErrorCodeBase) GetMessage() string {
	return _ErrorCodeBase[c]
}

const (
	InternalError ErrorCodeBase = iota
	InvalidParameter
	BadRequest
)

var _ErrorCodeBase = map[ErrorCodeBase]string{
	InternalError:    "internal error",
	InvalidParameter: "invalid parameter",
	BadRequest:       "bad request",
}
