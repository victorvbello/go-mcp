package server

import (
	"fmt"

	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
)

//High-level MCP server that provides a simpler API for working with resources, tools, and prompts.
//For advanced usage (like sending notifications or setting custom request handlers), use the underlying
//Server instance available via the `server` property.
type McpServer struct {
	//The underlying Server instance, useful for advanced operations like sending notifications.
	server                       *Server
	registeredResources          *muxMapRegisteredResource
	registeredResourceTemplates  *muxMapRegisteredResourceTemplate
	registeredTools              *muxMapRegisteredTool
	registeredPrompts            *muxMapRegisteredPrompt
	toolHandlersInitialized      bool
	completionHandlerInitialized bool
	resourceHandlersInitialized  bool
	promptHandlersInitialized    bool
}

func NewMcpServer(serverInfo types.Implementation, opts ServerOptions) (*McpServer, error) {
	var err error
	nMcpServer := &McpServer{
		registeredResources:         newMuxMapRegisteredResource(),
		registeredResourceTemplates: newMuxMapRegisteredResourceTemplate(),
		registeredTools:             newMuxMapRegisteredTool(),
		registeredPrompts:           newMuxMapRegisteredPrompt(),
	}
	nMcpServer.server, err = NewServer(serverInfo, opts)
	if err != nil {
		return nil, fmt.Errorf("newServer,%v", err)
	}
	return nMcpServer, nil
}

func (mcps *McpServer) wrapperOnErrorServer(fn func()) error {
	chanError := make(chan error)
	mcps.server.SetOnErrorCallBack(func(err error) {
		chanError <- err
	})
	fn()
	sError := <-chanError
	mcps.server.SetOnErrorCallBack(nil)
	return sError
}

func (mcps *McpServer) setToolRequestHandlers() error {
	if mcps.toolHandlersInitialized {
		return nil
	}

	ltr := types.NewListToolsRequest(nil)
	if err := mcps.server.AssertCanSetRequestHandler(ltr.Method); err != nil {
		return fmt.Errorf("mcps.server.AssertCanSetRequestHandler, %v", err)
	}

	if err := mcps.server.RegisterCapabilities(types.ServerCapabilities{
		Tools: &types.ServerCapabilitiesListChanged{ListChanged: true},
	}); err != nil {
		return fmt.Errorf("mcps.server.RegisterCapabilities, %v", err)
	}

	mcps.server.SetRequestHandler(types.NewListToolsRequest(nil),
		func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
			lt := new(types.ListToolsResult)
			for name, rT := range mcps.registeredTools.GetAll() {
				if !rT.Enabled {
					continue
				}
				tool := types.Tool{
					Description: rT.Description,
					InputSchema: rT.InputSchema,
					Annotations: rT.Annotations,
				}
				tool.Name = name
				tool.Title = rT.Title
				tool.OutputSchema = rT.OutputSchema
				lt.Tools = append(lt.Tools, tool)
			}
			return lt, nil
		})
	mcps.server.SetRequestHandler(types.NewCallToolRequest(nil),
		func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
			req, okType := request.(*types.CallToolRequest)
			if !okType {
				err := types.NewMcpError(types.ERROR_CODE_INVALID_PARAMS, "invalid request type CallToolRequest", nil)
				return nil, err.ToError()
			}
			tool, okTool := mcps.registeredTools.Get(req.Params.Name)
			if !okTool {
				err := types.NewMcpError(types.ERROR_CODE_INVALID_PARAMS,
					fmt.Sprintf("tool %s not found", req.Params.Name), nil)
				return nil, err.ToError()
			}
			if !tool.Enabled {
				err := types.NewMcpError(types.ERROR_CODE_INVALID_PARAMS,
					fmt.Sprintf("tool %s disabled", req.Params.Name), nil)
				return nil, err.ToError()
			}

			var result *types.CallToolResult
			var err error
			result, err = tool.Callback(req.Params.Arguments, extra)
			if err != nil {
				txtContent := types.NewTextContent(fmt.Sprintf("tool.Callback, %v", err))
				isErr := true
				result = &types.CallToolResult{
					Content: []types.Content{txtContent},
					IsError: &isErr,
				}
			}
			if tool.OutputSchema.Type != "" && result.IsError != nil && !*result.IsError {
				if result.StructuredContent == nil {
					err := types.NewMcpError(types.ERROR_CODE_INVALID_PARAMS,
						fmt.Sprintf("tool %s has an output schema but no structured content was provided", req.Params.Name), nil)
					return nil, err.ToError()
				}
			}
			return result, nil
		},
	)
	mcps.toolHandlersInitialized = true
	return nil
}

