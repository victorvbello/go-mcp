package types

import (
	"encoding/json"
	"fmt"

	"github.com/victorvbello/gomcp/mcp/methods"
)

//Definition for a tool the client can call.
type Tool struct {
	BaseMetadata
	//A human-readable description of the tool.
	//
	//This can be used by clients to improve the LLM's understanding of available tools. It can be thought of like a "hint" to the model.
	Description string `json:"description,omitempty"`
	//A JSON Schema object defining the expected parameters for the tool.
	InputSchema ToolInputSchema `json:"inputSchema"`
	//An optional JSON Schema object defining the structure of the tool's output returned in
	//the structuredContent field of a CallToolResult.
	OutputSchema ToolOutputSchema `json:"outputSchema"`
	//Optional additional tool information.
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
	//See [MCP specification](https://github.com/modelcontextprotocol/modelcontextprotocol/blob/47339c03c143bb4ec01a26e721a1b8fe66634ebe/docs/specification/draft/basic/index.mdx#general-fields)
	//for notes on _meta usage.
	Metadata map[string]interface{} `json:"_meta,omitempty"`
	//Attach additional properties, _meta is reserved by MCP
	AdditionalProperties map[string]interface{} `json:"-"`
}

func (t *Tool) MarshalJSON() ([]byte, error) {
	aux := struct {
		BaseMetadata
		Description  string                 `json:"description,omitempty"`
		InputSchema  ToolInputSchema        `json:"inputSchema"`
		OutputSchema ToolOutputSchema       `json:"outputSchema"`
		Annotations  *ToolAnnotations       `json:"annotations,omitempty"`
		Metadata     map[string]interface{} `json:"_meta,omitempty"`
	}{
		BaseMetadata: t.BaseMetadata,
		Description:  t.Description,
		InputSchema:  t.InputSchema,
		OutputSchema: t.OutputSchema,
		Annotations:  t.Annotations,
		Metadata:     t.Metadata,
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
	for key, value := range t.AdditionalProperties {
		if key == "_meta" {
			continue //Skip the _meta key is reserved by MCP
		}
		baseMap[key] = value
	}

	return json.Marshal(baseMap)
}

func (t *Tool) UnmarshalJSON(data []byte) error {
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
	if err := json.Unmarshal(bm, &t.Metadata); err != nil {
		return fmt.Errorf("error unmarshaling into metadata: %v", err)
	}
	delete(raw, "_meta")
	t.AdditionalProperties = raw
	return nil
}

type ToolInputSchemaProperties struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
type ToolInputSchema struct {
	Type       string                               `json:"type"`
	Properties map[string]ToolInputSchemaProperties `json:"properties"`
	Required   []string                             `json:"required"`
}

func (ti *ToolInputSchema) MarshalJSON() ([]byte, error) {
	safeProperties := ti.Properties
	safeRequired := ti.Required
	safeType := ti.Type
	if safeProperties == nil {
		safeProperties = make(map[string]ToolInputSchemaProperties)
	}
	if safeRequired == nil {
		safeRequired = []string{}
	}
	if safeType == "" {
		safeType = "object"
	}
	aux := struct {
		Type       string                               `json:"type"`
		Properties map[string]ToolInputSchemaProperties `json:"properties"`
		Required   []string                             `json:"required"`
	}{
		Type:       safeType,
		Properties: safeProperties,
		Required:   safeRequired,
	}
	return json.Marshal(&aux)
}

type ToolOutputSchemaProperties struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
type ToolOutputSchema struct {
	Type       string                                `json:"type"`
	Properties map[string]ToolOutputSchemaProperties `json:"properties"`
	Required   []string                              `json:"required"`
}

func (to *ToolOutputSchema) MarshalJSON() ([]byte, error) {
	safeProperties := to.Properties
	safeRequired := to.Required
	safeType := to.Type
	if safeProperties == nil {
		safeProperties = make(map[string]ToolOutputSchemaProperties)
	}
	if safeRequired == nil {
		safeRequired = []string{}
	}
	if safeType == "" {
		safeType = "object"
	}
	aux := struct {
		Type       string                                `json:"type"`
		Properties map[string]ToolOutputSchemaProperties `json:"properties"`
		Required   []string                              `json:"required"`
	}{
		Type:       safeType,
		Properties: safeProperties,
		Required:   safeRequired,
	}
	return json.Marshal(&aux)
}

//Additional properties describing a Tool to clients.
//
//NOTE: all properties in ToolAnnotations are **hints**.
//They are not guaranteed to provide a faithful description of
//tool behavior (including descriptive properties like `title`).
//
//Clients should never make tool use decisions based on ToolAnnotations
//received from untrusted servers.
type ToolAnnotations struct {
	//A human-readable title for the tool.
	Title string `json:"title,omitempty"`
	//If true, the tool does not modify its environment.
	//
	//Default: false
	ReadOnlyHint *bool `json:"readOnlyHint,omitempty"`
	//If true, the tool may perform destructive updates to its environment.
	//If false, the tool performs only additive updates.
	//
	//(This property is meaningful only when `readOnlyHint == false`)
	//
	//Default: true
	DestructiveHint *bool `json:"destructiveHint,omitempty"`
	//If true, calling the tool repeatedly with the same arguments
	//will have no additional effect on the its environment.
	//
	//(This property is meaningful only when `readOnlyHint == false`)
	//
	//Default: false
	IdempotentHint *bool `json:"idempotentHint,omitempty"`
	//If true, this tool may interact with an "open world" of external
	//entities. If false, the tool's domain of interaction is closed.
	//For example, the world of a web search tool is open, whereas that
	//of a memory tool is not.
	//
	//Default: true
	OpenWorldHint *bool `json:"openWorldHint,omitempty"`
}

//Sent from the client to request a list of tools the server has.
//
//Only method: METHOD_REQUEST_LIST_TOOLS
type ListToolsRequest struct {
	PaginatedRequest
}

func (lt *ListToolsRequest) TypeOfClientRequest() int { return LIST_TOOLS_REQUEST_CLIENT_REQUEST_TYPE }
func (lt *ListToolsRequest) TypeOfRequestInterface() int {
	return LIST_TOOLS_REQUEST_REQUEST_INTERFACE_TYPE
}
func (lt *ListToolsRequest) GetRequest() Request {
	return lt.Request
}

func NewListToolsRequest(params *PaginatedRequestParams) *ListToolsRequest {
	tr := new(ListToolsRequest)
	tr.Method = methods.METHOD_REQUEST_LIST_TOOLS
	tr.Params = params
	return tr
}

//The server's response to a tools/list request from the client.
type ListToolsResult struct {
	PaginatedResult
	Tools []Tool `json:"tools"`
}

func (ltr *ListToolsResult) TypeOfServerResult() int { return LIST_TOOLS_RESULT_SERVER_RESULT_TYPE }
func (ltr *ListToolsResult) TypeOfResultInterface() int {
	return LIST_TOOLS_RESULT_RESULT_INTERFACE_TYPE
}

//The server's response to a tool call.
//
//Any errors that originate from the tool SHOULD be reported inside the result
//object, with `isError` set to true, _not_ as an MCP protocol-level error
//response. Otherwise, the LLM would not be able to see that an error occurred
//and self-correct.
//
//However, any errors in _finding_ the tool, an error indicating that the
//server does not support tool calls, or any other exceptional conditions,
//should be reported as an MCP error response.
type CallToolResult struct {
	Result
	//Could be TextContent/ImageContent/AudioContent/EmbeddedResource
	Content []Content `json:"content"`
	//An object containing structured tool output.
	//
	//If the Tool defines an outputSchema, this field MUST be present in the result, and contain a JSON object that matches the schema.
	StructuredContent map[string]interface{} `json:"structuredContent,omitempty"`
	//Whether the tool call ended in an error.
	//
	//If not set, this is assumed to be false (the call was successful).
	IsError *bool `json:"isError,omitempty"`
}

func (ctr *CallToolResult) TypeOfServerResult() int    { return CALL_TOOL_RESULT_SERVER_RESULT_TYPE }
func (ctr *CallToolResult) TypeOfResultInterface() int { return CALL_TOOL_RESULT_RESULT_INTERFACE_TYPE }

//Used by the client to invoke a tool provided by the server.
//
//Only method: METHOD_REQUEST_CALL_TOOLS
type CallToolRequest struct {
	Request
	Params CallToolRequestParams `json:"params"`
}

func (ct *CallToolRequest) TypeOfClientRequest() int { return CALL_TOOL_REQUEST_CLIENT_REQUEST_TYPE }
func (ct *CallToolRequest) TypeOfRequestInterface() int {
	return CALL_TOOL_REQUEST_REQUEST_INTERFACE_TYPE
}
func (ct *CallToolRequest) GetRequest() Request {
	return ct.Request
}

func NewCallToolRequest(params *CallToolRequestParams) *CallToolRequest {
	ctr := new(CallToolRequest)
	ctr.Method = methods.METHOD_REQUEST_CALL_TOOLS
	if params != nil {
		ctr.Params = *params
	}
	return ctr
}

type CallToolRequestParams struct {
	BaseRequestParams
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

func (ctrP *CallToolRequestParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	}{
		Name:      ctrP.Name,
		Arguments: ctrP.Arguments,
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
	baseExtra, err := ctrP.BaseRequestParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (ctrP *CallToolRequestParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	ctrP.Name = aux.Name
	ctrP.Arguments = aux.Arguments

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
	if err := ctrP.BaseRequestParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("baseRequestParams.UnmarshalJSON: %w", err)
	}
	return nil
}

//An optional notification from the server to the client, informing it that the list of tools it offers has changed. This may be issued by servers without any previous subscription from the client.
//
//Only method: METHOD_NOTIFICATION_TOOLS_LIST_CHANGED
type ToolListChangedNotification struct {
	Notification
}

func NewToolListChangedNotification(params *BaseNotificationParams) *ToolListChangedNotification {
	tlcn := new(ToolListChangedNotification)
	tlcn.Method = methods.METHOD_NOTIFICATION_TOOLS_LIST_CHANGED
	tlcn.Params = params
	return tlcn
}

func (tln *ToolListChangedNotification) TypeOfServerNotification() int {
	return TOOL_LIST_CHANGED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
}
func (tln *ToolListChangedNotification) TypeOfNotification() int {
	return TOOL_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
}
func (tln *ToolListChangedNotification) GetNotification() Notification {
	return tln.Notification
}
