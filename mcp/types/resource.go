package types

import (
	"github.com/victorvbello/gomcp/mcp/methods"
	"github.com/victorvbello/gomcp/mcp/utils"
)

const (
	TEXT_RESOURCE_CONTENTS_TYPE = iota + 50
	BLOB_RESOURCE_CONTENTS_TYPE
)

const (
	AUTOCOMPLETE_REF_RESOURCE_TYPE = "ref/resource"
)

//A known resource that the server is capable of reading.
type Resource struct {
	BaseMetadata
	//The URI of this resource.
	URI string `json:"uri"`
	//A description of what this resource represents.
	//
	//This can be used by clients to improve the LLM's understanding of available resources. It can be thought of like a "hint" to the model.
	Description string `json:"description,omitempty"`
	//The MIME type of this resource, if known.
	MIMEType string `json:"mimeType,omitempty"`
	//See [MCP specification](https://github.com/modelcontextprotocol/modelcontextprotocol/blob/47339c03c143bb4ec01a26e721a1b8fe66634ebe/docs/specification/draft/basic/index.mdx#general-fields)
	//for notes on _meta usage.
	Meta `json:"_meta,omitempty"`
}

//A template description for resources available on the server.
type ResourceTemplate struct {
	BaseMetadata
	//A URI template (according to RFC 6570) that can be used to construct resource URIs.
	URITemplate utils.UriTemplate `json:"uriTemplate"`
	//A description of what this template is for.
	//
	//This can be used by clients to improve the LLM's understanding of available resources. It can be thought of like a "hint" to the model.
	Description string `json:"description,omitempty"`
	//The MIME type for all resources that match this template. This should only be included if all resources matching this template have the same type.
	MIMEType string `json:"mimeType,omitempty"`
	//See [MCP specification](https://github.com/modelcontextprotocol/modelcontextprotocol/blob/47339c03c143bb4ec01a26e721a1b8fe66634ebe/docs/specification/draft/basic/index.mdx#general-fields)
	//for notes on _meta usage.
	Meta `json:"_meta,omitempty"`
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

func NewResourceListChangedNotification(params *BaseNotificationParams) *ResourceListChangedNotification {
	rlcn := new(ResourceListChangedNotification)
	rlcn.Method = methods.METHOD_NOTIFICATION_RESOURCES_LIST_CHANGED
	rlcn.Params = params
	return rlcn
}

func (rln *ResourceListChangedNotification) TypeOfServerNotification() int {
	return RESOURCE_LIST_CHANGED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
}
func (rln *ResourceListChangedNotification) TypeOfNotification() int {
	return RESOURCE_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
}
func (rln *ResourceListChangedNotification) GetNotification() Notification {
	return rln.Notification
}

//A notification from the server to the client, informing it that a resource has changed and may need to be read again. This should only be sent if the client previously sent a resources/subscribe request.
//
//Only method: METHOD_NOTIFICATION_RESOURCES_UPDATED
type ResourceUpdatedNotification struct {
	Notification
	Params ResourceUpdatedNotificationParams `json:"params"`
}

func NewResourceUpdatedNotification(params *ResourceUpdatedNotificationParams) *ResourceUpdatedNotification {
	run := new(ResourceUpdatedNotification)
	run.Method = methods.METHOD_NOTIFICATION_RESOURCES_UPDATED
	if params != nil {
		run.Params = *params
	}
	return run
}

func (run *ResourceUpdatedNotification) TypeOfServerNotification() int {
	return RESOURCE_UPDATED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
}
func (run *ResourceUpdatedNotification) TypeOfNotification() int {
	return RESOURCE_UPDATED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
}
func (run *ResourceUpdatedNotification) GetNotification() Notification {
	return run.Notification
}

type ResourceUpdatedNotificationParams struct {
	BaseNotificationParams
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
func (lr *ListResourcesRequest) TypeOfRequestInterface() int {
	return LIST_RESOURCES_REQUEST_REQUEST_INTERFACE_TYPE
}
func (lr *ListResourcesRequest) GetRequest() Request {
	return lr.Request
}

func NewListResourcesRequest(params *PaginatedRequestParams) *ListResourcesRequest {
	lrr := new(ListResourcesRequest)
	lrr.Method = methods.METHOD_REQUEST_LIST_RESOURCES
	lrr.Params = params
	return lrr
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
func (lt *ListResourceTemplatesRequest) TypeOfRequestInterface() int {
	return LIST_RESOURCE_TEMPLATES_REQUEST_REQUEST_INTERFACE_TYPE
}
func (lt *ListResourceTemplatesRequest) GetRequest() Request {
	return lt.Request
}

func NewListResourceTemplatesRequest(params *PaginatedRequestParams) *ListResourceTemplatesRequest {
	lrtr := new(ListResourceTemplatesRequest)
	lrtr.Method = methods.METHOD_REQUEST_TEMPLATES_LIST_RESOURCES
	lrtr.Params = params
	return lrtr
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
func (rr *ReadResourceRequest) TypeOfRequestInterface() int {
	return READ_RESOURCE_REQUEST_REQUEST_INTERFACE_TYPE
}
func (rr *ReadResourceRequest) GetRequest() Request {
	return rr.Request
}

func NewReadResourceRequest(params *ReadResourceRequestParams) *ReadResourceRequest {
	rrr := new(ReadResourceRequest)
	rrr.Method = methods.METHOD_REQUEST_READ_RESOURCES
	if params != nil {
		rrr.Params = *params
	}
	return rrr
}

type ReadResourceRequestParams struct {
	BaseRequestParams
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
type ResourceTemplateReference struct {
	//Only AUTOCOMPLETE_REF_RESOURCE_TYPE
	Type string `json:"type"`
	//The URI or URI template of the resource.
	URI string `json:"uri"`
}

func (r *ResourceTemplateReference) AutoCompleteRefType() string {
	return AUTOCOMPLETE_REF_RESOURCE_TYPE
}

func NewResourceTemplateReference(uri string) *ResourceTemplateReference {
	npr := ResourceTemplateReference{
		Type: AUTOCOMPLETE_REF_RESOURCE_TYPE,
		URI:  uri,
	}
	return &npr
}