func (mcps *McpServer) setCompletionRequestHandler() error {
	if mcps.completionHandlerInitialized {
		return nil
	}

	cr := types.NewCompleteRequest(nil)
	if err := mcps.server.AssertCanSetRequestHandler(cr.Method); err != nil {
		return fmt.Errorf("mcps.server.AssertCanSetRequestHandler, %v", err)
	}

	if err := mcps.server.RegisterCapabilities(types.ServerCapabilities{
		Completions: map[string]interface{}{},
	}); err != nil {
		return fmt.Errorf("mcps.server.RegisterCapabilities, %v", err)
	}

	mcps.server.SetRequestHandler(types.NewCompleteRequest(nil),
		func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
			req, okType := request.(*types.CompleteRequest)
			if !okType {
				err := types.NewMcpError(types.ERROR_CODE_INVALID_PARAMS, "invalid request type CompleteRequest", nil)
				return nil, err.ToError()
			}
			switch rt := req.Params.Ref.(type) {
			case *types.PromptReference:
				return mcps.handlePromptCompletion(req, rt)
			case *types.ResourceTemplateReference:
				return mcps.handleResourceCompletion(req, rt)
			default:
				err := types.NewMcpError(
					types.ERROR_CODE_INVALID_PARAMS,
					fmt.Sprintf("invalid completion reference: %T", rt), nil)
				return nil, err.ToError()
			}
		})
	mcps.completionHandlerInitialized = true
	return nil
}

func (mcps *McpServer) handlePromptCompletion(request *types.CompleteRequest, ref *types.PromptReference) (*types.CompleteResult, error) {
	prompt, okPrompt := mcps.registeredPrompts.Get(ref.Name)
	if !okPrompt {
		err := types.NewMcpError(
			types.ERROR_CODE_INVALID_PARAMS,
			fmt.Sprintf("prompt %s not found", ref.Name), nil)
		return nil, err.ToError()
	}

	if !prompt.Enabled {
		err := types.NewMcpError(
			types.ERROR_CODE_INVALID_PARAMS,
			fmt.Sprintf("prompt %s disabled", ref.Name), nil)
		return nil, err.ToError()
	}
	if prompt.ArgsSchema != nil {
		return &types.CompleteResult{}, nil
	}
	argSchema, okArg := prompt.ArgsSchema[request.Params.Argument.Name]
	if !okArg {
		return &types.CompleteResult{}, nil
	}
	suggestions := argSchema.Complete(request.Params.Argument.Value, request.Params.Context)
	return mcps.createCompletionResult(suggestions), nil
}

func (mcps *McpServer) handleResourceCompletion(request *types.CompleteRequest, ref *types.ResourceTemplateReference) (*types.CompleteResult, error) {
	var template *RegisteredResourceTemplate
	reqRef := request.Params.Ref.(*types.ResourceTemplateReference)
	for _, t := range mcps.registeredResourceTemplates.GetAll() {
		uri := t.ResourceTemplate.GetUriTemplate()
		if uri.String() == ref.URI {
			template = &t
		}
	}
	if template == nil {
		if _, ok := mcps.registeredResources.Get(ref.URI); ok {
			return &types.CompleteResult{}, nil
		}
		err := types.NewMcpError(
			types.ERROR_CODE_INVALID_PARAMS,
			fmt.Sprintf("resource template %s not found", reqRef.URI), nil)
		return nil, err.ToError()
	}
	completer := template.ResourceTemplate.CompleteCallback(request.Params.Argument.Name)
	if completer == nil {
		return &types.CompleteResult{}, nil
	}
	suggestions, err := completer(request.Params.Argument.Value, request.Params.Context)
	if err != nil {
		err := types.NewMcpError(
			types.ERROR_CODE_INVALID_PARAMS,
			fmt.Sprintf("resource template %s completer error %v", reqRef.URI, err), nil)
		return nil, err.ToError()
	}
	return mcps.createCompletionResult(suggestions), nil
}

