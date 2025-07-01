package types

//A request from the client to the server, to ask for completion options.
//
//Only method: METHOD_AUTOCOMPLETE_COMPLETE
type CompleteRequest struct {
	Request
	Params CompleteParams `json:"params"`
}

func (c *CompleteRequest) TypeOfClientRequest() int { return COMPLETE_REQUEST_CLIENT_REQUEST_TYPE }

type AutoCompleteReference interface {
	AutoCompleteRefType() string
}

type CompleteParams struct {
	//Could be PromptReference/ResourceReference
	Ref AutoCompleteReference `json:"ref"`
	//The argument's information
	Argument struct {
		//The name of the argument
		Name string `json:"name"`
		//The value of the argument to use for completion matching.
		Value string `json:"value"`
	} `json:"argument"`
}

//The server's response to a completion/complete request
type CompleteResult struct {
	Result
	Completion struct {
		//An array of completion values. Must not exceed 100 items.
		Values []string `json:"values"`
		//The total number of completion options available. This can exceed the number of values actually sent in the response.
		Total int `json:"total,omitempty"`
		//Indicates whether there are additional completion options beyond those provided in the current response, even if the exact total is unknown.
		HasMore *bool `json:"hasMore,omitempty"`
	} `json:"completion"`
}

func (cr *CompleteResult) TypeOfServerResult() int    { return COMPLETE_RESULT_SERVER_RESULT_TYPE }
func (cr *CompleteResult) TypeOfResultInterface() int { return COMPLETE_RESULT_RESULT_INTERFACE_TYPE }
