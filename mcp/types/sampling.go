package types

import (
	"encoding/json"
	"fmt"

	"github.com/victorvbello/gomcp/mcp/methods"
)

//A request from the server to sample an LLM via the client. The client has full discretion over which model to select. The client should also inform the user before beginning sampling, to allow them to inspect the request (human in the loop) and decide whether to approve it.
//
//Only method: METHOD_SAMPLING_CREATE_MESSAGE
type CreateMessageRequest struct {
	Request
	CreateMessageParams `json:"params"`
}

func (cmr *CreateMessageRequest) TypeOfServerRequest() int {
	return CREATE_MESSAGE_REQUEST_SERVER_REQUEST_TYPE
}
func (cmr *CreateMessageRequest) TypeOfRequestInterface() int {
	return CREATE_MESSAGE_REQUEST_REQUEST_INTERFACE_TYPE
}
func (cmr *CreateMessageRequest) GetRequest() Request {
	return cmr.Request
}

func NewCreateMessageRequest(params *CreateMessageParams) *CreateMessageRequest {
	cmr := new(CreateMessageRequest)
	cmr.Method = methods.METHOD_SAMPLING_CREATE_MESSAGE
	if params != nil {
		cmr.CreateMessageParams = *params
	}
	return cmr
}

type CreateMessageResult struct {
	Result
	SamplingMessage
	Model string `json:"model"`
	//endTurn/stopSequence/maxTokens
	StopReason string `json:"stopReason,omitempty"`
}

func (cmr *CreateMessageResult) TypeOfClientResult() int {
	return CREATE_MESSAGE_RESULT_CLIENT_RESULT_TYPE
}
func (cmr *CreateMessageResult) TypeOfResultInterface() int {
	return CREATE_MESSAGE_RESULT_RESULT_INTERFACE_TYPE
}

type CreateMessageParams struct {
	BaseRequestParams
	Messages []SamplingMessage `json:"messages"`
	//The server's preferences for which model to select. The client MAY ignore these preferences.
	ModelPreferences *ModelPreferences `json:"modelPreferences,omitempty"`
	//An optional system prompt the server wants to use for sampling. The client MAY modify or omit this prompt.
	SystemPrompt string `json:"systemPrompt,omitempty"`
	//A request to include context from one or more MCP servers (including the caller), to be attached to the prompt. The client MAY ignore this request.
	//
	//Could be none/thisServer/allServers
	IncludeContext string  `json:"includeContext,omitempty"`
	Temperature    float64 `json:"temperature,omitempty"`
	//The maximum number of tokens to sample, as requested by the server. The client MAY choose to sample fewer tokens than requested.
	MaxTokens     int      `json:"maxTokens"`
	StopSequences []string `json:"stopSequences,omitempty"`
	//Optional metadata to pass through to the LLM provider. The format of this metadata is provider-specific.
	Metadata interface{} `json:"metadata,omitempty"`
}

//Describes a message issued to or received from an LLM API.
type SamplingMessage struct {
	Role Role `json:"role"`
	//Could be TextContent/ImageContent/AudioContent
	Content Content `json:"content"`
}

func (sm *SamplingMessage) UnmarshalJSON(data []byte) error {
	var meta struct {
		Role    Role            `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("error unmarshaling global meta: %v", err)
	}
	sm.Role = meta.Role
	contentDataMap := make(map[string]interface{})
	if err := json.Unmarshal(meta.Content, &contentDataMap); err != nil {
		return fmt.Errorf("error unmarshaling global data in map: %v", err)
	}

	var c Content
	var contentFactories = map[string]func() Content{
		"text":     func() Content { return new(TextContent) },
		"image":    func() Content { return new(ImageContent) },
		"audio":    func() Content { return new(AudioContent) },
		"resource": func() Content { return new(EmbeddedResource) },
	}

	for key, builder := range contentFactories {
		if _, ok := contentDataMap[key]; ok {
			c = builder()
			break
		}
	}
	if c == nil {
		c = new(TextContent)
	}

	if err := json.Unmarshal(meta.Content, &c); err != nil {
		return fmt.Errorf("error unmarshaling err: %v", err)
	}
	sm.Content = c
	return nil
}

//The server's preferences for model selection, requested of the client during sampling.
//
//Because LLMs can vary along multiple dimensions, choosing the "best" model is
//rarely straightforward.  Different models excel in different areas—some are
//faster but less capable, others are more capable but more expensive, and so
//on. This interface allows servers to express their priorities across multiple
//dimensions to help clients make an appropriate selection for their use case.
//
//These preferences are always advisory. The client MAY ignore them. It is also
//up to the client to decide how to interpret these preferences and how to
//balance them against other considerations.
type ModelPreferences struct {
	//Optional hints to use for model selection.
	//
	//If multiple hints are specified, the client MUST evaluate them in order
	//(such that the first match is taken).
	//
	//The client SHOULD prioritize these hints over the numeric priorities, but
	//MAY still use the priorities to select from ambiguous matches.
	Hints []ModelHint `json:"hints,omitempty"`
	//How much to prioritize cost when selecting a model. A value of 0 means cost
	//is not important, while a value of 1 means cost is the most important
	//factor.
	CostPriority float64 `json:"costPriority,omitempty"`
	//How much to prioritize sampling speed (latency) when selecting a model. A
	//value of 0 means speed is not important, while a value of 1 means speed is
	//the most important factor.
	SpeedPriority float64 `json:"speedPriority,omitempty"`
	//How much to prioritize intelligence and capabilities when selecting a
	//model. A value of 0 means intelligence is not important, while a value of 1
	//means intelligence is the most important factor.
	IntelligencePriority float64 `json:"intelligencePriority,omitempty"`
}

//Hints to use for model selection.
//
//Keys not declared here are currently left unspecified by the spec and are up
//to the client to interpret.
type ModelHint struct {
	//A hint for a model name.
	//
	//The client SHOULD treat this as a substring of a model name; for example:
	//- `claude-3-5-sonnet` should match `claude-3-5-sonnet-20241022`
	//- `sonnet` should match `claude-3-5-sonnet-20241022`, `claude-3-sonnet-20240229`, etc.
	//- `claude` should match any Claude model
	//
	//The client MAY also map the string to a different provider's model name or a different model family, as long as it fills a similar niche; for example:
	//- `gemini-1.5-flash` could match `claude-3-haiku-20240307`
	Name string `json:"name,omitempty"`
}