func (mcps *McpServer) setResourceRequestHandlers() error {
	if mcps.resourceHandlersInitialized {
		return nil
	}

	methodsToCheck := []string{
		types.NewListResourcesRequest(nil).Method,
		types.NewListResourceTemplatesRequest(nil).Method,
		types.NewReadResourceRequest(nil).Method,
	}

	for _, m := range methodsToCheck {
		if err := mcps.server.AssertCanSetRequestHandler(m); err != nil {
			return fmt.Errorf("mcps.server.AssertCanSetRequestHandler, %v", err)
		}
	}

	scr := new(types.ServerCapabilitiesResources)
	scr.ListChanged = true

	if err := mcps.server.RegisterCapabilities(types.ServerCapabilities{
		Resources: scr,
	}); err != nil {
		return fmt.Errorf("mcps.server.RegisterCapabilities, %v", err)
	}

	mcps.server.SetRequestHandler(types.NewListResourcesRequest(nil),
		func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
			var resources, templateResources []types.Resource
			for uri, rr := range mcps.registeredResources.GetAll() {
				if !rr.Enabled {
					continue
				}
				nR := types.Resource{
					URI: uri,
				}
				nR.Name = rr.Name
				if rr.Metadata != nil {
					nR.Metadata = rr.Metadata.Metadata
				}
				resources = append(resources, nR)
			}
			for uri, template := range mcps.registeredResourceTemplates.GetAll() {
				listCallback := template.ResourceTemplate.GetListCallback()
				if listCallback == nil {
					continue
				}
				result, err := listCallback(extra)
				if err != nil {
					err := types.NewMcpError(
						types.ERROR_CODE_INVALID_PARAMS,
						fmt.Sprintf("listCallback of %s, %v", uri, err), nil)
					return nil, err.ToError()
				}
				if result == nil {
					err := types.NewMcpError(
						types.ERROR_CODE_INVALID_PARAMS,
						fmt.Sprintf("listCallback of %s, empty result", uri), nil)
					return nil, err.ToError()
				}
				for _, resource := range result.Resources {
					newResource := resource
					if template.Metadata != nil {
						newResource.Metadata = template.Metadata.Metadata
					}
					templateResources = append(templateResources, newResource)
				}
			}
			result := new(types.ListResourcesResult)
			result.Resources = append(result.Resources, resources...)
			result.Resources = append(result.Resources, templateResources...)
			return result, nil
		})

	mcps.server.SetRequestHandler(types.NewListResourceTemplatesRequest(nil),
		func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
			var resourceTemplates []types.ResourceTemplate
			for name, template := range mcps.registeredResourceTemplates.GetAll() {
				uri := template.ResourceTemplate.GetUriTemplate()
				nrt := types.ResourceTemplate{
					URITemplate: uri,
					Metadata:    template.Metadata.Metadata,
				}
				nrt.Name = name
				if template.Metadata != nil {
					nrt.Metadata = template.Metadata.Metadata
				}
				resourceTemplates = append(resourceTemplates, nrt)
			}
			result := new(types.ListResourceTemplatesResult)
			result.ResourceTemplates = append(result.ResourceTemplates, resourceTemplates...)
			return result, nil
		})

	mcps.server.SetRequestHandler(types.NewReadResourceRequest(nil),
		func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
			req, okType := request.(*types.ReadResourceRequest)
			if !okType {
				err := types.NewMcpError(types.ERROR_CODE_INVALID_PARAMS, "invalid request type ReadResourceRequest", nil)
				return nil, err.ToError()
			}
			uri := req.Params.URI

			//First check for exact resource match
			resource, okResource := mcps.registeredResources.Get(uri)
			if okResource {
				if !resource.Enabled {
					err := types.NewMcpError(types.ERROR_CODE_INVALID_PARAMS,
						fmt.Sprintf("resource %s disabled", uri), nil)
					return nil, err.ToError()
				}
				return resource.ReadCallback(uri, extra)
			}

			//Then check templates
			for _, template := range mcps.registeredResourceTemplates.GetAll() {
				variables, err := template.ResourceTemplate.uriTemplate.Match(uri)
				if err != nil {
					err := types.NewMcpError(types.ERROR_CODE_INVALID_PARAMS,
						fmt.Sprintf("template.ResourceTemplate.uriTemplate.Match %s, %v", uri, err), nil)
					return nil, err.ToError()
				}
				if variables != nil {
					return template.ReadCallback(uri, variables, extra)
				}
			}

			err := types.NewMcpError(types.ERROR_CODE_INVALID_PARAMS,
				fmt.Sprintf("resource %s not found", uri), nil)
			return nil, err.ToError()
		})

	if err := mcps.setCompletionRequestHandler(); err != nil {
		return fmt.Errorf("mcps.setCompletionRequestHandler, %v", err)
	}

	mcps.resourceHandlersInitialized = true
	return nil
}

