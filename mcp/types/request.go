package types

import (
	"encoding/json"
	"fmt"

	"github.com/victorvbello/gomcp/mcp/methods"
)

const (
	PING_REQUEST_CLIENT_REQUEST_TYPE = iota + 20
	INITIALIZE_REQUEST_CLIENT_REQUEST_TYPE
	COMPLETE_REQUEST_CLIENT_REQUEST_TYPE
	SET_LEVEL_REQUEST_CLIENT_REQUEST_TYPE
	GET_PROMPT_REQUEST_CLIENT_REQUEST_TYPE
	LIST_PROMPTS_REQUEST_CLIENT_REQUEST_TYPE
	LIST_RESOURCES_REQUEST_CLIENT_REQUEST_TYPE
	LIST_RESOURCE_TEMPLATES_REQUEST_CLIENT_REQUEST_TYPE
	READ_RESOURCE_REQUEST_CLIENT_REQUEST_TYPE
	SUBSCRIBE_REQUEST_CLIENT_REQUEST_TYPE
	UNSUBSCRIBE_REQUEST_CLIENT_REQUEST_TYPE
	CALL_TOOL_REQUEST_CLIENT_REQUEST_TYPE
	LIST_TOOLS_REQUEST_CLIENT_REQUEST_TYPE
)

const (
	PING_REQUEST_SERVER_REQUEST_TYPE = iota + 40
	CREATE_MESSAGE_REQUEST_SERVER_REQUEST_TYPE
	LIST_ROOTS_REQUEST_SERVER_REQUEST_TYPE
)

const (
	PING_REQUEST_REQUEST_INTERFACE_TYPE = iota + 500
	INITIALIZE_REQUEST_REQUEST_INTERFACE_TYPE
	REQUEST_REQUEST_INTERFACE_TYPE
	CREATE_MESSAGE_REQUEST_REQUEST_INTERFACE_TYPE
	LIST_ROOTS_REQUEST_REQUEST_INTERFACE_TYPE
	SET_LEVEL_REQUEST_REQUEST_INTERFACE_TYPE
	GET_PROMPT_REQUEST_REQUEST_INTERFACE_TYPE
	LIST_PROMPTS_REQUEST_REQUEST_INTERFACE_TYPE
	LIST_RESOURCES_REQUEST_REQUEST_INTERFACE_TYPE
	LIST_RESOURCE_TEMPLATES_REQUEST_REQUEST_INTERFACE_TYPE
	READ_RESOURCE_REQUEST_REQUEST_INTERFACE_TYPE
	CALL_TOOL_REQUEST_REQUEST_INTERFACE_TYPE
	LIST_TOOLS_REQUEST_REQUEST_INTERFACE_TYPE
)

//A uniquely identifying ID for a request in JSON-RPC, number.
type RequestID int

type Request struct {
	Method string             `json:"method"`
	Params *BaseRequestParams `json:"params,omitempty"`
}

func (r *Request) TypeOfRequestInterface() int { return REQUEST_REQUEST_INTERFACE_TYPE }
func (r *Request) GetRequest() Request         { return *r }

type MetadataRequest struct {
	ProgressToken ProgressToken `json:"progressToken,omitempty"`
}

func NewMetadataRequestFromMetadata(meta map[string]interface{}) (*MetadataRequest, error) {
	metaB, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("error marshal  metadata: %v", err)
	}
	var result MetadataRequest
	if err := json.Unmarshal(metaB, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling metadata: %v", err)
	}
	return &result, nil
}

type BaseRequestParams struct {
	//Attach additional metadata to their notifications.
	Metadata map[string]interface{} `json:"_meta,omitempty"`
	//Attach additional properties, _meta is reserved by MCP
	AdditionalProperties map[string]interface{} `json:"-"`
}

