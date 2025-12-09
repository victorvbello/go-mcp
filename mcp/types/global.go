package types

const (
	LATEST_PROTOCOL_VERSION             = "2025-03-26"
	DEFAULT_NEGOTIATED_PROTOCOL_VERSION = "2025-03-26"
)

var SUPPORTED_PROTOCOL_VERSIONS = map[string]struct{}{
	LATEST_PROTOCOL_VERSION: struct{}{},
	"2024-11-05":            struct{}{},
	"2024-10-07":            struct{}{},
}

//A progress token, used to associate progress notifications with the original request, string/number.
type ProgressToken interface{}

//An opaque token used to represent a cursor for pagination.
type Cursor string

//The sender or recipient of messages and data in a conversation, user/assistant
type Role string

type Meta map[string]interface{}

// GetMeta returns metadata from a value.
func (m Meta) GetMeta() map[string]interface{} { return m }

// SetMeta sets the metadata on a value.
func (m *Meta) SetMeta(x map[string]interface{}) { *m = x }

//Base metadata interface for common properties across resources, tools, prompts, and implementations.
type BaseMetadata struct {
	// Intended for programmatic or logical use, but used as a display name in past specs or fallback
	Name string `json:"name"`
	//Intended for UI and end-user contexts — optimized to be human-readable and easily understood,
	//even by those unfamiliar with domain-specific terminology.
	//
	//If not provided, the name should be used for display (except for Tool,
	//where `annotations.title` should be given precedence over using `name`,
	//if present).
	Title string `json:"title"`
}

//Describes the name and version of an MCP implementation.
type Implementation struct {
	BaseMetadata
	Version string `json:"version"`
}