func (mcps *McpServer) setPromptRequestHandlers() error {
	if mcps.promptHandlersInitialized {
		return nil
	}

	methodsToCheck := []string{
		types.NewListPromptsRequest(nil).Method,
		types.NewGetPromptRequest(nil).Method,
	}

	for _, m := range methodsToCheck {
		if err := mcps.server.AssertCanSetRequestHandler(m); err != nil {
			return fmt.Errorf("mcps.server.AssertCanSetRequestHandler, %v", err)
		}
	}

	scr := new(types.ServerCapabilitiesListChanged)
	scr.ListChanged = true

	if err := mcps.server.RegisterCapabilities(types.ServerCapabilities{
		Prompts: scr,
	}); err != nil {
		return fmt.Errorf("mcps.server.RegisterCapabilities, %v", err)
	}

	mcps.server.SetRequestHandler(types.NewListPromptsRequest(nil),
		func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
			var result *types.ListPromptsResult
			var prompts []types.Prompt
			for name, prompt := range mcps.registeredPrompts.GetAll() {
				if !prompt.Enabled {
					continue
				}
				np := types.Prompt{
					Description: prompt.Description,
					Arguments:   mcps.promptArgumentsFromSchema(prompt.ArgsSchema),
				}
				np.Name = name
				np.Title = prompt.Title
				prompts = append(prompts, np)
			}
			result.Prompts = append(result.Prompts, prompts...)
			return result, nil
		})

	mcps.server.SetRequestHandler(types.NewGetPromptRequest(nil),
		func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
			req, okType := request.(*types.GetPromptRequest)
			if !okType {
				err := types.NewMcpError(types.ERROR_CODE_INVALID_PARAMS, "invalid request type GetPromptRequest", nil)
				return nil, err.ToError()
			}

			prompt, okPrompt := mcps.registeredPrompts.Get(req.Params.Name)
			if !okPrompt {
				err := types.NewMcpError(
					types.ERROR_CODE_INVALID_PARAMS,
					fmt.Sprintf("prompt %s not found", req.Params.Name), nil)
				return nil, err.ToError()
			}
			if !prompt.Enabled {
				err := types.NewMcpError(
					types.ERROR_CODE_INVALID_PARAMS,
					fmt.Sprintf("prompt %s disabled", req.Params.Name), nil)
				return nil, err.ToError()
			}
			var args map[string]string
			if prompt.ArgsSchema != nil {
				args = req.Params.Arguments
				return prompt.Callback(args, extra)
			} else {
				return prompt.Callback(nil, extra)
			}

		})

	if err := mcps.setCompletionRequestHandler(); err != nil {
		return fmt.Errorf("mcps.setCompletionRequestHandler, %v", err)
	}

	mcps.promptHandlersInitialized = true
	return nil
}

