package types

const (
	JSONRPC_MESSAGE_JSONRPC_REQUEST_TYPE = iota + 80
	JSONRPC_MESSAGE_JSONRPC_NOTIFICATION_TYPE
	JSONRPC_MESSAGE_JSONRPC_BATCH_REQUEST_TYPE
	JSONRPC_MESSAGE_JSONRPC_RESPONSE_TYPE
	JSONRPC_MESSAGE_JSONRPC_ERROR_TYPE
	JSONRPC_MESSAGE_JSONRPC_BATCH_RESPONSE_TYPE
	JSONRPC_MESSAGE_CANCELLED_NOTIFICATION_TYPE
)

const (
	JSONRPC_BATCH_REQUEST_JSONRPC_REQUEST_TYPE = iota + 90
	JSONRPC_BATCH_REQUEST_JSONRPC_NOTIFICATION_TYPE
)

const (
	JSONRPC_BATCH_RESPONSE_JSONRPC_RESPONSE_TYPE = iota + 100
	JSONRPC_BATCH_RESPONSE_JSONRPC_ERROR_TYPE
)

const (
	JSONRPC_GENERAL_RESPONSE_JSONRPC_RESPONSE_TYPE = iota + 200
	JSONRPC_GENERAL_RESPONSE_JSONRPC_ERROR_TYPE
)

const (
	JSONRPC_VERSION = "2.0"
)

//Refers to any valid JSON-RPC object that can be decoded off the wire, or encoded to be sent.
type JSONRPCMessage interface {
	JSONRPCMessageType() int
}

type JSONRPCBatchRequestInterface interface {
	JSONRPCBatchRequestType() int
}

type JSONRPCBatchResponseInterface interface {
	JSONRPCBatchResponseType() int
}

type JSONRPCGeneralResponse interface {
	JSONRPCGeneralResponseType() int
	GetRequestID() RequestID
}

type JSONRPCRequest struct {
	//JSONRPC version, should be "2.0"
	JSONRPC string    `json:"jsonrpc"`
	ID      RequestID `json:"id"`
	Request
}

func (jr *JSONRPCRequest) JSONRPCMessageType() int { return JSONRPC_MESSAGE_JSONRPC_REQUEST_TYPE }
func (jr *JSONRPCRequest) JSONRPCBatchRequestType() int {
	return JSONRPC_BATCH_REQUEST_JSONRPC_REQUEST_TYPE
}
func (jr *JSONRPCRequest) TypeOfRequestInterface() int { return JSONRPC_REQUEST_REQUEST_INTERFACE_TYPE }
func (jr *JSONRPCRequest) GetRequest() Request         { return jr.Request }

type JSONRPCNotification struct {
	//JSONRPC version, should be "2.0"
	JSONRPC      string `json:"jsonrpc"`
	Notification NotificationInterface
}

func (jn *JSONRPCNotification) JSONRPCMessageType() int {
	return JSONRPC_MESSAGE_JSONRPC_NOTIFICATION_TYPE
}
func (jn *JSONRPCNotification) JSONRPCBatchRequestType() int {
	return JSONRPC_BATCH_REQUEST_JSONRPC_NOTIFICATION_TYPE
}

type JSONRPCResponse struct {
	//JSONRPC version, should be "2.0"
	JSONRPC string          `json:"jsonrpc"`
	ID      RequestID       `json:"id"`
	Result  ResultInterface `json:"result"`
}

func (jr *JSONRPCResponse) JSONRPCMessageType() int { return JSONRPC_MESSAGE_JSONRPC_RESPONSE_TYPE }
func (jr *JSONRPCResponse) JSONRPCBatchResponseType() int {
	return JSONRPC_BATCH_RESPONSE_JSONRPC_RESPONSE_TYPE
}
func (jr *JSONRPCResponse) JSONRPCGeneralResponseType() int {
	return JSONRPC_GENERAL_RESPONSE_JSONRPC_RESPONSE_TYPE
}
func (jr *JSONRPCResponse) GetRequestID() RequestID    { return jr.ID }
func (jr *JSONRPCResponse) TypeOfResultInterface() int { return JSONRPC_RESPONSE_RESULT_INTERFACE_TYPE }

type JSONRPCError struct {
	//JSONRPC version, should be "2.0"
	JSONRPC string         `json:"jsonrpc"`
	ID      RequestID      `json:"id"`
	Error   ErrorInterface `json:"error"`
}

func (je *JSONRPCError) JSONRPCMessageType() int { return JSONRPC_MESSAGE_JSONRPC_ERROR_TYPE }
func (je *JSONRPCError) JSONRPCBatchResponseType() int {
	return JSONRPC_BATCH_RESPONSE_JSONRPC_ERROR_TYPE
}
func (je *JSONRPCError) JSONRPCGeneralResponseType() int {
	return JSONRPC_GENERAL_RESPONSE_JSONRPC_ERROR_TYPE
}
func (je *JSONRPCError) GetRequestID() RequestID { return je.ID }

//A JSON-RPC batch request, as described in https://www.jsonrpc.org/specification#batch.
type JSONRPCBatchRequest []JSONRPCBatchRequestInterface

func (jbr *JSONRPCBatchRequest) JSONRPCMessageType() int {
	return JSONRPC_MESSAGE_JSONRPC_BATCH_REQUEST_TYPE
}

//A JSON-RPC batch response, as described in https://www.jsonrpc.org/specification#batch.
type JSONRPCBatchResponse []JSONRPCBatchResponseInterface

func (jbr *JSONRPCBatchResponse) JSONRPCMessageType() int {
	return JSONRPC_MESSAGE_JSONRPC_BATCH_RESPONSE_TYPE
}
