package types

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
func (ep *EmptyResult) GetResult() Result          { return Result(*ep) }

type Result struct {
	//Attach additional metadata to their notifications.
	Meta `json:"_meta,omitempty"`
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

func (sc *ServerCapabilities) UpdateAll(new *ServerCapabilities) {
	capUpdaters := []func(dst, src *ServerCapabilities){
		func(dst, src *ServerCapabilities) {
			if src.Experimental != nil {
				dst.Experimental = src.Experimental
			}
		},
		func(dst, src *ServerCapabilities) {
			if src.Logging != nil {
				dst.Logging = src.Logging
			}
		},
		func(dst, src *ServerCapabilities) {
			if src.Completions != nil {
				dst.Completions = src.Completions
			}
		},
		func(dst, src *ServerCapabilities) {
			if src.Sampling != nil {
				dst.Sampling = src.Sampling
			}
		},
		func(dst, src *ServerCapabilities) {
			if src.Prompts != nil {
				dst.Prompts = src.Prompts
			}
		},
		func(dst, src *ServerCapabilities) {
			if src.Resources != nil {
				dst.Resources = src.Resources
			}
		},
		func(dst, src *ServerCapabilities) {
			if src.Tools != nil {
				dst.Tools = src.Tools
			}
		},
	}
	for _, update := range capUpdaters {
		update(sc, new)
	}
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
func (ir *InitializeResult) GetResult() Result { return ir.Result }

type PaginatedResult struct {
	Result
	//An opaque token representing the pagination position after the last returned result.
	//
	//If present, there may be more results available.
	NextCursor Cursor `json:"nextCursor,omitempty"`
}

func (pr *PaginatedResult) TypeOfResultInterface() int { return PAGINATED_RESULT_RESULT_INTERFACE_TYPE }
func (pr *PaginatedResult) GetResult() Result          { return pr.Result }

type ResultInterface interface {
	TypeOfResultInterface() int
	GetResult() Result
}

type ClientResult interface {
	TypeOfClientResult() int
}

type ServerResult interface {
	TypeOfServerResult() int
}

//Configuration for list changed notification handlers.
//
//Use this to configure handlers for tools, prompts, and resources list changes
//when creating a client.
//
//Note: Handlers are only activated if the server advertises the corresponding
//`listChanged` capability (e.g., `tools.listChanged: true`). If the server
//doesn't advertise this capability, the handler will not be set up.
type ClientCapabilitiesListChangedHandlers struct {
	//Handler for tool list changes.
	Tools *ClientCapabilitiesListChangedToolsOptions
	//Handler for prompt list changes.
	Prompts *ClientCapabilitiesListChangedPromptOptions
	//Handler for resource list changes.
	Resources *ClientCapabilitiesListChangedResourcesOptions
}
type ClientCapabilitiesListChangedToolsCallback = func(items []Tool, err error)

//Options for subscribing to list changed notifications tools
type ClientCapabilitiesListChangedToolsOptions struct {
	//If true, the list will be refreshed automatically when a list changed notification is received.
	//default true
	AutoRefresh bool
	//Debounce time in milliseconds. Set to 0 to disable.
	//@default 300
	DebounceMs int
	//Callback invoked when the list changes.

	//If autoRefresh is true, items contains the updated list.
	//If autoRefresh is false, items is null (caller should refresh manually).
	OnChanged ClientCapabilitiesListChangedToolsCallback
}

type ClientCapabilitiesListChangedPromptCallback = func(items []Tool, err error)

//Options for subscribing to list changed notifications prompts
type ClientCapabilitiesListChangedPromptOptions struct {
	//If true, the list will be refreshed automatically when a list changed notification is received.
	//default true
	AutoRefresh bool
	//Debounce time in milliseconds. Set to 0 to disable.
	//@default 300
	DebounceMs int
	//Callback invoked when the list changes.

	//If autoRefresh is true, items contains the updated list.
	//If autoRefresh is false, items is null (caller should refresh manually).
	OnChanged ClientCapabilitiesListChangedPromptCallback
}

type ClientCapabilitiesListChangedResourcesCallback = func(items []Tool, err error)

//Options for subscribing to list changed notifications resources
type ClientCapabilitiesListChangedResourcesOptions struct {
	//If true, the list will be refreshed automatically when a list changed notification is received.
	//default true
	AutoRefresh bool
	//Debounce time in milliseconds. Set to 0 to disable.
	//@default 300
	DebounceMs int
	//Callback invoked when the list changes.

	//If autoRefresh is true, items contains the updated list.
	//If autoRefresh is false, items is null (caller should refresh manually).
	OnChanged ClientCapabilitiesListChangedResourcesCallback
}
