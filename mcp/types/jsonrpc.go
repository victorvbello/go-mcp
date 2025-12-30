package types

import (
	"encoding/json"
	"fmt"

	"github.com/victorvbello/gomcp/mcp/methods"
)

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

func JSONRPCMessageMarshalJSON(msg JSONRPCMessage) ([]byte, error) {
	var result []byte
	var err error

	jmTyp := msg.JSONRPCMessageType()
	switch jmTyp {
	case JSONRPC_MESSAGE_JSONRPC_REQUEST_TYPE:
		msReq, okType := msg.(*JSONRPCRequest)
		if !okType {
			return nil, fmt.Errorf("invalid type for JSONRPC_MESSAGE_JSONRPC_REQUEST_TYPE")
		}
		result, err = json.Marshal(msReq)
		if err != nil {
			return nil, fmt.Errorf("json.Marshal JSONRPC_MESSAGE_JSONRPC_REQUEST_TYPE, %v", err)
		}
	case JSONRPC_MESSAGE_JSONRPC_NOTIFICATION_TYPE:
		msN, okType := msg.(*JSONRPCNotification)
		if !okType {
			return nil, fmt.Errorf("invalid type for JSONRPC_MESSAGE_JSONRPC_NOTIFICATION_TYPE")
		}
		result, err = json.Marshal(msN)
		if err != nil {
			return nil, fmt.Errorf("json.Marshal JSONRPC_MESSAGE_JSONRPC_NOTIFICATION_TYPE, %v", err)
		}
	case JSONRPC_MESSAGE_JSONRPC_BATCH_REQUEST_TYPE:
		break
	case JSONRPC_MESSAGE_JSONRPC_RESPONSE_TYPE:
		msRT, okType := msg.(*JSONRPCResponse)
		if !okType {
			return nil, fmt.Errorf("invalid type for JSONRPC_MESSAGE_JSONRPC_RESPONSE_TYPE")
		}
		result, err = json.Marshal(msRT)
		if err != nil {
			return nil, fmt.Errorf("json.Marshal JSONRPC_MESSAGE_JSONRPC_RESPONSE_TYPE, %v", err)
		}
	case JSONRPC_MESSAGE_JSONRPC_ERROR_TYPE:
		msErr, okType := msg.(*JSONRPCError)
		if !okType {
			return nil, fmt.Errorf("invalid type for JSONRPC_MESSAGE_JSONRPC_ERROR_TYPE")
		}
		result, err = json.Marshal(msErr)
		if err != nil {
			return nil, fmt.Errorf("json.Marshal JSONRPC_MESSAGE_JSONRPC_ERROR_TYPE, %v", err)
		}
	case JSONRPC_MESSAGE_JSONRPC_BATCH_RESPONSE_TYPE:
		break
	case JSONRPC_MESSAGE_CANCELLED_NOTIFICATION_TYPE:
		msCn, okType := msg.(*CancelledNotification)
		if !okType {
			return nil, fmt.Errorf("invalid type for JSONRPC_MESSAGE_CANCELLED_NOTIFICATION_TYPE")
		}
		result, err = json.Marshal(msCn)
		if err != nil {
			return nil, fmt.Errorf("json.Marshal JSONRPC_MESSAGE_CANCELLED_NOTIFICATION_TYPE, %v", err)
		}
	default:
		return nil, fmt.Errorf("JSONRPCMessage invalid type for %d", jmTyp)
	}
	return result, nil
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
	RequestInterface
}

func (jr *JSONRPCRequest) JSONRPCMessageType() int { return JSONRPC_MESSAGE_JSONRPC_REQUEST_TYPE }
func (jr *JSONRPCRequest) JSONRPCBatchRequestType() int {
	return JSONRPC_BATCH_REQUEST_JSONRPC_REQUEST_TYPE
}

