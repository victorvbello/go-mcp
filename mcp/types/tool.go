package types

//Definition for a tool the client can call.
type Tool struct {
	//The name of the tool.
	Name string `json:"name"`
	//A human-readable description of the tool.
	//
	//This can be used by clients to improve the LLM's understanding of available tools. It can be thought of like a "hint" to the model.
	Description string `json:"description,omitempty"`
	//A JSON Schema object defining the expected parameters for the tool.
	InputSchema ToolInputSchema `json:"inputSchema"`
	//Optional additional tool information.
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

type ToolInputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
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

type CallToolRequestParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

//An optional notification from the server to the client, informing it that the list of tools it offers has changed. This may be issued by servers without any previous subscription from the client.
//
//Only method: METHOD_NOTIFICATION_TOOLS_LIST_CHANGED
type ToolListChangedNotification struct {
	Notification
}

func (tln *ToolListChangedNotification) TypeOfServerNotification() int {
	return TOOL_LIST_CHANGED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
}
