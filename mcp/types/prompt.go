package types

import (
	"encoding/json"
	"fmt"

	"github.com/victorvbello/gomcp/mcp/methods"
)

const (
	AUTOCOMPLETE_REF_PROMPT_TYPE = "ref/prompt"
)

//A prompt or prompt template that the server offers.
type Prompt struct {
	BaseMetadata
	//An optional description of what this prompt provides
	Description string `json:"description,omitempty"`
	//A list of arguments to use for templating the prompt.
	Arguments []PromptArgument `json:"arguments,omitempty"`
	//See [MCP specification](https://github.com/modelcontextprotocol/modelcontextprotocol/blob/47339c03c143bb4ec01a26e721a1b8fe66634ebe/docs/specification/draft/basic/index.mdx#general-fields)
	//for notes on _meta usage.
	Metadata map[string]interface{} `json:"_meta,omitempty"`
	//Attach additional properties, _meta is reserved by MCP
	AdditionalProperties map[string]interface{} `json:"-"`
}

func (p *Prompt) MarshalJSON() ([]byte, error) {
	raw := make(map[string]interface{})
	if p.Metadata != nil {
		raw["_meta"] = p.Metadata
	}
	for key, value := range p.AdditionalProperties {
		if key == "_meta" {
			continue //Skip the _meta key is reserved by MCP
		}
		raw[key] = value
	}

	return json.Marshal(raw)
}

func (p *Prompt) UnmarshalJSON(data []byte) error {
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
	if err := json.Unmarshal(bm, &p.Metadata); err != nil {
		return fmt.Errorf("error unmarshaling into metadata: %v", err)
	}
	delete(raw, "_meta")
	p.AdditionalProperties = raw
	return nil
}

//Describes an argument that a prompt can accept.
type PromptArgument struct {
	//The name of the argument.
	Name string `json:"name"`
	//A human-readable description of the argument.
	Description string `json:"description,omitempty"`
	//Whether this argument must be provided.
	Required bool `json:"required,omitempty"`
}

//Describes a message returned as part of a prompt.
//
//This is similar to `SamplingMessage`, but also supports the embedding of
//resources from the MCP server.
type PromptMessage struct {
	Role Role `json:"role"`
	//Could be TextContent/ImageContent/AudioContent/EmbeddedResource
	Content Content `json:"content"`
}

//An optional notification from the server to the client, informing it that the list of prompts it offers has changed. This may be issued by servers without any previous subscription from the client.
//
//Only method: METHOD_NOTIFICATION_PROMPTS_LIST_CHANGED
type PromptListChangedNotification struct {
	Notification
}

//An optional notification from the server to the client, informing it that the list of prompts it offers has changed. This may be issued by servers without any previous subscription from the client.
func (plcn *PromptListChangedNotification) TypeOfServerNotification() int {
	return PROMPT_LIST_CHANGED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
}
func (plcn *PromptListChangedNotification) TypeOfNotification() int {
	return PROMPT_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
}
func (plcn *PromptListChangedNotification) GetNotification() Notification {
	return plcn.Notification
}

func NewPromptListChangedNotification(params *BaseNotificationParams) *PromptListChangedNotification {
	plcn := new(PromptListChangedNotification)
	plcn.Method = methods.METHOD_NOTIFICATION_PROMPTS_LIST_CHANGED
	plcn.Params = params
	return plcn
}

//Sent from the client to request a list of prompts and prompt templates the server has.
//
//Only method: METHOD_REQUEST_LIST_PROMPTS
type ListPromptsRequest struct {
	PaginatedRequest
}

func (lp *ListPromptsRequest) TypeOfClientRequest() int {
	return LIST_PROMPTS_REQUEST_CLIENT_REQUEST_TYPE
}

func (lp *ListPromptsRequest) TypeOfRequestInterface() int {
	return LIST_PROMPTS_REQUEST_REQUEST_INTERFACE_TYPE
}
func (lp *ListPromptsRequest) GetRequest() Request {
	return lp.Request
}

func NewListPromptsRequest(params *PaginatedRequestParams) *ListPromptsRequest {
	lpr := new(ListPromptsRequest)
	lpr.Method = methods.METHOD_REQUEST_LIST_PROMPTS
	lpr.Params = params
	return lpr
}

//Used by the client to get a prompt provided by the server.
//
//Only method: METHOD_REQUEST_GET_PROMPTS
type GetPromptRequest struct {
	Request
	Params GetPromptParams `json:"params"`
}

func (g *GetPromptRequest) TypeOfClientRequest() int { return GET_PROMPT_REQUEST_CLIENT_REQUEST_TYPE }
func (g *GetPromptRequest) TypeOfRequestInterface() int {
	return GET_PROMPT_REQUEST_REQUEST_INTERFACE_TYPE
}
func (g *GetPromptRequest) GetRequest() Request {
	return g.Request
}

func NewGetPromptRequest(params *GetPromptParams) *GetPromptRequest {
	gpr := new(GetPromptRequest)
	gpr.Method = methods.METHOD_REQUEST_GET_PROMPTS
	if params != nil {
		gpr.Params = *params
	}
	return gpr
}

type GetPromptParams struct {
	BaseRequestParams
	//The name of the prompt or prompt template.
	Name string `json:"name"`
	//Arguments to use for templating the prompt.
	Arguments map[string]string `json:"arguments,omitempty"`
}

func (gpp *GetPromptParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments,omitempty"`
	}{
		Name:      gpp.Name,
		Arguments: gpp.Arguments,
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
	baseExtra, err := gpp.BaseRequestParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (gpp *GetPromptParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments,omitempty"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	gpp.Name = aux.Name
	gpp.Arguments = aux.Arguments

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "name")
	delete(raw, "arguments")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := gpp.BaseRequestParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("baseRequestParams.UnmarshalJSON: %w", err)
	}
	return nil
}

//The server's response to a prompts/list request from the client.
type ListPromptsResult struct {
	PaginatedResult
	Prompts []Prompt `json:"prompts"`
}

func (lpr *ListPromptsResult) TypeOfServerResult() int { return LIST_PROMPTS_RESULT_SERVER_RESULT_TYPE }
func (lpr *ListPromptsResult) TypeOfResultInterface() int {
	return LIST_PROMPTS_RESULT_RESULT_INTERFACE_TYPE
}

//The server's response to a prompts/get request from the client.
type GetPromptResult struct {
	Result
	//An optional description for the prompt.
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

func (gpr *GetPromptResult) TypeOfServerResult() int { return GET_PROMPT_RESULT_SERVER_RESULT_TYPE }
func (gpr *GetPromptResult) TypeOfResultInterface() int {
	return GET_PROMPT_RESULT_RESULT_INTERFACE_TYPE
}

//Identifies a prompt.
type PromptReference struct {
	//Only AUTOCOMPLETE_REF_PROMPT_TYPE
	Type string `json:"type"`
	//The name of the prompt or prompt template
	Name string `json:"name"`
}

func (p *PromptReference) AutoCompleteRefType() string { return AUTOCOMPLETE_REF_PROMPT_TYPE }

func NewPromptReference(name string) *PromptReference {
	npr := PromptReference{
		Type: AUTOCOMPLETE_REF_PROMPT_TYPE,
		Name: name,
	}
	return &npr
}
