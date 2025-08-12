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
	switch req.GetRequest().Method {
	case methods.METHOD_REQUEST_INITIALIZE:
		var ri InitializeRequest
		err = json.Unmarshal(b, &ri)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling %s request: %v", method, err)
		}
		req.RequestInterface = &ri
	case methods.METHOD_REQUEST_PING:
		var p PingRequest
		err = json.Unmarshal(b, &p)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling %s request: %v", method, err)
		}
		req.RequestInterface = &p
		break
	case methods.METHOD_REQUEST_LIST_RESOURCES:
		break
	case methods.METHOD_REQUEST_TEMPLATES_LIST_RESOURCES:
		break
	case methods.METHOD_REQUEST_READ_RESOURCES:
		break
	case methods.METHOD_REQUEST_SUBSCRIBE_RESOURCES:
		break
	case methods.METHOD_REQUEST_UNSUBSCRIBE_RESOURCES:
		break
	case methods.METHOD_REQUEST_LIST_PROMPTS:
		break
	case methods.METHOD_REQUEST_GET_PROMPTS:
		break
	case methods.METHOD_REQUEST_LIST_TOOLS:
		var ltr ListToolsRequest
		err = json.Unmarshal(b, &ltr)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling %s request: %v", method, err)
		}
		req.RequestInterface = &ltr
	case methods.METHOD_REQUEST_CALL_TOOLS:
		var ctr CallToolRequest
		err = json.Unmarshal(b, &ctr)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling %s request: %v", method, err)
		}
		req.RequestInterface = &ctr
		break
	case methods.METHOD_REQUEST_SET_LEVEL_LOGGING:
		break
	case methods.METHOD_SAMPLING_CREATE_MESSAGE:
		var cmr CreateMessageRequest
		err = json.Unmarshal(b, &cmr)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling %s request: %v", method, err)
		}
		req.RequestInterface = &cmr
		break
	case methods.METHOD_AUTOCOMPLETE_COMPLETE:
		break
	case methods.METHOD_LIST_ROOTS:
		break
	}
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
	switch req.GetNotification().Method {
	case methods.METHOD_NOTIFICATION_CANCELLED:
		break
	case methods.METHOD_NOTIFICATION_INITIALIZED:
		var in InitializedNotification
		err = json.Unmarshal(b, &in)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling %s notification: %v", method, err)
		}
		req.NotificationInterface = &in
	case methods.METHOD_NOTIFICATION_PROGRESS:
		break
	case methods.METHOD_NOTIFICATION_RESOURCES_LIST_CHANGED:
		break
	case methods.METHOD_NOTIFICATION_RESOURCES_UPDATED:
		break
	case methods.METHOD_NOTIFICATION_PROMPTS_LIST_CHANGED:
		break
	case methods.METHOD_NOTIFICATION_TOOLS_LIST_CHANGED:
		break
	case methods.METHOD_NOTIFICATION_MESSAGE:
		break
	case methods.METHOD_NOTIFICATION_ROOTS_LIST_CHANGED:
		break
	}
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
