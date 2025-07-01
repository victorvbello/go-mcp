package types

const (
	TEXT_RESOURCE_CONTENTS_TYPE = iota + 50
	BLOB_RESOURCE_CONTENTS_TYPE
)

const (
	AUTOCOMPLETE_REF_RESOURCE_TYPE = "ref/resource"
)

//A known resource that the server is capable of reading.
type Resource struct {
	//The URI of this resource.
	URI string `json:"uri"`
	//A human-readable name for this resource.
	//
	//This can be used by clients to populate UI elements.
	Name string `json:"name"`
	//A description of what this resource represents.
	//
	//This can be used by clients to improve the LLM's understanding of available resources. It can be thought of like a "hint" to the model.
	Description string `json:"description,omitempty"`
	//The MIME type of this resource, if known.
	MIMEType string `json:"mimeType,omitempty"`
	//Optional annotations for the client.
	Annotations *Annotations `json:"annotations,omitempty"`
	//The size of the raw resource content, in bytes (i.e., before base64 encoding or any tokenization), if known.
	//
	//This can be used by Hosts to display file sizes and estimate context window usage.
	Size int `json:"size,omitempty"`
}

//Optional annotations for the client. The client can use annotations to inform how objects are used or displayed
type Annotations struct {
	//Describes who the intended customer of this object or data is.
	//
	//It can include multiple entries to indicate content useful for multiple audiences (e.g., `["user", "assistant"]`).
	Audience []Role `json:"audience,omitempty"`
	//Describes how important this data is for operating the server.
	//
	//A value of 1 means "most important," and indicates that the data is
	//effectively required, while 0 means "least important," and indicates that
	//the data is entirely optional.
	Priority int `json:"priority,omitempty"`
}

//A template description for resources available on the server.
type ResourceTemplate struct {
	//A URI template (according to RFC 6570) that can be used to construct resource URIs.
	URITemplate string `json:"uriTemplate"`
	//A human-readable name for the type of resource this template refers to.
	//
	//This can be used by clients to populate UI elements.
	Name string `json:"name"`
	//A description of what this template is for.
	//
	//This can be used by clients to improve the LLM's understanding of available resources. It can be thought of like a "hint" to the model.
	Description string `json:"description,omitempty"`
	//The MIME type for all resources that match this template. This should only be included if all resources matching this template have the same type.
	MIMEType string `json:"mimeType,omitempty"`
	//Optional annotations for the client.
	Annotations *Annotations `json:"annotations,omitempty"`
}

//The contents of a specific resource or sub-resource.
type ResourceContents interface {
	TypeOfResource() int
}

//The base struct for ResourceContents
type BaseResourceContents struct {
	//The URI of this resource.
	URI string `json:"uri"`
	//The MIME type of this resource, if known.
	MIMEType string `json:"mimeType,omitempty"`
}

type TextResourceContents struct {
	BaseResourceContents
	//The text of the item. This must only be set if the item can actually be represented as text (not binary data).
	Text string `json:"text"`
}

func (TextResourceContents) TypeOfResource() int { return TEXT_RESOURCE_CONTENTS_TYPE }

type BlobResourceContents struct {
	BaseResourceContents
	//A base64-encoded string representing the binary data of the item.
	Blob string `json:"blob"`
}

func (BlobResourceContents) TypeOfResource() int { return BLOB_RESOURCE_CONTENTS_TYPE }

//An optional notification from the server to the client, informing it that the list of resources it can read from has changed. This may be issued by servers without any previous subscription from the client.
//
//Only method: METHOD_NOTIFICATION_RESOURCES_LIST_CHANGED
type ResourceListChangedNotification struct {
	Notification
}

func (rln *ResourceListChangedNotification) TypeOfServerNotification() int {
	return RESOURCE_LIST_CHANGED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
}

