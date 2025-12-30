package client

import (
	"context"
	"fmt"

	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	utils "github.com/victorvbello/gomcp/mcp/utils/logger"
)

type ClientOptions struct {
	shared.ProtocolOptions
	//Capabilities to advertise as being supported by this client.
	Capabilities types.ClientCapabilities
	//Configure handlers for list changed notifications (tools, prompts, resources).
	ListChanged types.ClientCapabilitiesListChangedHandlers
	//Optional instructions describing how to use the client and its features.
	Instructions string
}

//An MCP client on top of a pluggable transport.
//
//The client will automatically begin the initialization flow with the server when connect() is called.
type Client struct {
	*shared.Protocol
	serverCapabilities         types.ServerCapabilities
	serverVersion              types.Implementation
	capabilities               types.ClientCapabilities
	instructions               string
	clientInfo                 types.Implementation
	cachedToolOutputValidators *muxCachedToolOutputValidators
	onErrorCallBack            func(err error)
	logger                     utils.LogService
	wrapperOnErrorChan         chan error
}

func NewClient(clientInfo types.Implementation, opts ClientOptions) (*Client, error) {
	cli := &Client{
		clientInfo:                 clientInfo,
		capabilities:               opts.Capabilities,
		instructions:               opts.Instructions,
		cachedToolOutputValidators: newMuxCachedToolOutputValidators(),
		logger:                     utils.NewLoggerService("client"),
	}
	protocol := shared.NewProtocol(&opts.ProtocolOptions, cli)
	cli.Protocol = protocol
	return cli, nil
}

//ProtocolInterface Methods
func (c *Client) ProtocolInterfaceType() int {
	return shared.CLIENT_PROTOCOLO_INTERFACE_TYPE
}

//Callback for when the connection is closed for any reason.
//
//This is invoked when close() is called as well.
func (c *Client) OnClose() error {
	return nil
}

//Callback for when an error occurs.
//
//Note that errors are not necessarily fatal; they are used for reporting any kind of exceptional condition out of band.
func (c *Client) OnError(err error) error {
	if err != nil {
		c.logger.Error(nil, err.Error())
	}
	fmt.Println("client-on-error")
	c.onErrorCallBack(err)
	go func() {
		if c.wrapperOnErrorChan == nil {
			return
		}
		c.wrapperOnErrorChan <- err
	}()
	return nil
}

//Add external Action on error
func (c *Client) SetOnErrorCallBack(fn func(err error)) {
	c.onErrorCallBack = fn
}

//A handler to invoke for any request types that do not have their own handler installed.
func (c *Client) FallbackRequestHandler() shared.RequestHandler {
	return nil
}

//A handler to invoke for any notification types that do not have their own handler installed.
func (c *Client) FallbackNotificationHandler() shared.NotificationHandler {
	return nil
}

//A method to check if a capability is supported by the remote side, for the given method to be called.
//
//This should be implemented by parent struct
func (c *Client) AssertCapabilityForMethod(sReq types.RequestInterface) error {
	switch r := sReq.(type) {
	case *types.SetLevelRequest:
		if c.serverCapabilities.Logging == nil {
			return fmt.Errorf("server does not support logging (required for %s)", r.Method)
		}
		return nil
	case *types.ListPromptsRequest:
	case *types.GetPromptRequest:
		if c.serverCapabilities.Prompts == nil {
			return fmt.Errorf("server does not support prompts (required for %s)", r.Method)
		}
		return nil
	case *types.ListResourcesRequest:
	case *types.ListResourceTemplatesRequest:
	case *types.ReadResourceRequest:
	case *types.SubscribeRequest:
	case *types.UnsubscribeRequest:
		if c.serverCapabilities.Resources == nil {
			return fmt.Errorf("server does not support resources (required for %s)", r.Method)
		}
		if !c.serverCapabilities.Resources.Subscribe {
			return fmt.Errorf("server does not support resources subscriptions (required for %s)", r.Method)
		}
		return nil
	case *types.ListToolsRequest:
	case *types.CallToolRequest:
		if c.serverCapabilities.Tools == nil {
			return fmt.Errorf("server does not support tools (required for %s)", r.Method)
		}
		return nil
	case *types.CompleteRequest:
		if c.serverCapabilities.Completions == nil {
			return fmt.Errorf("server does not support completions (required for %s)", r.Method)
		}
		return nil
	case *types.InitializeRequest:
		//No specific capability required for initialize
		return nil
	case *types.PingRequest:
		//No specific capability required for ping
		return nil
	}
	return nil
}

//A method to check if a notification is supported by the local side, for the given method to be sent.
//
//This should be implemented by parent struct
func (c *Client) AssertNotificationCapability(sNotify types.NotificationInterface) error {
	switch n := sNotify.(type) {
	case *types.RootsListChangedNotification:
		if !c.capabilities.Roots.ListChanged {
			return fmt.Errorf("client does not support roots list changed notifications (required for %s)", n.Method)
		}
		return nil
	case *types.InitializedNotification:
		//No specific capability required for initialized
		return nil
	case *types.CancelledNotification:
		//Cancellation notifications are always allowed
		return nil
	case *types.ProgressNotification:
		//Progress notifications are always allowed
		return nil
	}
	return nil
}

