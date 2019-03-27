package error_code

type ErrorCode interface {
	GetCode() string
	GetMessage() string
}