func (mcps *McpServer) createCompletionResult(suggestions []string) *types.CompleteResult {
	cr := new(types.CompleteResult)
	hasMore := len(suggestions) > 100
	cr.Completion = types.CompleteResultCompletion{
		Values:  suggestions[0:100],
		Total:   len(suggestions),
		HasMore: &hasMore,
	}
	return cr
}

func (mcps *McpServer) promptArgumentsFromSchema(args map[string]PromptArgsSchemaField) []types.PromptArgument {
	var result []types.PromptArgument
	for name, arg := range args {
		result = append(result, types.PromptArgument{
			Name:        name,
			Description: arg.Description,
			Required:    arg.IsOptional,
		})
	}
	return result
}

//Registers a resource `name` at a fixed URI, which will use the given callback to respond to read requests.
//name and uri are required
func (mcps *McpServer) RegisterResource(name string, uri string, meta *ResourceMetadata, rcb ReadResourceCallback) (*RegisteredResource, error) {
	if name == "" || uri == "" {
		return nil, fmt.Errorf("name and uri are required")
	}

	if _, ok := mcps.registeredResources.Get(uri); !ok {
		return nil, fmt.Errorf("resource %s is already registered", uri)
	}
	result := RegisteredResource{
		Name:         name,
		Metadata:     meta,
		ReadCallback: rcb,
		Enabled:      true,
	}
	result.Disable = func() {
		result.Update(RegisteredResourceUpdateOpts{Enabled: false})
	}
	result.Enable = func() {
		result.Update(RegisteredResourceUpdateOpts{Enabled: true})
	}
	result.Remove = func() {
		result.Update(RegisteredResourceUpdateOpts{URI: ""})
	}
	if meta != nil {
		result.Title = meta.Title
	}
	result.Update = func(updates RegisteredResourceUpdateOpts) error {
		if updates.URI == "" {
			mcps.registeredResources.Delete(uri)
		}
		if updates.URI != uri {
			mcps.registeredResources.Set(updates.URI, result)
		}
		if updates.Name != "" {
			result.Name = updates.Name
		}
		if updates.Title != "" {
			result.Title = updates.Title
		}
		if updates.Metadata != nil {
			result.Metadata = updates.Metadata
		}
		if updates.Callback != nil {
			result.ReadCallback = updates.Callback
		}
		if updates.Enabled != result.Enabled {
			result.Enabled = updates.Enabled
		}
		if err := mcps.SendResourceListChanged(); err != nil {
			return fmt.Errorf("mcps.SendResourceListChanged, %v", err)
		}
		return nil
	}
	mcps.registeredResources.Set(uri, result)
	if err := mcps.setResourceRequestHandlers(); err != nil {
		return nil, fmt.Errorf("mcps.setResourceRequestHandlers, %v", err)
	}
	if err := mcps.SendResourceListChanged(); err != nil {
		return nil, fmt.Errorf("mcps.setResourceRequestHandlers, %v", err)
	}
	return &result, nil
}

