package server

import (
	"sync"

	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	"github.com/victorvbello/gomcp/mcp/utils"
)

//Callback to read a resource at a given URI.
type ReadResourceCallback = func(uri string, extra *shared.RequestHandlerExtra) (*types.ReadResourceResult, error)

type ResourceMetadata struct {
	types.Resource
}

type RegisteredResourceUpdateOpts struct {
	Name     string
	Title    string
	URI      string
	Metadata *ResourceMetadata
	Callback ReadResourceCallback
	Enabled  bool
}

type RegisteredResource struct {
	Name         string
	Title        string
	Metadata     *ResourceMetadata
	ReadCallback ReadResourceCallback
	Enabled      bool
	Enable       func()
	Disable      func()
	Update       func(updates RegisteredResourceUpdateOpts) error
	Remove       func()
}

//Callback to list all resources matching a given template.
type ListResourcesCallback func(extra *shared.RequestHandlerExtra) (*types.ListResourcesResult, error)

//A callback to complete one variable within a resource template's URI template.
type CompleteResourceTemplateCallback func(value string, context types.CompleteParamsContext) ([]string, error)

type ResourceTemplateCallbacks struct {
	//A callback to list all resources matching this template. This is required to specified, even if `undefined`, to avoid accidentally forgetting resource listing.
	List ListResourcesCallback
	//An optional callback to autocomplete variables within the URI template. Useful for clients and users to discover possible values.
	Complete map[string]CompleteResourceTemplateCallback
}

//A resource template combines a URI pattern with optional functionality to enumerate
//all resources matching that pattern.
type ResourceTemplate struct {
	uriTemplate utils.UriTemplate
	callbacks   ResourceTemplateCallbacks
}

//Gets the URI template pattern.
func (rt *ResourceTemplate) GetUriTemplate() utils.UriTemplate {
	return rt.uriTemplate
}

//Gets the list callback, if one was provided.
func (rt *ResourceTemplate) GetListCallback() ListResourcesCallback {
	return rt.callbacks.List
}

//Gets the callback for completing a specific URI template variable, if one was provided.
func (rt *ResourceTemplate) CompleteCallback(variable string) CompleteResourceTemplateCallback {
	callBack := rt.callbacks.Complete[variable]
	return callBack
}

func NewResourceTemplate(uriTemplate utils.UriTemplate, callbacks ResourceTemplateCallbacks) *ResourceTemplate {
	nrt := new(ResourceTemplate)
	nrt.uriTemplate = uriTemplate
	nrt.callbacks = callbacks
	return nrt
}

//Callback to read a resource at a given URI, following a filled-in URI template.
type ReadResourceTemplateCallback = func(uri string, variables utils.UriVariables, extra *shared.RequestHandlerExtra) (*types.ReadResourceResult, error)

type RegisteredResourceTemplateUpdateOpts struct {
	Name     string
	Title    string
	Template *ResourceTemplate
	Metadata *ResourceMetadata
	Callback ReadResourceTemplateCallback
	Enabled  bool
}

type RegisteredResourceTemplate struct {
	ResourceTemplate ResourceTemplate
	Title            string
	Metadata         *ResourceMetadata
	ReadCallback     ReadResourceTemplateCallback
	Enabled          bool
	Enable           func()
	Disable          func()
	Update           func(updates RegisteredResourceTemplateUpdateOpts) error
	Remove           func()
}

type ToolCallback func(args map[string]interface{}, extra *shared.RequestHandlerExtra) (*types.CallToolResult, error)

type RegisteredToolUpdateOpts struct {
	Name         string
	Title        string
	Description  string
	ParamsSchema types.ToolInputSchema
	OutputSchema types.ToolOutputSchema
	Annotations  *types.ToolAnnotations
	Callback     ToolCallback
	Enabled      bool
}

type RegisteredTool struct {
	Title        string
	Description  string
	InputSchema  types.ToolInputSchema
	OutputSchema types.ToolOutputSchema
	Annotations  *types.ToolAnnotations
	Callback     ToolCallback
	Enabled      bool
	Enable       func()
	Disable      func()
	Update       func(updates RegisteredToolUpdateOpts) error
	Remove       func()
}

type PromptCallback func(args map[string]string, extra *shared.RequestHandlerExtra) (*types.GetPromptResult, error)

type RegisteredPromptUpdateOpts struct {
	Name        string
	Title       string
	Description string
	ArgsSchema  map[string]PromptArgsSchemaField
	Callback    PromptCallback
	Enabled     bool
}

type PromptArgsSchemaField struct {
	Description string
	Complete    func(values string, ctx types.CompleteParamsContext) []string
	IsOptional  bool
}

