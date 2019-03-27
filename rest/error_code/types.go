package error_code

import (
	"encoding/json"
	"github.com/weishi258/http-log-collector/rest/model"
)

type ResponseError interface {
	GetCode() string
	GetMessage() string
	SetCode(code string)
	SetMessage(message string)
	MarshalJsonByte() []byte
}
type responseError struct {
	code    string
	message string
}

func (c responseError) GetCode() string {
	return c.code
}
func (c responseError) GetMessage() string {
	return c.message
}
func (c responseError) SetCode(code string) {
	c.code = code
}
func (c responseError) SetMessage(message string) {
	c.message = message
}

func (c responseError) MarshalJsonByte() []byte {
	// ignore error here
	ret, _ := json.Marshal(&model.Error{Code: c.code, Message: c.message})
	return ret
}
func NewResponseErrorCustom(code string, message string) ResponseError {
	return responseError{code, message}
}

func NewResponseError(errorCode ErrorCode) ResponseError {
	return responseError{errorCode.GetCode(), errorCode.GetMessage()}
}