//Registers a resource `name` with a template pattern, which will use the given callback to respond to read requests.
//name is required
func (mcps *McpServer) RegisterResourceTemplate(name string, title string, temp ResourceTemplate, meta *ResourceMetadata, rcb ReadResourceTemplateCallback) (*RegisteredResourceTemplate, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if _, ok := mcps.registeredResourceTemplates.Get(name); !ok {
		return nil, fmt.Errorf("resource template %s is already registered", name)
	}
	result := RegisteredResourceTemplate{
		ResourceTemplate: temp,
		Title:            title,
		Metadata:         meta,
		ReadCallback:     rcb,
		Enabled:          true,
	}
	result.Disable = func() {
		result.Update(RegisteredResourceTemplateUpdateOpts{Enabled: false})
	}
	result.Enable = func() {
		result.Update(RegisteredResourceTemplateUpdateOpts{Enabled: true})
	}
	result.Remove = func() {
		result.Update(RegisteredResourceTemplateUpdateOpts{Name: ""})
	}
	result.Update = func(updates RegisteredResourceTemplateUpdateOpts) error {
		if updates.Name == "" {
			mcps.registeredResourceTemplates.Delete(name)
		}
		if updates.Name != name {
			mcps.registeredResourceTemplates.Set(updates.Name, result)
		}
		if updates.Title != "" {
			result.Title = updates.Title
		}
		if updates.Template != nil {
			result.ResourceTemplate = *updates.Template
		}
		if updates.Metadata != nil {
			result.Metadata = updates.Metadata
		}
		if updates.Callback != nil {
			result.ReadCallback = updates.Callback
		}
		if updates.Enabled != result.Enabled {
			result.Enabled = updates.Enabled
		}
		if err := mcps.SendResourceListChanged(); err != nil {
			return fmt.Errorf("mcps.SendResourceListChanged, %v", err)
		}
		return nil
	}
	mcps.registeredResourceTemplates.Set(name, result)
	if err := mcps.setResourceRequestHandlers(); err != nil {
		return nil, fmt.Errorf("mcps.setResourceRequestHandlers, %v", err)
	}
	if err := mcps.SendResourceListChanged(); err != nil {
		return nil, fmt.Errorf("mcps.SendResourceListChanged, %v", err)
	}
	return &result, nil
}

//Registers a tool with a config object and callback.
//name is required
type RegisterToolOpts struct {
	Name         string
	Title        string
	Description  string
	InputSchema  types.ToolInputSchema
	OutputSchema types.ToolOutputSchema
	Annotations  *types.ToolAnnotations
	Callback     ToolCallback
}

func (mcps *McpServer) RegisterTool(opts RegisterToolOpts) (*RegisteredTool, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if _, ok := mcps.registeredTools.Get(opts.Name); ok {
		return nil, fmt.Errorf("tool %s is already registered", opts.Name)
	}

	result := RegisteredTool{
		Title:        opts.Title,
		Description:  opts.Description,
		InputSchema:  opts.InputSchema,
		OutputSchema: opts.OutputSchema,
		Annotations:  opts.Annotations,
		Callback:     opts.Callback,
		Enabled:      true,
	}
	result.Disable = func() {
		result.Update(RegisteredToolUpdateOpts{Enabled: false})
	}
	result.Enable = func() {
		result.Update(RegisteredToolUpdateOpts{Enabled: true})
	}
	result.Remove = func() {
		result.Update(RegisteredToolUpdateOpts{Name: ""})
	}
	result.Update = func(updates RegisteredToolUpdateOpts) error {
		if updates.Name == "" {
			mcps.registeredTools.Delete(opts.Name)
		}
		if updates.Name != opts.Name {
			mcps.registeredTools.Set(updates.Name, result)
		}
		if updates.Title != "" {
			result.Title = updates.Title
		}
		if updates.Description != "" {
			result.Description = updates.Description
		}
		if updates.ParamsSchema.Type != "" {
			result.InputSchema = updates.ParamsSchema
		}
		if updates.OutputSchema.Type != "" {
			result.OutputSchema = updates.OutputSchema
		}
		if updates.Callback != nil {
			result.Callback = updates.Callback
		}
		if updates.Annotations != nil {
			result.Annotations = updates.Annotations
		}
		if updates.Enabled != result.Enabled {
			result.Enabled = updates.Enabled
		}
		if err := mcps.SendToolListChanged(); err != nil {
			return fmt.Errorf("mcps.SendToolListChanged, %v", err)
		}
		return nil
	}
	mcps.registeredTools.Set(opts.Name, result)
	if err := mcps.setToolRequestHandlers(); err != nil {
		return nil, fmt.Errorf("mcps.setToolRequestHandlers, %v", opts.Name)
	}
	if err := mcps.SendToolListChanged(); err != nil {
		return nil, fmt.Errorf("mcps.SendToolListChanged, %v", err)
	}
	return &result, nil
}