func (jr *JSONRPCRequest) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		JSONRPC string    `json:"jsonrpc"`
		ID      RequestID `json:"id"`
	}{
		JSONRPC: jr.JSONRPC,
		ID:      jr.ID,
	}
	knownFields, err := json.Marshal(&aux)
	if err != nil {
		return nil, fmt.Errorf("marshal known fields: %w", err)
	}
	//Marshal knownFields to map
	baseMap := make(map[string]interface{})
	if err := json.Unmarshal(knownFields, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal known fields to map: %w", err)
	}
	//Marshal RequestInterface
	var reqInB []byte
	switch jr.RequestInterface.TypeOfRequestInterface() {
	case PING_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*PingRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal PING fields: %w", err)
		}
	case INITIALIZE_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*InitializeRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_REQUEST_INITIALIZE fields: %w", err)
		}
	case CREATE_MESSAGE_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*CreateMessageRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_SAMPLING_CREATE_MESSAGE fields: %w", err)
		}
	case LIST_ROOTS_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*ListRootsRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_LIST_ROOTS fields: %w", err)
		}
	case SET_LEVEL_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*SetLevelRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_REQUEST_SET_LEVEL_LOGGING fields: %w", err)
		}
	case GET_PROMPT_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*GetPromptRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_REQUEST_GET_PROMPTS fields: %w", err)
		}
	case LIST_PROMPTS_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*ListPromptsRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_REQUEST_LIST_PROMPTS fields: %w", err)
		}
	case LIST_RESOURCES_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*ListResourcesRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_REQUEST_LIST_RESOURCES fields: %w", err)
		}
	case LIST_RESOURCE_TEMPLATES_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*ListResourceTemplatesRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_REQUEST_TEMPLATES_LIST_RESOURCES fields: %w", err)
		}
	case READ_RESOURCE_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*ReadResourceRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_REQUEST_READ_RESOURCES fields: %w", err)
		}
	case CALL_TOOL_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*CallToolRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_REQUEST_CALL_TOOLS fields: %w", err)
		}
	case LIST_TOOLS_REQUEST_REQUEST_INTERFACE_TYPE:
		reqInB, err = json.Marshal(jr.RequestInterface.(*ListToolsRequest))
		if err != nil {
			return nil, fmt.Errorf("marshal METHOD_REQUEST_LIST_TOOLS fields: %w", err)
		}
	}
	if err := json.Unmarshal(reqInB, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields for:%d err:%w", jr.RequestInterface.TypeOfRequestInterface(), err)
	}
	return json.Marshal(baseMap)
}
func (jr *JSONRPCRequest) UnmarshalJSON(data []byte) error {
	var meta struct {
		JSONRPC string    `json:"jsonrpc"`
		ID      RequestID `json:"id"`
		Method  string    `json:"method"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("error unmarshaling global meta: %v", err)
	}
	jr.JSONRPC = meta.JSONRPC
	jr.ID = meta.ID

	var r RequestInterface
	switch meta.Method {
	case methods.METHOD_REQUEST_INITIALIZE:
		r = new(InitializeRequest)
	case methods.METHOD_REQUEST_PING:
		r = new(PingRequest)
	case methods.METHOD_REQUEST_LIST_RESOURCES:
		r = new(ListResourcesRequest)
	case methods.METHOD_REQUEST_TEMPLATES_LIST_RESOURCES:
		r = new(ListResourceTemplatesRequest)
	case methods.METHOD_REQUEST_READ_RESOURCES:
		r = new(ReadResourceRequest)
	case methods.METHOD_REQUEST_SUBSCRIBE_RESOURCES:
		r = new(SubscribeRequest)
	case methods.METHOD_REQUEST_UNSUBSCRIBE_RESOURCES:
		r = new(UnsubscribeRequest)
	case methods.METHOD_REQUEST_LIST_PROMPTS:
		r = new(ListPromptsRequest)
	case methods.METHOD_REQUEST_GET_PROMPTS:
		r = new(GetPromptRequest)
	case methods.METHOD_REQUEST_LIST_TOOLS:
		r = new(ListToolsRequest)
	case methods.METHOD_REQUEST_CALL_TOOLS:
		r = new(CallToolRequest)
	case methods.METHOD_REQUEST_SET_LEVEL_LOGGING:
		r = new(SetLevelRequest)
	case methods.METHOD_SAMPLING_CREATE_MESSAGE:
		r = new(CreateMessageRequest)
	case methods.METHOD_AUTOCOMPLETE_COMPLETE:
		r = new(CompleteRequest)
	case methods.METHOD_LIST_ROOTS:
		r = new(ListRootsRequest)
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("error unmarshaling method: %s, err: %v", meta.Method, err)
	}
	jr.RequestInterface = r
	return nil
}
func (jr *JSONRPCRequest) String() string {
	return fmt.Sprintf("jsonrpc: %s requestID: %d request: %s", jr.JSONRPC, jr.ID, jr.GetRequest().Method)
}

type JSONRPCNotification struct {
	//JSONRPC version, should be "2.0"
	JSONRPC string `json:"jsonrpc"`
	NotificationInterface
}

func (jn *JSONRPCNotification) JSONRPCMessageType() int {
	return JSONRPC_MESSAGE_JSONRPC_NOTIFICATION_TYPE
}
func (jn *JSONRPCNotification) JSONRPCBatchRequestType() int {
	return JSONRPC_BATCH_REQUEST_JSONRPC_NOTIFICATION_TYPE
}
func (jn *JSONRPCNotification) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		JSONRPC string `json:"jsonrpc"`
	}{
		JSONRPC: jn.JSONRPC,
	}
	knownFields, err := json.Marshal(&aux)
	if err != nil {
		return nil, fmt.Errorf("marshal known fields: %w", err)
	}
	//Marshal knownFields to map
	baseMap := make(map[string]interface{})
	if err := json.Unmarshal(knownFields, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal known fields to map: %w", err)
	}
	//Marshal NotificationInterface
	var nifInB []byte
	switch jn.NotificationInterface.TypeOfNotification() {
	case CANCELLED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE:
		nifInB, err = json.Marshal(jn.NotificationInterface.(*CancelledNotification))
		if err != nil {
			return nil, fmt.Errorf("marshal CANCELLED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE fields: %w", err)
		}
	case INITIALIZED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE:
		nifInB, err = json.Marshal(jn.NotificationInterface.(*InitializedNotification))
		if err != nil {
			return nil, fmt.Errorf("marshal INITIALIZED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE fields: %w", err)
		}
	case PROGRESS_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE:
		nifInB, err = json.Marshal(jn.NotificationInterface.(*ProgressNotification))
		if err != nil {
			return nil, fmt.Errorf("marshal PROGRESS_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE fields: %w", err)
		}
	case NOTIFICATION_NOTIFICATION_INTERFACE_TYPE:
		nifInB, err = json.Marshal(jn.NotificationInterface.(*ProgressNotification))
		if err != nil {
			return nil, fmt.Errorf("marshal PROGRESS_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE fields: %w", err)
		}
	case LOGGING_MESSAGE_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE:
		nifInB, err = json.Marshal(jn.NotificationInterface.(*LoggingMessageNotification))
		if err != nil {
			return nil, fmt.Errorf("marshal LOGGING_MESSAGE_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE fields: %w", err)
		}
	case RESOURCE_UPDATED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE:
		nifInB, err = json.Marshal(jn.NotificationInterface.(*ResourceUpdatedNotification))
		if err != nil {
			return nil, fmt.Errorf("marshal RESOURCE_UPDATED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE fields: %w", err)
		}
	case RESOURCE_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE:
		nifInB, err = json.Marshal(jn.NotificationInterface.(*ResourceListChangedNotification))
		if err != nil {
			return nil, fmt.Errorf("marshal RESOURCE_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE fields: %w", err)
		}
	case TOOL_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE:
		nifInB, err = json.Marshal(jn.NotificationInterface.(*ToolListChangedNotification))
		if err != nil {
			return nil, fmt.Errorf("marshal TOOL_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE fields: %w", err)
		}
	case PROMPT_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE:
		nifInB, err = json.Marshal(jn.NotificationInterface.(*PromptListChangedNotification))
		if err != nil {
			return nil, fmt.Errorf("marshal PROMPT_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE fields: %w", err)
		}
	}
	if err := json.Unmarshal(nifInB, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields for:%d err:%w", jn.NotificationInterface.TypeOfNotification(), err)
	}
	return json.Marshal(baseMap)
}
func (jn *JSONRPCNotification) UnmarshalJSON(data []byte) error {
	var meta struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("error unmarshaling global meta: %v", err)
	}
	var n NotificationInterface
	switch meta.Method {
	case methods.METHOD_NOTIFICATION_CANCELLED:
		n = new(CancelledNotification)
	case methods.METHOD_NOTIFICATION_INITIALIZED:
		n = new(InitializedNotification)
	case methods.METHOD_NOTIFICATION_PROGRESS:
		n = new(ProgressNotification)
	case methods.METHOD_NOTIFICATION_RESOURCES_LIST_CHANGED:
		n = new(ResourceListChangedNotification)
	case methods.METHOD_NOTIFICATION_RESOURCES_UPDATED:
		n = new(ResourceUpdatedNotification)
	case methods.METHOD_NOTIFICATION_PROMPTS_LIST_CHANGED:
		n = new(PromptListChangedNotification)
	case methods.METHOD_NOTIFICATION_TOOLS_LIST_CHANGED:
		n = new(ToolListChangedNotification)
	case methods.METHOD_NOTIFICATION_MESSAGE:
		n = new(LoggingMessageNotification)
	case methods.METHOD_NOTIFICATION_ROOTS_LIST_CHANGED:
		n = new(RootsListChangedNotification)
	}
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("error unmarshaling method: %s, err: %v", meta.Method, err)
	}
	jn.NotificationInterface = n
	return nil
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
func (jr *JSONRPCResponse) GetRequestID() RequestID { return jr.ID }

func (jr *JSONRPCResponse) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		JSONRPC string    `json:"jsonrpc"`
		ID      RequestID `json:"id"`
	}{
		JSONRPC: jr.JSONRPC,
		ID:      jr.ID,
	}
	knownFields, err := json.Marshal(&aux)
	if err != nil {
		return nil, fmt.Errorf("marshal known fields: %w", err)
	}
	//Marshal knownFields to map
	baseMap := make(map[string]interface{})
	if err := json.Unmarshal(knownFields, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal known fields to map: %w", err)
	}
	if jr.Result == nil {
		return json.Marshal(baseMap)
	}
	//Process ResultInterface
	jmTyp := jr.Result.TypeOfResultInterface()
	switch jmTyp {
	case EMPTY_RESULT_RESULT_INTERFACE_TYPE:
		msER, okType := jr.Result.(*EmptyResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for EMPTY_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msER
	case CREATE_MESSAGE_RESULT_RESULT_INTERFACE_TYPE:
		msCMR, okType := jr.Result.(*CreateMessageResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for CREATE_MESSAGE_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msCMR
	case LIST_ROOTS_RESULT_RESULT_INTERFACE_TYPE:
		msIR, okType := jr.Result.(*ListRootsResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for LIST_ROOTS_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msIR
	case INITIALIZE_RESULT_RESULT_INTERFACE_TYPE:
		msIR, okType := jr.Result.(*InitializeResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for INITIALIZE_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msIR
	case COMPLETE_RESULT_RESULT_INTERFACE_TYPE:
		msCR, okType := jr.Result.(*CompleteResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for COMPLETE_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msCR
	case GET_PROMPT_RESULT_RESULT_INTERFACE_TYPE:
		msGPR, okType := jr.Result.(*GetPromptResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for GET_PROMPT_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msGPR
	case LIST_PROMPTS_RESULT_RESULT_INTERFACE_TYPE:
		msLRR, okType := jr.Result.(*ListPromptsResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for LIST_PROMPTS_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msLRR
	case LIST_RESOURCE_TEMPLATES_RESULT_RESULT_INTERFACE_TYPE:
		msLRR, okType := jr.Result.(*ListResourceTemplatesResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for LIST_RESOURCE_TEMPLATES_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msLRR
	case LIST_RESOURCES_RESULT_RESULT_INTERFACE_TYPE:
		msLRR, okType := jr.Result.(*ListResourcesResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for LIST_RESOURCES_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msLRR
	case READ_RESOURCE_RESULT_RESULT_INTERFACE_TYPE:
		msRRR, okType := jr.Result.(*ReadResourceResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for READ_RESOURCE_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msRRR
	case CALL_TOOL_RESULT_RESULT_INTERFACE_TYPE:
		msCTR, okType := jr.Result.(*CallToolResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for CALL_TOOL_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msCTR
	case LIST_TOOLS_RESULT_RESULT_INTERFACE_TYPE:
		msLTR, okType := jr.Result.(*ListToolsResult)
		if !okType {
			return nil, fmt.Errorf("invalid type for LIST_TOOLS_RESULT_RESULT_INTERFACE_TYPE")
		}
		baseMap["result"] = *msLTR
	default:
		return nil, fmt.Errorf("ResultInterface invalid type for %d", jmTyp)
	}
	return json.Marshal(baseMap)
}

func (jr *JSONRPCResponse) UnmarshalJSON(data []byte) error {
	var meta struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      RequestID       `json:"id"`
		Result  json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("error unmarshaling global meta: %v", err)
	}
	jr.JSONRPC = meta.JSONRPC
	jr.ID = meta.ID
	resultDataMap := make(map[string]interface{})
	if err := json.Unmarshal(meta.Result, &resultDataMap); err != nil {
		return fmt.Errorf("error unmarshaling global data in map: %v", err)
	}

	var r ResultInterface
	var resultFactoriesSortedKeys = []string{
		"prompts",
		"resources",
		"resourceTemplates",
		"completion",
		"tools",
		"messages",
		"contents",
		"protocolVersion",
		"nextCursor",
		"roots",
		"model",
		"content",
	}
	var resultFactories = map[string]func() ResultInterface{
		"prompts":           func() ResultInterface { return new(ListPromptsResult) },
		"resources":         func() ResultInterface { return new(ListResourcesResult) },
		"resourceTemplates": func() ResultInterface { return new(ListResourceTemplatesResult) },
		"completion":        func() ResultInterface { return new(CompleteResult) },
		"tools":             func() ResultInterface { return new(ListToolsResult) },
		"messages":          func() ResultInterface { return new(GetPromptResult) },
		"contents":          func() ResultInterface { return new(ReadResourceResult) },
		"protocolVersion":   func() ResultInterface { return new(InitializeResult) },
		"nextCursor":        func() ResultInterface { return new(PaginatedResult) },
		"roots":             func() ResultInterface { return new(ListRootsResult) },
		"model":             func() ResultInterface { return new(CreateMessageResult) },
		"content":           func() ResultInterface { return new(CallToolResult) },
	}

	for _, key := range resultFactoriesSortedKeys {
		if _, ok := resultDataMap[key]; ok {
			if builder, okBuilder := resultFactories[key]; okBuilder {
				r = builder()
				break
			}

		}
	}
	if r == nil {
		r = new(EmptyResult)
	}

	if err := json.Unmarshal(meta.Result, &r); err != nil {
		return fmt.Errorf("error unmarshaling err: %v", err)
	}
	jr.Result = r
	return nil
}

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
