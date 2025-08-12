package types

import "fmt"

//Standard JSON-RPC error codes
const (
	//-32700
	ERROR_CODE_PARSE_ERROR = -32700
	//-32600
	ERROR_CODE_INVALID_REQUEST = -32600
	//-32601
	ERROR_CODE_METHOD_NOT_FOUND = -32601
	//-32602
	ERROR_CODE_INVALID_PARAMS = -32602
	//-32603
	ERROR_CODE_INTERNAL_ERROR = -32603
)

//SDK error codes
const (
	//-32000
	ERROR_CODE_CONNECTION_CLOSED = -32000
	//-32001
	ERROR_CODE_REQUEST_TIMEOUT = -32001
	//-32002
	ERROR_CODE_SESSION_ID_NOT_FOUND = -32002
	//-32003
	ERROR_CODE_METHOD_NOT_ALLOWED = -32003
)

const (
	ERROR_ERROR_INTERFACE_TYPE = iota + 600
	MCP_ERROR_ERROR_INTERFACE_TYPE
)

type ErrorInterface interface {
	TypeOfError() int
	GetErrorCode() int
	GetErrorMessage() string
	GetErrorData() interface{}
	ToError() error
}

type Error struct {
	//The error type that occurred.
	Code int `json:"code"`
	//A short description of the error. The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"`
	//Additional information about the error. The value of this member is defined by the sender (e.g. detailed error information, nested errors etc.).
	Data interface{} `json:"data,omitempty"`
}

func (e *Error) TypeOfError() int          { return ERROR_ERROR_INTERFACE_TYPE }
func (e *Error) GetErrorCode() int         { return e.Code }
func (e *Error) GetErrorMessage() string   { return e.Message }
func (e *Error) GetErrorData() interface{} { return e.Data }
func (e *Error) ToError() error {
	return fmt.Errorf("code: %d, message: %s, data: %v", e.Code, e.Message, e.Data)
}

type McpError struct {
	Err Error
}

func (e *McpError) TypeOfError() int          { return MCP_ERROR_ERROR_INTERFACE_TYPE }
func (e *McpError) GetErrorCode() int         { return e.Err.Code }
func (e *McpError) GetErrorMessage() string   { return e.Err.Message }
func (e *McpError) GetErrorData() interface{} { return e.Err.Data }
func (e *McpError) ToError() error {
	return fmt.Errorf("code: %d, message: %s, data: %v", e.Err.Code, e.Err.Message, e.Err.Data)
}

func NewMcpError(code int, msg string, data interface{}) *McpError {
	finalMsg := fmt.Sprintf("MCP error %s", msg)
	return &McpError{Error{Code: code, Message: finalMsg, Data: data}}
}