//Registers a prompt with a config object and callback.
//name is required
func (mcps *McpServer) RegisterPrompt(name string, title string, des string, argSch map[string]PromptArgsSchemaField, cb PromptCallback) (*RegisteredPrompt, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if _, ok := mcps.registeredPrompts.Get(name); !ok {
		return nil, fmt.Errorf("prompt %s is already registered", name)
	}
	result := RegisteredPrompt{
		Title:       title,
		Description: des,
		ArgsSchema:  argSch,
		Callback:    cb,
	}
	result.Disable = func() {
		result.Update(RegisteredPromptUpdateOpts{Enabled: false})
	}
	result.Enable = func() {
		result.Update(RegisteredPromptUpdateOpts{Enabled: true})
	}
	result.Remove = func() {
		result.Update(RegisteredPromptUpdateOpts{Name: ""})
	}
	result.Update = func(updates RegisteredPromptUpdateOpts) error {
		if updates.Name == "" {
			mcps.registeredPrompts.Delete(name)
		}
		if updates.Name != name {
			mcps.registeredPrompts.Set(updates.Name, result)
		}
		if updates.Title != "" {
			result.Title = updates.Title
		}
		if updates.Description != "" {
			result.Description = updates.Description
		}
		if updates.ArgsSchema != nil {
			result.ArgsSchema = updates.ArgsSchema
		}
		if updates.Callback != nil {
			result.Callback = updates.Callback
		}
		if updates.Enabled != result.Enabled {
			result.Enabled = updates.Enabled
		}
		mcps.SendPromptListChanged()
		if err := mcps.SendPromptListChanged(); err != nil {
			return fmt.Errorf("mcps.SendPromptListChanged, %v", err)
		}
		return nil
	}

	if err := mcps.setPromptRequestHandlers(); err != nil {
		return nil, fmt.Errorf("mcps.setPromptRequestHandlers, %v", err)
	}
	if err := mcps.SendPromptListChanged(); err != nil {
		return nil, fmt.Errorf("mcps.SendPromptListChanged, %v", err)
	}
	return &result, nil
}

//Checks if the server is connected to a transport.
//returns True if the server is connected
func (mcps *McpServer) IsConnected() bool {
	return mcps.server.GetTransport() != nil
}

//Sends a resource list changed event to the client, if connected.
func (mcps *McpServer) SendResourceListChanged() error {
	if !mcps.IsConnected() {
		return nil
	}
	if err := mcps.server.SendResourceListChanged(); err != nil {
		return fmt.Errorf("mcps.server.SendResourceListChanged, %v", err)
	}
	return nil
}

//Sends a tool list changed event to the client, if connected.
func (mcps *McpServer) SendToolListChanged() error {
	if !mcps.IsConnected() {
		return nil
	}
	if err := mcps.server.SendToolListChanged(); err != nil {
		return fmt.Errorf("mcps.server.SendToolListChanged, %v", err)
	}
	return nil
}

//Sends a prompt list changed event to the client, if connected.
func (mcps *McpServer) SendPromptListChanged() error {
	if !mcps.IsConnected() {
		return nil
	}
	if err := mcps.server.SendPromptListChanged(); err != nil {
		return fmt.Errorf("mcps.server.SendPromptListChanged, %v", err)
	}
	return nil
}

//Attaches to the given transport, starts it, and starts listening for messages.
//
//The `server` object assumes ownership of the Transport, replacing any callbacks that have already been set, and expects that it is the only user of the Transport instance going forward.
func (mcps *McpServer) Connect(transport shared.Transport) error {
	connError := mcps.wrapperOnErrorServer(func() {
		mcps.server.Protocol.Connect(transport)
	})
	return connError
}

//Closes the connection.
func (mcps *McpServer) Close() error {
	connError := mcps.wrapperOnErrorServer(func() {
		mcps.server.Close()
	})
	return connError
}

//Get server read only
func (mcps *McpServer) GetServer() *Server {
	return mcps.server
}

//Add external Action on onInitialized
//Callback for when initialization has fully completed (i.e., the client has sent an `initialized` notification).
func (mcps *McpServer) SetOnInitialized(fn func() error) {
	mcps.server.OnInitialized = fn
}
