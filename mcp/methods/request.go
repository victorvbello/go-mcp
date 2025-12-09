package methods

const (
	METHOD_REQUEST_INITIALIZE               = "initialize"
	METHOD_REQUEST_PING                     = "ping"
	METHOD_REQUEST_LIST_RESOURCES           = "resources/list"
	METHOD_REQUEST_TEMPLATES_LIST_RESOURCES = "resources/templates/list"
	METHOD_REQUEST_READ_RESOURCES           = "resources/read"
	METHOD_REQUEST_SUBSCRIBE_RESOURCES      = "resources/subscribe"
	METHOD_REQUEST_UNSUBSCRIBE_RESOURCES    = "resources/unsubscribe"
	METHOD_REQUEST_LIST_PROMPTS             = "prompts/list"
	METHOD_REQUEST_GET_PROMPTS              = "prompts/get"
	METHOD_REQUEST_LIST_TOOLS               = "tools/list"
	METHOD_REQUEST_CALL_TOOLS               = "tools/call"
	METHOD_REQUEST_SET_LEVEL_LOGGING        = "logging/setLevel"
	METHOD_SAMPLING_CREATE_MESSAGE          = "sampling/createMessage"
	METHOD_AUTOCOMPLETE_COMPLETE            = "completion/complete"
	METHOD_LIST_ROOTS                       = "roots/list"
)

var REQUEST_METHODS = map[string]struct{}{
	METHOD_REQUEST_INITIALIZE:               struct{}{},
	METHOD_REQUEST_PING:                     struct{}{},
	METHOD_REQUEST_LIST_RESOURCES:           struct{}{},
	METHOD_REQUEST_TEMPLATES_LIST_RESOURCES: struct{}{},
	METHOD_REQUEST_READ_RESOURCES:           struct{}{},
	METHOD_REQUEST_SUBSCRIBE_RESOURCES:      struct{}{},
	METHOD_REQUEST_UNSUBSCRIBE_RESOURCES:    struct{}{},
	METHOD_REQUEST_LIST_PROMPTS:             struct{}{},
	METHOD_REQUEST_GET_PROMPTS:              struct{}{},
	METHOD_REQUEST_LIST_TOOLS:               struct{}{},
	METHOD_REQUEST_CALL_TOOLS:               struct{}{},
	METHOD_REQUEST_SET_LEVEL_LOGGING:        struct{}{},
	METHOD_SAMPLING_CREATE_MESSAGE:          struct{}{},
	METHOD_AUTOCOMPLETE_COMPLETE:            struct{}{},
	METHOD_LIST_ROOTS:                       struct{}{},
}