//A notification from the server to the client, informing it that a resource has changed and may need to be read again. This should only be sent if the client previously sent a resources/subscribe request.
//
//Only method: METHOD_NOTIFICATION_RESOURCES_UPDATED
type ResourceUpdatedNotification struct {
	Notification
	Params ResourceUpdatedNotificationParams `json:"params"`
}

func (run *ResourceUpdatedNotification) TypeOfServerNotification() int {
	return RESOURCE_UPDATED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
}

type ResourceUpdatedNotificationParams struct {
	//The URI of the resource that has been updated. This might be a sub-resource of the one that the client actually subscribed to.
	URI string `json:"uri"`
}

//Sent from the client to request a list of resources the server has.
//
//Only method: METHOD_REQUEST_LIST_RESOURCES
type ListResourcesRequest struct {
	PaginatedRequest
}

func (lr *ListResourcesRequest) TypeOfClientRequest() int {
	return LIST_RESOURCES_REQUEST_CLIENT_REQUEST_TYPE
}

//Sent from the client to request a list of resource templates the server has.
//
//Only method: METHOD_REQUEST_TEMPLATES_LIST_RESOURCES
type ListResourceTemplatesRequest struct {
	PaginatedRequest
}

func (lt *ListResourceTemplatesRequest) TypeOfClientRequest() int {
	return LIST_RESOURCE_TEMPLATES_REQUEST_CLIENT_REQUEST_TYPE
}

//Sent from the client to the server, to read a specific resource URI.
//
//Only method: METHOD_REQUEST_READ_RESOURCES
type ReadResourceRequest struct {
	Request
	Params ReadResourceRequestParams `json:"params"`
}

func (rr *ReadResourceRequest) TypeOfClientRequest() int {
	return READ_RESOURCE_REQUEST_CLIENT_REQUEST_TYPE
}

type ReadResourceRequestParams struct {
	//The URI of the resource to read. The URI can use any protocol; it is up to the server how to interpret it.
	URI string `json:"uri"`
}

//The server's response to a resources/list request from the client.
type ListResourcesResult struct {
	PaginatedResult
	Resources []Resource `json:"resources"`
}

func (lrr *ListResourcesResult) TypeOfServerResult() int {
	return LIST_RESOURCES_RESULT_SERVER_RESULT_TYPE
}
func (lrr *ListResourcesResult) TypeOfResultInterface() int {
	return LIST_RESOURCES_RESULT_RESULT_INTERFACE_TYPE
}

//The server's response to a resources/templates/list request from the client.
type ListResourceTemplatesResult struct {
	PaginatedResult
	ResourceTemplates []ResourceTemplate `json:"resourceTemplates"`
}

func (lrt *ListResourceTemplatesResult) TypeOfServerResult() int {
	return LIST_RESOURCE_TEMPLATES_RESULT_SERVER_RESULT_TYPE
}

func (lrt *ListResourceTemplatesResult) TypeOfResultInterface() int {
	return LIST_RESOURCE_TEMPLATES_RESULT_RESULT_INTERFACE_TYPE
}

//The server's response to a resources/read request from the client.
type ReadResourceResult struct {
	Result
	//Could be TextResourceContents/BlobResourceContents
	Contents []ResourceContents `json:"contents"`
}

func (rrr *ReadResourceResult) TypeOfServerResult() int {
	return READ_RESOURCE_RESULT_SERVER_RESULT_TYPE
}
func (rrr *ReadResourceResult) TypeOfResultInterface() int {
	return READ_RESOURCE_RESULT_RESULT_INTERFACE_TYPE
}

//A reference to a resource or resource template definition.
type ResourceReference struct {
	//Only AUTOCOMPLETE_REF_RESOURCE_TYPE
	Type string `json:"type"`
	//The URI or URI template of the resource.
	URI string `json:"uri"`
}

func (r *ResourceReference) AutoCompleteRefType() string { return AUTOCOMPLETE_REF_RESOURCE_TYPE }