func (rp *BaseRequestParams) MarshalJSON() ([]byte, error) {
	raw := make(map[string]interface{})
	if rp.Metadata != nil {
		raw["_meta"] = rp.Metadata
	}
	for key, value := range rp.AdditionalProperties {
		if key == "_meta" {
			continue //Skip the _meta key is reserved by MCP
		}
		raw[key] = value
	}
	return json.Marshal(raw)
}

func (rp *BaseRequestParams) UnmarshalJSON(data []byte) error {
	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	if _, ok := raw["_meta"]; !ok {
		return nil //No _meta field, nothing to unmarshal
	}
	bm, err := json.Marshal(raw["_meta"])
	if err != nil {
		return fmt.Errorf("error marshaling _meta: %v", err)
	}
	meta := map[string]interface{}{}
	if err := json.Unmarshal(bm, &meta); err != nil {
		return fmt.Errorf("error unmarshaling into metadata: %v", err)
	}
	delete(raw, "_meta")
	rp.Metadata = meta
	rp.AdditionalProperties = raw
	return nil
}

//This request is sent from the client to the server when it first connects, asking it to begin initialization.
//
//Only method: METHOD_REQUEST_INITIALIZE
type InitializeRequest struct {
	Request
	Params InitializeRequestParams `json:"params"`
}

func (i *InitializeRequest) TypeOfClientRequest() int { return INITIALIZE_REQUEST_CLIENT_REQUEST_TYPE }
func (i *InitializeRequest) TypeOfRequestInterface() int {
	return INITIALIZE_REQUEST_REQUEST_INTERFACE_TYPE
}
func (i *InitializeRequest) GetRequest() Request { return i.Request }

func NewInitializeRequest(params *InitializeRequestParams) *InitializeRequest {
	newIR := InitializeRequest{
		Request: Request{Method: methods.METHOD_REQUEST_INITIALIZE},
	}
	if params != nil {
		newIR.Params = *params
	}
	return &newIR
}

type InitializeRequestParams struct {
	BaseRequestParams
	//The latest version of the Model Context Protocol that the client supports. The client MAY decide to support older versions as well.
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      Implementation     `json:"clientInfo"`
}

