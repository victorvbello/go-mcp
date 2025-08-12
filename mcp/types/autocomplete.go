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

func (cp *CompleteParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		Ref      AutoCompleteReference  `json:"ref"`
		Argument CompleteParamsArgument `json:"argument"`
	}{
		Ref:      cp.Ref,
		Argument: cp.Argument,
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
	baseExtra, err := cp.BaseRequestParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (cp *CompleteParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		Ref      AutoCompleteReference  `json:"ref"`
		Argument CompleteParamsArgument `json:"argument"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	cp.Ref = aux.Ref
	cp.Argument = aux.Argument

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "ref")
	delete(raw, "argument")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := cp.BaseRequestParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("baseRequestParams.UnmarshalJSON: %w", err)
	}
	return nil
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
