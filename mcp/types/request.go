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
	JSONRPC_REQUEST_REQUEST_INTERFACE_TYPE
)

//A uniquely identifying ID for a request in JSON-RPC, number.
type RequestID int

type Request struct {
	Method string         `json:"method"`
	Params *RequestParams `json:"params,omitempty"`
}

type MetadataRequest struct {
	ProgressToken ProgressToken `json:"progressToken,omitempty"`
}

type RequestParams struct {
	//Attach additional metadata to their notifications.
	Metadata *MetadataRequest `json:"_meta,omitempty"`
	//Attach additional properties, _meta is reserved by MCP
	AdditionalProperties map[string]interface{} `json:"-"`
}

func (rp *RequestParams) MarshalJSON() ([]byte, error) {
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

func (rp *RequestParams) UnmarshalJSON(data []byte) error {
	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	bm, err := json.Marshal(raw["_meta"])
	if err != nil {
		return fmt.Errorf("error marshaling _meta: %v", err)
	}
	if err := json.Unmarshal(bm, rp.Metadata); err != nil {
		return fmt.Errorf("error unmarshaling into metadata: %v", err)
	}

	delete(raw, "_meta")
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

type InitializeRequestParams struct {
	//The latest version of the Model Context Protocol that the client supports. The client MAY decide to support older versions as well.
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      Implementation     `json:"clientInfo"`
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
	//An opaque token representing the current pagination position.
	//If provided, the server should return results starting after this cursor.
	Cursor Cursor `json:"cursor,omitempty"`
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
	//The URI of the resource to subscribe to. The URI can use any protocol; it is up to the server how to interpret it.
	URI string `json:"uri"`
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
	//The URI of the resource to unsubscribe from.
	URI string `json:"uri"`
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
