package types

import (
	"encoding/json"
	"fmt"

	"github.com/victorvbello/gomcp/mcp/methods"
)

type RawMessage map[string]interface{}

func (msg RawMessage) IsInitializeRequest() bool {
	if method, ok := msg["method"].(string); ok && method == methods.METHOD_REQUEST_INITIALIZE {
		return true
	}
	return false
}

func (msg RawMessage) ToJSONRPCMessage() (JSONRPCMessage, error) {
	if msg == nil {
		return nil, fmt.Errorf("rawMessage is nil")
	}
	method := msg["method"]
	var safeMethod string
	if method != nil {
		safeMethod = method.(string)
	}
	_, okError := msg["error"]
	_, okResult := msg["result"]
	switch {
	case methods.MethodIn(methods.REQUEST_METHODS, safeMethod):
		return msg.ToJSONRPCRequest()
	case methods.MethodIn(methods.NOTIFICATION_METHODS, safeMethod):
		return msg.ToJSONRPCNotification()
	case okResult:
		return msg.ToJSONRPCResponse()
	case okError:
		return msg.ToJSONRPCError()
	default:
		return nil, nil
	}
}

func (msg RawMessage) ToJSONRPCRequest() (*JSONRPCRequest, error) {
	method := msg["method"].(string)
	if method == "" {
		return nil, nil
	}
	if _, ok := methods.REQUEST_METHODS[method]; !ok {
		return nil, nil
	}
	var req JSONRPCRequest
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}
	err = json.Unmarshal(b, &req)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling %s request: %v", method, err)
	}
	var r RequestInterface
	switch req.GetRequest().Method {
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
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("error unmarshaling method: %s, err: %v", req.GetRequest().Method, err)
	}
	req.RequestInterface = r

	return &req, nil
}

func (msg RawMessage) ToJSONRPCNotification() (*JSONRPCNotification, error) {
	method := msg["method"].(string)
	if method == "" {
		return nil, nil
	}
	if _, ok := methods.NOTIFICATION_METHODS[method]; !ok {
		return nil, nil
	}
	var req JSONRPCNotification
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("error marshalling notification: %v", err)
	}
	err = json.Unmarshal(b, &req)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling %s notification: %v", method, err)
	}
	var n NotificationInterface
	switch req.GetNotification().Method {
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
	err = json.Unmarshal(b, &n)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling %s notification: %v", method, err)
	}
	req.NotificationInterface = n
	return &req, nil
}

func (msg RawMessage) ToJSONRPCResponse() (*JSONRPCResponse, error) {
	if _, ok := msg["result"]; !ok {
		return nil, nil
	}
	var res JSONRPCResponse
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("error marshalling result: %v", err)
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling result: %v", err)
	}
	switch res.Result.TypeOfResultInterface() {
	case EMPTY_RESULT_RESULT_INTERFACE_TYPE:
		break
	case CREATE_MESSAGE_RESULT_RESULT_INTERFACE_TYPE:
		break
	case LIST_ROOTS_RESULT_RESULT_INTERFACE_TYPE:
		break
	case INITIALIZE_RESULT_RESULT_INTERFACE_TYPE:
		break
	case COMPLETE_RESULT_RESULT_INTERFACE_TYPE:
		break
	case GET_PROMPT_RESULT_RESULT_INTERFACE_TYPE:
		break
	case LIST_PROMPTS_RESULT_RESULT_INTERFACE_TYPE:
		break
	case LIST_RESOURCE_TEMPLATES_RESULT_RESULT_INTERFACE_TYPE:
		break
	case LIST_RESOURCES_RESULT_RESULT_INTERFACE_TYPE:
		break
	case READ_RESOURCE_RESULT_RESULT_INTERFACE_TYPE:
		break
	case CALL_TOOL_RESULT_RESULT_INTERFACE_TYPE:
		break
	case LIST_TOOLS_RESULT_RESULT_INTERFACE_TYPE:
		break
	}
	return &res, nil
}

func (msg RawMessage) ToJSONRPCError() (*JSONRPCError, error) {

	if _, ok := msg["error"]; !ok {
		return nil, nil
	}
	var res JSONRPCError
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("error marshalling error type: %v", err)
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling error type: %v", err)
	}
	return &res, nil
}

func MessagesHasSomeInitializeRequest(messages []RawMessage) bool {
	for _, msg := range messages {
		if msg.IsInitializeRequest() {
			return true
		}
	}
	return false
}

func MessagesHasSomeJSONRPCRequest(messages []RawMessage) bool {
	for _, msg := range messages {
		method := msg["method"].(string)
		if method == "" {
			continue
		}
		if _, ok := methods.REQUEST_METHODS[method]; ok {
			return true
		}
	}
	return false
}
