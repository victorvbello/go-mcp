package types

const (
	LATEST_PROTOCOL_VERSION = "2025-03-26"
)

//A progress token, used to associate progress notifications with the original request, string/number.
type ProgressToken interface{}

type Metadata map[string]interface{}

//An opaque token used to represent a cursor for pagination.
type Cursor string

//The sender or recipient of messages and data in a conversation, user/assistant
type Role string

//Describes the name and version of an MCP implementation.
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