type RegisteredPrompt struct {
	Title       string
	Description string
	ArgsSchema  map[string]PromptArgsSchemaField
	Callback    PromptCallback
	Enabled     bool
	Enable      func()
	Disable     func()
	Update      func(updates RegisteredPromptUpdateOpts) error
	Remove      func()
}

//muxMapRegisteredResource
type muxMapRegisteredResource struct {
	mu sync.RWMutex
	m  map[string]RegisteredResource
}

func newMuxMapRegisteredResource() *muxMapRegisteredResource {
	return &muxMapRegisteredResource{
		m: make(map[string]RegisteredResource),
	}
}

func (xm *muxMapRegisteredResource) Clear() {
	xm.mu.Lock()
	xm.m = make(map[string]RegisteredResource)
	xm.mu.Unlock()
}

func (xm *muxMapRegisteredResource) Get(key string) (RegisteredResource, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapRegisteredResource) GetAll() map[string]RegisteredResource {
	xm.mu.RLock()
	clonedMap := make(map[string]RegisteredResource)
	for key, value := range xm.m {
		clonedMap[key] = value
	}
	xm.mu.RUnlock()
	return clonedMap
}

func (xm *muxMapRegisteredResource) Set(key string, value RegisteredResource) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}
func (xm *muxMapRegisteredResource) Delete(key string) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}

//muxMapRegisteredResourceTemplate
type muxMapRegisteredResourceTemplate struct {
	mu sync.RWMutex
	m  map[string]RegisteredResourceTemplate
}

func newMuxMapRegisteredResourceTemplate() *muxMapRegisteredResourceTemplate {
	return &muxMapRegisteredResourceTemplate{
		m: make(map[string]RegisteredResourceTemplate),
	}
}

func (xm *muxMapRegisteredResourceTemplate) Clear() {
	xm.mu.Lock()
	xm.m = make(map[string]RegisteredResourceTemplate)
	xm.mu.Unlock()
}

func (xm *muxMapRegisteredResourceTemplate) Get(key string) (RegisteredResourceTemplate, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapRegisteredResourceTemplate) GetAll() map[string]RegisteredResourceTemplate {
	xm.mu.RLock()
	clonedMap := make(map[string]RegisteredResourceTemplate)
	for key, value := range xm.m {
		clonedMap[key] = value
	}
	xm.mu.RUnlock()
	return clonedMap
}

func (xm *muxMapRegisteredResourceTemplate) Set(key string, value RegisteredResourceTemplate) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}
func (xm *muxMapRegisteredResourceTemplate) Delete(key string) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}

//muxMapRegisteredTool
type muxMapRegisteredTool struct {
	mu sync.RWMutex
	m  map[string]RegisteredTool
}

func newMuxMapRegisteredTool() *muxMapRegisteredTool {
	return &muxMapRegisteredTool{
		m: make(map[string]RegisteredTool),
	}
}

func (xm *muxMapRegisteredTool) Clear() {
	xm.mu.Lock()
	xm.m = make(map[string]RegisteredTool)
	xm.mu.Unlock()
}

func (xm *muxMapRegisteredTool) Get(key string) (RegisteredTool, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapRegisteredTool) GetAll() map[string]RegisteredTool {
	xm.mu.RLock()
	clonedMap := make(map[string]RegisteredTool)
	for key, value := range xm.m {
		clonedMap[key] = value
	}
	xm.mu.RUnlock()
	return clonedMap
}

func (xm *muxMapRegisteredTool) Set(key string, value RegisteredTool) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}
func (xm *muxMapRegisteredTool) Delete(key string) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}

//muxMapRegisteredPrompt
type muxMapRegisteredPrompt struct {
	mu sync.RWMutex
	m  map[string]RegisteredPrompt
}

func newMuxMapRegisteredPrompt() *muxMapRegisteredPrompt {
	return &muxMapRegisteredPrompt{
		m: make(map[string]RegisteredPrompt),
	}
}

func (xm *muxMapRegisteredPrompt) Clear() {
	xm.mu.Lock()
	xm.m = make(map[string]RegisteredPrompt)
	xm.mu.Unlock()
}

func (xm *muxMapRegisteredPrompt) Get(key string) (RegisteredPrompt, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapRegisteredPrompt) GetAll() map[string]RegisteredPrompt {
	xm.mu.RLock()
	clonedMap := make(map[string]RegisteredPrompt)
	for key, value := range xm.m {
		clonedMap[key] = value
	}
	xm.mu.RUnlock()
	return clonedMap
}

func (xm *muxMapRegisteredPrompt) Set(key string, value RegisteredPrompt) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}
func (xm *muxMapRegisteredPrompt) Delete(key string) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}
