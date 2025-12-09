package types

import (
	"encoding/json"
	"fmt"

	"github.com/victorvbello/gomcp/mcp/methods"
)

//A request from the client to the server, to ask for completion options.
//
//Only method: METHOD_AUTOCOMPLETE_COMPLETE
type CompleteRequest struct {
	Request
	Params CompleteParams `json:"params"`
}

func (c *CompleteRequest) TypeOfClientRequest() int { return COMPLETE_REQUEST_CLIENT_REQUEST_TYPE }
func (c *CompleteRequest) UnmarshalJSON(data []byte) error {
	var meta struct {
		Method string `json:"method"`
		Params struct {
			BaseRequestParams
			Ref      json.RawMessage        `json:"ref"`
			Argument CompleteParamsArgument `json:"argument"`
			Context  CompleteParamsContext  `json:"context"`
		} `json:"params"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("error unmarshaling global meta: %v", err)
	}
	c.Method = meta.Method
	c.Params.BaseRequestParams.Meta = meta.Params.BaseRequestParams.Meta
	c.Params.Argument = meta.Params.Argument
	c.Params.Context = meta.Params.Context

	refDataMap := make(map[string]interface{})
	if err := json.Unmarshal(meta.Params.Ref, &refDataMap); err != nil {
		return fmt.Errorf("error unmarshaling global data in map: %v", err)
	}
	var acr AutoCompleteReference
	switch refDataMap["type"] {
	case AUTOCOMPLETE_REF_PROMPT_TYPE:
		acr = new(PromptReference)
	case AUTOCOMPLETE_REF_RESOURCE_TYPE:
		acr = new(ResourceTemplateReference)
	}
	if err := json.Unmarshal(meta.Params.Ref, &acr); err != nil {
		return fmt.Errorf("error unmarshaling method: %s, err: %v", meta.Method, err)
	}
	c.Params.Ref = acr
	return nil
}

func NewCompleteRequest(params *CompleteParams) *CompleteRequest {
	ncr := new(CompleteRequest)
	ncr.Method = methods.METHOD_AUTOCOMPLETE_COMPLETE
	if params != nil {
		ncr.Params = *params
	}
	return ncr
}

type AutoCompleteReference interface {
	AutoCompleteRefType() string
}

type CompleteParamsArgument struct {
	//The name of the argument
	Name string `json:"name"`
	//The value of the argument to use for completion matching.
	Value string `json:"value"`
}

type CompleteParamsContext struct {
	Arguments map[string]string `json:"arguments"`
}

type CompleteParams struct {
	BaseRequestParams
	//Could be PromptReference/ResourceReference
	Ref AutoCompleteReference `json:"ref"`
	//The argument's information
	Argument CompleteParamsArgument `json:"argument"`
	Context  CompleteParamsContext  `json:"context"`
}

type CompleteResultCompletion struct {
	//An array of completion values. Must not exceed 100 items.
	Values []string `json:"values"`
	//The total number of completion options available. This can exceed the number of values actually sent in the response.
	Total int `json:"total,omitempty"`
	//Indicates whether there are additional completion options beyond those provided in the current response, even if the exact total is unknown.
	HasMore *bool `json:"hasMore,omitempty"`
}

//The server's response to a completion/complete request
type CompleteResult struct {
	Result
	Completion CompleteResultCompletion `json:"completion"`
}

func (cr *CompleteResult) TypeOfServerResult() int    { return COMPLETE_RESULT_SERVER_RESULT_TYPE }
func (cr *CompleteResult) TypeOfResultInterface() int { return COMPLETE_RESULT_RESULT_INTERFACE_TYPE }