//A method to check if a request handler is supported by the local side, for the given method to be handled.
//
//This should be implemented by parent struct
func (s *Client) AssertRequestHandlerCapability(req types.RequestInterface) error {
	switch r := req.(type) {
	case *types.CreateMessageRequest:
		if s.capabilities.Sampling == nil {
			return fmt.Errorf("client does not support sampling capability (required for %s)", r.Method)
		}
		return nil
	case *types.ListRootsRequest:
		if s.capabilities.Roots == nil {
			return fmt.Errorf("client does not support roots capability (required for %s)", r.Method)
		}
		return nil
	case *types.InitializeRequest:
	case *types.PingRequest:
		//No specific capability required for these methods
		return nil
	}
	return nil
}

//Client Methods

//Registers new capabilities. This can only be called before connecting to a transport.
//
//The new capabilities will be merged with any existing capabilities previously given (e.g., at initialization).
func (c *Client) RegisterCapabilities(capabilities types.ClientCapabilities) error {
	if c.GetTransport() != nil {
		return fmt.Errorf("cannot register capabilities after connecting to transport")
	}
	c.capabilities.UpdateAll(&capabilities)
	return nil
}

func (c *Client) Connect(ctx context.Context, transport shared.Transport, options *shared.RequestOptions) error {
	c.Protocol.Connect(ctx, transport)
	//When transport sessionId is already set this means we are trying to reconnect.
	//In this case we don't need to initialize again.

	var err error
	var result types.ResultInterface
	for {
		result, err = c.Protocol.Request(types.NewInitializeRequest(&types.InitializeRequestParams{
			ProtocolVersion: types.LATEST_PROTOCOL_VERSION,
			Capabilities:    c.capabilities,
			ClientInfo:      c.clientInfo,
		}), options)
		if err != nil {
			return fmt.Errorf("request of InitializeRequest error:%v", err)
		}
		if result != nil {
			break
		}
	}
	realResult := result.(*types.InitializeResult)
	if _, ok := types.SUPPORTED_PROTOCOL_VERSIONS[realResult.ProtocolVersion]; !ok {
		return fmt.Errorf("server's protocol version is not supported: %s", realResult.ProtocolVersion)
	}
	c.serverCapabilities = realResult.Capabilities
	c.serverVersion = realResult.ServerInfo
	c.instructions = realResult.Instructions

	err = c.Notification(types.NewInitializedNotification(nil), nil)
	if err != nil {
		c.Close()
		return fmt.Errorf("initializedNotification error:%v", err)
	}
	return nil
}

func (c *Client) Close() error {
	closeError := c.wrapperOnError(func() {
		c.Protocol.Close()
	})
	return closeError
}

//After initialization has completed, this will be populated with the server's reported capabilities.
func (c *Client) GetServerCapabilities() types.ServerCapabilities {
	return c.serverCapabilities
}

//After initialization has completed, this will be populated with information about the server's name and version.
func (c *Client) GetServerVersion() types.Implementation {
	return c.serverVersion
}

//After initialization has completed, this may be populated with information about the server's instructions.
func (c *Client) GetInstructions() string {
	return c.instructions
}

func (c *Client) Ping(opts *shared.RequestOptions) (*types.EmptyResult, error) {
	result, err := c.Protocol.Request(types.NewPingRequest(), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request PingRequest %v", err)
	}
	realResult := result.(*types.EmptyResult)
	return realResult, nil
}

func (c *Client) Complete(params types.CompleteParams, opts *shared.RequestOptions) (*types.CompleteResult, error) {
	result, err := c.Protocol.Request(types.NewCompleteRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request CompleteRequest %v", err)
	}
	realResult := result.(*types.CompleteResult)
	return realResult, nil
}

func (c *Client) SetLoggingLevel(level types.LoggingLevel, opts *shared.RequestOptions) (*types.EmptyResult, error) {
	result, err := c.Protocol.Request(types.NewSetLevelRequest(&types.SetLevelRequestParams{Level: level}), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request SetLevelRequest %v", err)
	}
	realResult := result.(*types.EmptyResult)
	return realResult, nil
}

func (c *Client) GetPrompt(params types.GetPromptParams, opts *shared.RequestOptions) (*types.GetPromptResult, error) {
	result, err := c.Protocol.Request(types.NewGetPromptRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request GetPromptRequest %v", err)
	}
	realResult := result.(*types.GetPromptResult)
	return realResult, nil
}

func (c *Client) ListPrompts(params types.PaginatedRequestParams, opts *shared.RequestOptions) (*types.ListPromptsResult, error) {
	result, err := c.Protocol.Request(types.NewListPromptsRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request ListPromptsRequest %v", err)
	}
	realResult := result.(*types.ListPromptsResult)
	return realResult, nil
}

func (c *Client) ListResources(params types.PaginatedRequestParams, opts *shared.RequestOptions) (*types.ListResourcesResult, error) {
	result, err := c.Protocol.Request(types.NewListResourcesRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request ListResourcesRequest %v", err)
	}
	realResult := result.(*types.ListResourcesResult)
	return realResult, nil
}

