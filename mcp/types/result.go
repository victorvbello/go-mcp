package types

import (
	"encoding/json"
	"fmt"
)

const (
	EMPTY_RESULT_CLIENT_RESULT_TYPE = iota + 60
	CREATE_MESSAGE_RESULT_CLIENT_RESULT_TYPE
	LIST_ROOTS_RESULT_CLIENT_RESULT_TYPE
)

const (
	EMPTY_RESULT_SERVER_RESULT_TYPE = iota + 70
	INITIALIZE_RESULT_SERVER_RESULT_TYPE
	PAGINATED_RESULT_RESULT_INTERFACE_TYPE
	COMPLETE_RESULT_SERVER_RESULT_TYPE
	GET_PROMPT_RESULT_SERVER_RESULT_TYPE
	LIST_PROMPTS_RESULT_SERVER_RESULT_TYPE
	LIST_RESOURCE_TEMPLATES_RESULT_SERVER_RESULT_TYPE
	LIST_RESOURCES_RESULT_SERVER_RESULT_TYPE
	READ_RESOURCE_RESULT_SERVER_RESULT_TYPE
	CALL_TOOL_RESULT_SERVER_RESULT_TYPE
	LIST_TOOLS_RESULT_SERVER_RESULT_TYPE
)

const (
	EMPTY_RESULT_RESULT_INTERFACE_TYPE = iota + 400
	CREATE_MESSAGE_RESULT_RESULT_INTERFACE_TYPE
	LIST_ROOTS_RESULT_RESULT_INTERFACE_TYPE
	INITIALIZE_RESULT_RESULT_INTERFACE_TYPE
	COMPLETE_RESULT_RESULT_INTERFACE_TYPE
	GET_PROMPT_RESULT_RESULT_INTERFACE_TYPE
	LIST_PROMPTS_RESULT_RESULT_INTERFACE_TYPE
	LIST_RESOURCE_TEMPLATES_RESULT_RESULT_INTERFACE_TYPE
	LIST_RESOURCES_RESULT_RESULT_INTERFACE_TYPE
	READ_RESOURCE_RESULT_RESULT_INTERFACE_TYPE
	CALL_TOOL_RESULT_RESULT_INTERFACE_TYPE
	LIST_TOOLS_RESULT_RESULT_INTERFACE_TYPE
)

//A response that indicates success but carries no data.
type EmptyResult Result

func (ep *EmptyResult) TypeOfClientResult() int    { return EMPTY_RESULT_CLIENT_RESULT_TYPE }
func (ep *EmptyResult) TypeOfServerResult() int    { return EMPTY_RESULT_SERVER_RESULT_TYPE }
func (ep *EmptyResult) TypeOfResultInterface() int { return EMPTY_RESULT_RESULT_INTERFACE_TYPE }

type Result struct {
	//Attach additional metadata to their notifications.
	Metadata map[string]interface{} `json:"_meta,omitempty"`
	//Attach additional properties, _meta is reserved by MCP
	AdditionalProperties map[string]interface{} `json:"-"`
}

func (re *Result) MarshalJSON() ([]byte, error) {
	raw := make(map[string]interface{})
	if re.Metadata != nil {
		raw["_meta"] = re.Metadata
	}
	for key, value := range re.AdditionalProperties {
		if key == "_meta" {
			continue //Skip the _meta key is reserved by MCP
		}
		raw[key] = value
	}

	return json.Marshal(raw)
}

func (re *Result) UnmarshalJSON(data []byte) error {
	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if _, ok := raw["_meta"]; !ok {
		return nil //No _meta field, nothing to unmarshal
	}
	bm, err := json.Marshal(raw["_meta"])
	if err != nil {
		return fmt.Errorf("error marshaling _meta: %v", err)
	}
	if err := json.Unmarshal(bm, &re.Metadata); err != nil {
		return fmt.Errorf("error unmarshaling into metadata: %v", err)
	}
	delete(raw, "_meta")
	re.AdditionalProperties = raw
	return nil
}

type ServerCapabilitiesListChanged struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ServerCapabilitiesResources struct {
	//Whether this server supports subscribing to resource updates.
	Subscribe bool `json:"subscribe,omitempty"`
	//Whether this server supports notifications for changes to the resource list.
	ServerCapabilitiesListChanged
}

//Capabilities that a server may support. Known capabilities are defined here, in this schema, but this is not a closed set: any server can define its own, additional capabilities.
type ServerCapabilities struct {
	//Experimental, non-standard capabilities that the server supports.
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	//Present if the server supports sending log messages to the client.
	Logging interface{} `json:"logging,omitempty"`
	//Present if the server supports argument autocompletion suggestions.
	Completions interface{} `json:"completions,omitempty"`
	//Present if the server supports sampling from an LLM.
	Sampling interface{} `json:"sampling,omitempty"`
	//Present if the server offers any prompt templates.
	//Whether this server supports notifications for changes to the prompt list.
	Prompts *ServerCapabilitiesListChanged `json:"prompts,omitempty"`
	//Present if the server offers any resources to read.
	Resources *ServerCapabilitiesResources `json:"resources,omitempty"`
	//Present if the server offers any tools to call.
	//Whether this server supports notifications for changes to the tool list.
	Tools *ServerCapabilitiesListChanged `json:"tools,omitempty"`
}

//After receiving an initialize request from the client, the server sends this response.
type InitializeResult struct {
	Result
	//The version of the Model Context Protocol that the server wants to use. This may not match the version that the client requested. If the client cannot support this version, it MUST disconnect.
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
	//Instructions describing how to use the server and its features.
	//
	//This can be used by clients to improve the LLM's understanding of available tools, resources, etc. It can be thought of like a "hint" to the model. For example, this information MAY be added to the system prompt.
	Instructions string `json:"instructions,omitempty"`
}

func (ir *InitializeResult) TypeOfServerResult() int { return INITIALIZE_RESULT_SERVER_RESULT_TYPE }
func (ir *InitializeResult) TypeOfResultInterface() int {
	return INITIALIZE_RESULT_RESULT_INTERFACE_TYPE
}

type PaginatedResult struct {
	Result
	//An opaque token representing the pagination position after the last returned result.
	//
	//If present, there may be more results available.
	NextCursor Cursor `json:"nextCursor,omitempty"`
}

func (pr *PaginatedResult) TypeOfResultInterface() int { return PAGINATED_RESULT_RESULT_INTERFACE_TYPE }

type ResultInterface interface {
	TypeOfResultInterface() int
}

type ClientResult interface {
	TypeOfClientResult() int
}

type ServerResult interface {
	TypeOfServerResult() int
}
