package types

const (
	AUTOCOMPLETE_REF_PROMPT_TYPE = "ref/prompt"
)

//A prompt or prompt template that the server offers.
type Prompt struct {
	//The name of the prompt or prompt template.
	Name string `json:"name"`
	//An optional description of what this prompt provides
	Description string `json:"description,omitempty"`
	//A list of arguments to use for templating the prompt.
	Arguments []PromptArgument `json:"arguments,omitempty"`
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
func (tln *PromptListChangedNotification) TypeOfServerNotification() int {
	return PROMPT_LIST_CHANGED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
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

//Used by the client to get a prompt provided by the server.
//
//Only method: METHOD_REQUEST_GET_PROMPTS
type GetPromptRequest struct {
	Request
	Params GetPromptParams `json:"params"`
}

func (g *GetPromptRequest) TypeOfClientRequest() int { return GET_PROMPT_REQUEST_CLIENT_REQUEST_TYPE }

type GetPromptParams struct {
	//The name of the prompt or prompt template.
	Name string `json:"name"`
	//Arguments to use for templating the prompt.
	Arguments map[string]string `json:"arguments,omitempty"`
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