func (c *Client) ListResourceTemplates(params types.PaginatedRequestParams, opts *shared.RequestOptions) (*types.ListResourceTemplatesResult, error) {
	result, err := c.Protocol.Request(types.NewListResourceTemplatesRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request ListResourceTemplatesRequest %v", err)
	}
	realResult := result.(*types.ListResourceTemplatesResult)
	return realResult, nil
}

func (c *Client) ReadResource(params types.ReadResourceRequestParams, opts *shared.RequestOptions) (*types.ReadResourceResult, error) {
	result, err := c.Protocol.Request(types.NewReadResourceRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request ReadResourceRequest %v", err)
	}
	realResult := result.(*types.ReadResourceResult)
	return realResult, nil
}

func (c *Client) SubscribeResource(params types.SubscribeRequestParams, opts *shared.RequestOptions) (*types.EmptyResult, error) {
	result, err := c.Protocol.Request(types.NewSubscribeRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request SubscribeRequest %v", err)
	}
	realResult := result.(*types.EmptyResult)
	return realResult, nil
}

func (c *Client) UnsubscribeResource(params types.UnsubscribeRequestParams, opts *shared.RequestOptions) (*types.EmptyResult, error) {
	result, err := c.Protocol.Request(types.NewUnsubscribeRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request UnsubscribeRequest %v", err)
	}
	realResult := result.(*types.EmptyResult)
	return realResult, nil
}

func (c *Client) CallTool(params types.CallToolRequestParams, opts *shared.RequestOptions) (*types.CallToolResult, error) {
	result, err := c.Protocol.Request(types.NewCallToolRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request CallToolRequest %v", err)
	}
	if result == nil {
		return nil, fmt.Errorf("server sent invalid initialize result: %v", result)
	}
	realResult := result.(*types.CallToolResult)
	//Check if the tool has an outputSchema
	validator := c.getToolOutputValidator(params.Name)
	if validator == nil {
		//If don't has an outputSchema registered return
		c.logger.Info(nil, fmt.Sprintf("tool %s not has an outputSchema registered", params.Name))
		return realResult, nil
	}
	if realResult.StructuredContent == nil {
		//If don't has an outputSchema return
		c.logger.Info(nil, fmt.Sprintf("tool %s not has an outputSchema defined", params.Name))
		return realResult, nil
	}
	//If tool has outputSchema, it MUST return structuredContent (unless it's an error)
	if realResult.StructuredContent == nil && realResult.IsError != nil {
		return nil, fmt.Errorf("tool %s has an output schema but did not return structured content", params.Name)
	}
	//Only validate structured content if present (not when there's an error)
	isValid, err := validator(realResult.StructuredContent)
	if !isValid {
		return nil, types.NewMcpError(
			types.ERROR_CODE_INVALID_PARAMS,
			"structured content does not match the tool's output schema", err).ToError()
	}
	if err != nil {
		return nil, types.NewMcpError(
			types.ERROR_CODE_INVALID_PARAMS,
			"failed to validate structured content", err).ToError()
	}
	return realResult, nil
}

func (c *Client) ListTools(params types.PaginatedRequestParams, opts *shared.RequestOptions) (*types.ListToolsResult, error) {
	result, err := c.Protocol.Request(types.NewListToolsRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request ListToolsRequest %v", err)
	}
	realResult := result.(*types.ListToolsResult)
	// Cache the tools and their output schemas for future validation
	c.cacheToolOutputSchemas(realResult.Tools)
	return realResult, nil
}

func (c *Client) SendRootsListChanged(params *types.BaseRequestParams, opts *shared.RequestOptions) (*types.ListRootsResult, error) {
	result, err := c.Protocol.Request(types.NewListRootsRequest(params), opts)
	if err != nil {
		return nil, fmt.Errorf("c.Protocol.Request ListRootsRequest %v", err)
	}
	realResult := result.(*types.ListRootsResult)
	return realResult, nil
}

func (c *Client) getToolOutputValidator(toolName string) clientToolOutputValidator {
	validator, ok := c.cachedToolOutputValidators.Get(toolName)
	if !ok {
		return nil
	}
	return validator
}

func (c *Client) cacheToolOutputSchemas(tools []types.Tool) {
	c.cachedToolOutputValidators.Clear()
	for _, t := range tools {
		if t.OutputSchema == nil {
			continue
		}
		validator, err := compileOutputToolSchema(t.OutputSchema)
		if err != nil {
			c.logger.Warning(nil, fmt.Sprintf("failed to compile output schema for tool %s: %v", t.Name, err))
			return
		}
		c.cachedToolOutputValidators.Set(t.Name, validator)

	}
}

func (c *Client) wrapperOnError(fn func()) error {
	c.wrapperOnErrorChan = make(chan error)
	fn()
	fmt.Println("wrapperOnErrorServer -1 ")
	sError := <-c.wrapperOnErrorChan
	fmt.Println("wrapperOnErrorServer -2 ")
	c.wrapperOnErrorChan = nil
	return sError
}
