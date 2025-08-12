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

func (cmr *CreateMessageRequest) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		Request
	}{
		Request: cmr.Request,
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
	//Marshal CreateMessageParams
	params, err := cmr.CreateMessageParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal CreateMessageParams fields: %w", err)
	}
	paramsMap := make(map[string]interface{})
	if err := json.Unmarshal(params, &paramsMap); err != nil {
		return nil, fmt.Errorf("unmarshal params fields: %w", err)
	}
	baseMap["params"] = paramsMap
	return json.Marshal(baseMap)
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

func (cmp *CreateMessageParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		Messages         []SamplingMessage `json:"messages"`
		ModelPreferences *ModelPreferences `json:"modelPreferences,omitempty"`
		SystemPrompt     string            `json:"systemPrompt,omitempty"`
		IncludeContext   string            `json:"includeContext,omitempty"`
		Temperature      float64           `json:"temperature,omitempty"`
		MaxTokens        int               `json:"maxTokens"`
		StopSequences    []string          `json:"stopSequences,omitempty"`
		Metadata         interface{}       `json:"metadata,omitempty"`
	}{
		Messages:         cmp.Messages,
		ModelPreferences: cmp.ModelPreferences,
		SystemPrompt:     cmp.SystemPrompt,
		IncludeContext:   cmp.IncludeContext,
		Temperature:      cmp.Temperature,
		MaxTokens:        cmp.MaxTokens,
		StopSequences:    cmp.StopSequences,
		Metadata:         cmp.Metadata,
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
	baseExtra, err := cmp.BaseRequestParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (cmp *CreateMessageParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		Messages         []SamplingMessage `json:"messages"`
		ModelPreferences *ModelPreferences `json:"modelPreferences,omitempty"`
		SystemPrompt     string            `json:"systemPrompt,omitempty"`
		IncludeContext   string            `json:"includeContext,omitempty"`
		Temperature      float64           `json:"temperature,omitempty"`
		MaxTokens        int               `json:"maxTokens"`
		StopSequences    []string          `json:"stopSequences,omitempty"`
		Metadata         interface{}       `json:"metadata,omitempty"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	aux.Messages = cmp.Messages
	aux.ModelPreferences = cmp.ModelPreferences
	aux.SystemPrompt = cmp.SystemPrompt
	aux.IncludeContext = cmp.IncludeContext
	aux.Temperature = cmp.Temperature
	aux.MaxTokens = cmp.MaxTokens
	aux.StopSequences = cmp.StopSequences
	aux.Metadata = cmp.Metadata

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "messages")
	delete(raw, "modelPreferences")
	delete(raw, "systemPrompt")
	delete(raw, "includeContext")
	delete(raw, "temperature")
	delete(raw, "maxTokens")
	delete(raw, "stopSequences")
	delete(raw, "metadata")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := cmp.BaseRequestParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("baseRequestParams.UnmarshalJSON: %w", err)
	}
	return nil
}

//Describes a message issued to or received from an LLM API.
type SamplingMessage struct {
	Role Role `json:"role"`
	//Could be TextContent/ImageContent/AudioContent
	Content Content `json:"content"`
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