func (iRP *InitializeRequestParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		ProtocolVersion string             `json:"protocolVersion"`
		Capabilities    ClientCapabilities `json:"capabilities"`
		ClientInfo      Implementation     `json:"clientInfo"`
	}{
		ProtocolVersion: iRP.ProtocolVersion,
		Capabilities:    iRP.Capabilities,
		ClientInfo:      iRP.ClientInfo,
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
	//Marshal base.BaseRequestParams
	baseExtra, err := iRP.BaseRequestParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (iRP *InitializeRequestParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		ProtocolVersion string             `json:"protocolVersion"`
		Capabilities    ClientCapabilities `json:"capabilities"`
		ClientInfo      Implementation     `json:"clientInfo"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	iRP.ProtocolVersion = aux.ProtocolVersion
	iRP.Capabilities = aux.Capabilities
	iRP.ClientInfo = aux.ClientInfo

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "protocolVersion")
	delete(raw, "capabilities")
	delete(raw, "clientInfo")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := iRP.BaseRequestParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("baseRequestParams.UnmarshalJSON: %w", err)
	}
	return nil
}

//Capabilities a client may support. Known capabilities are defined here, in this schema, but this is not a closed set: any client can define its own, additional capabilities.
type ClientCapabilities struct {
	//Experimental, non-standard capabilities that the client supports.
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	//Present if the client supports listing roots.
	Roots *struct {
		//Whether the client supports notifications for changes to the roots list.
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"roots,omitempty"`
	//Present if the client supports sampling from an LLM.
	Sampling interface{} `json:"sampling,omitempty"`
}

//A ping, issued by either the server or the client, to check that the other party is still alive. The receiver must promptly respond, or else may be disconnected.
//
//Only method: METHOD_REQUEST_PING
type PingRequest struct {
	Request
}

func (p *PingRequest) TypeOfClientRequest() int    { return PING_REQUEST_CLIENT_REQUEST_TYPE }
func (p *PingRequest) TypeOfServerRequest() int    { return PING_REQUEST_SERVER_REQUEST_TYPE }
func (p *PingRequest) TypeOfRequestInterface() int { return PING_REQUEST_REQUEST_INTERFACE_TYPE }
func (p *PingRequest) GetRequest() Request         { return p.Request }

func NewPingRequest() *PingRequest {
	newPing := PingRequest{
		Request: Request{Method: methods.METHOD_REQUEST_PING},
	}
	return &newPing
}

type PaginatedRequest struct {
	Request
	Params *PaginatedRequestParams `json:"params,omitempty"`
}

type PaginatedRequestParams struct {
	BaseRequestParams
	//An opaque token representing the current pagination position.
	//If provided, the server should return results starting after this cursor.
	Cursor Cursor `json:"cursor,omitempty"`
}

func (prp *PaginatedRequestParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		Cursor Cursor `json:"cursor,omitempty"`
	}{
		Cursor: prp.Cursor,
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
	//Marshal base.BaseRequestParams
	baseExtra, err := prp.BaseRequestParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (prp *PaginatedRequestParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		Cursor Cursor `json:"cursor,omitempty"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	prp.Cursor = aux.Cursor

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "cursor")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := prp.BaseRequestParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("baseRequestParams.UnmarshalJSON: %w", err)
	}
	return nil
}

//Sent from the client to request resources/updated notifications from the server whenever a particular resource changes.
//
//Only method: METHOD_REQUEST_SUBSCRIBE_RESOURCES
type SubscribeRequest struct {
	Request
	Params SubscribeRequestParams `json:"params"`
}

func (sr *SubscribeRequest) TypeOfClientRequest() int { return SUBSCRIBE_REQUEST_CLIENT_REQUEST_TYPE }

type SubscribeRequestParams struct {
	BaseRequestParams
	//The URI of the resource to subscribe to. The URI can use any protocol; it is up to the server how to interpret it.
	URI string `json:"uri"`
}

func (srp *SubscribeRequestParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		URI string `json:"uri"`
	}{
		URI: srp.URI,
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
	//Marshal base.BaseRequestParams
	baseExtra, err := srp.BaseRequestParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (srp *SubscribeRequestParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		URI string `json:"uri"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	srp.URI = aux.URI

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "uri")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := srp.BaseRequestParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("baseRequestParams.UnmarshalJSON: %w", err)
	}
	return nil
}

//Sent from the client to request cancellation of resources/updated notifications from the server. This should follow a previous resources/subscribe request.
//
//Only method: METHOD_REQUEST_UNSUBSCRIBE_RESOURCES
type UnsubscribeRequest struct {
	Request
	Params UnsubscribeRequestParams `json:"params"`
}

func (ur *UnsubscribeRequest) TypeOfClientRequest() int {
	return UNSUBSCRIBE_REQUEST_CLIENT_REQUEST_TYPE
}

type UnsubscribeRequestParams struct {
	BaseRequestParams
	//The URI of the resource to unsubscribe from.
	URI string `json:"uri"`
}

func (urp *UnsubscribeRequestParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		URI string `json:"uri"`
	}{
		URI: urp.URI,
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
	//Marshal base.BaseRequestParams
	baseExtra, err := urp.BaseRequestParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (urp *UnsubscribeRequestParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		URI string `json:"uri"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	urp.URI = aux.URI

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "uri")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := urp.BaseRequestParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("baseRequestParams.UnmarshalJSON: %w", err)
	}
	return nil
}

type RequestInterface interface {
	TypeOfRequestInterface() int
	GetRequest() Request
}

//ClientRequest
type ClientRequest interface {
	TypeOfClientRequest() int
}

//Server messages
type ServerRequest interface {
	TypeOfServerRequest() int
}
