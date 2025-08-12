package methods

const (
	METHOD_NOTIFICATION_CANCELLED   = "notifications/cancelled"
	METHOD_NOTIFICATION_INITIALIZED = "notifications/initialized"
	METHOD_NOTIFICATION_PROGRESS    = "notifications/progress"
	METHOD_NOTIFICATION_MESSAGE     = "notifications/message"
)

var NOTIFICATION_METHODS = map[string]struct{}{
	METHOD_NOTIFICATION_CANCELLED:              struct{}{},
	METHOD_NOTIFICATION_INITIALIZED:            struct{}{},
	METHOD_NOTIFICATION_PROGRESS:               struct{}{},
	METHOD_NOTIFICATION_RESOURCES_LIST_CHANGED: struct{}{},
	METHOD_NOTIFICATION_RESOURCES_UPDATED:      struct{}{},
	METHOD_NOTIFICATION_PROMPTS_LIST_CHANGED:   struct{}{},
	METHOD_NOTIFICATION_TOOLS_LIST_CHANGED:     struct{}{},
	METHOD_NOTIFICATION_MESSAGE:                struct{}{},
	METHOD_NOTIFICATION_ROOTS_LIST_CHANGED:     struct{}{},
}
