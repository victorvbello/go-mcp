package server

import (
	"fmt"

	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	utils "github.com/victorvbello/gomcp/mcp/utils/logger"
)

type ServerOptions struct {
	shared.ProtocolOptions
	//Capabilities to advertise as being supported by this server.
	Capabilities types.ServerCapabilities
	//Optional instructions describing how to use the server and its features.
	Instructions string
}

//An MCP server on top of a pluggable transport.
//
//This server will automatically respond to the initialization flow as initiated from the client.
//
//To use with custom types, extend the base Request/Notification/Result types and pass them as type parameters:
type Server struct {
	*shared.Protocol
	clientCapabilities *types.ClientCapabilities
	clientVersion      *types.Implementation
	capabilities       types.ServerCapabilities
	instructions       string
	serverInfo         types.Implementation
	onErrorCallBack    func(err error)
	logger             utils.LogService
	//Callback for when initialization has fully completed (i.e., the client has sent an `initialized` notification).
	OnInitialized func() error
}

//Initializes this server with the given name and version information.
func NewServer(serverInfo types.Implementation, opts ServerOptions) (*Server, error) {
	srv := &Server{
		serverInfo:   serverInfo,
		capabilities: opts.Capabilities,
		instructions: opts.Instructions,
		logger:       utils.NewLoggerService(),
	}
	protocol := shared.NewProtocol(&opts.ProtocolOptions, srv)
	srv.Protocol = protocol
	srv.SetRequestHandler(types.NewInitializeRequest(nil), func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
		r, err := srv.onInitialize(request.(*types.InitializeRequest))
		if err != nil {
			return nil, fmt.Errorf("onInitialized %v", err)
		}
		return r, nil
	})

	srv.SetNotificationHandler(types.NewInitializedNotification(nil), func(notification types.NotificationInterface) error {
		if srv.OnInitialized != nil {
			err := srv.OnInitialized()
			if err != nil {
				return fmt.Errorf("OnInitialized %v", err)
			}
		}
		return nil
	})

	return srv, nil
}

//ProtocolInterface Methods
func (s *Server) ProtocolInterfaceType() int {
	return shared.SERVER_PROTOCOLO_INTERFACE_TYPE
}

//Callback for when the connection is closed for any reason.
//
//This is invoked when close() is called as well.
func (s *Server) OnClose() error {
	return nil
}

//Callback for when an error occurs.
//
//Note that errors are not necessarily fatal; they are used for reporting any kind of exceptional condition out of band.
func (s *Server) OnError(err error) error {
	s.logger.Error(nil, err.Error())
	return nil
}

//Add external Action on error
func (s *Server) SetOnErrorCallBack(fn func(err error)) {
	s.onErrorCallBack = fn
}

//A handler to invoke for any request types that do not have their own handler installed.
func (s *Server) FallbackRequestHandler() shared.RequestHandler {
	return func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
		return nil, nil
	}
}

//A handler to invoke for any notification types that do not have their own handler installed.
func (s *Server) FallbackNotificationHandler() shared.NotificationHandler {
	return func(notification types.NotificationInterface) error {
		return nil
	}
}

//A method to check if a capability is supported by the remote side, for the given method to be called.
//
//This should be implemented by parent struct
func (s *Server) AssertCapabilityForMethod(sReq types.RequestInterface) error {
	switch r := sReq.(type) {
	case *types.CreateMessageRequest:
		if s.clientCapabilities == nil || s.clientCapabilities.Sampling == nil {
			return fmt.Errorf("client does not support sampling (required for %s)", r.Method)
		}
		return nil
	case *types.ListRootsRequest:
		if s.clientCapabilities == nil || s.clientCapabilities.Roots == nil {
			return fmt.Errorf("client does not support listing roots (required for %s)", r.Method)
		}
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
func (s *Server) AssertNotificationCapability(sNotify types.NotificationInterface) error {
	switch n := sNotify.(type) {
	case *types.LoggingMessageNotification:
		if s.capabilities.Logging == nil {
			return fmt.Errorf("server does not support logging (required for %s)", n.Method)
		}
	case *types.ResourceUpdatedNotification:
	case *types.ResourceListChangedNotification:
		if s.capabilities.Resources == nil {
			return fmt.Errorf("server does not support notifying about resources (required for %s)", n.Method)
		}
		return nil
	case *types.ToolListChangedNotification:
		if s.capabilities.Tools == nil {
			return fmt.Errorf("server does not support notifying of tool list changes (required for %s)", n.Method)
		}
		return nil
	case *types.PromptListChangedNotification:
		if s.capabilities.Prompts == nil {
			return fmt.Errorf("server does not support notifying of prompt list changes (required for %s)", n.Method)
		}
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
func (s *Server) AssertRequestHandlerCapability(req types.RequestInterface) error {
	switch r := req.(type) {
	case *types.CreateMessageRequest:
		if s.capabilities.Sampling == nil {
			return fmt.Errorf("server does not support sampling (required for %s)", r.Method)
		}
		return nil
	case *types.SetLevelRequest:
		if s.capabilities.Logging == nil {
			return fmt.Errorf("server does not support logging (required for %s)", r.Method)
		}
		return nil
	case *types.GetPromptRequest:
	case *types.ListPromptsRequest:
		if s.capabilities.Prompts == nil {
			return fmt.Errorf("server does not support prompts (required for %s)", r.Method)
		}
		return nil
	case *types.ListResourcesRequest:
	case *types.ListResourceTemplatesRequest:
	case *types.ReadResourceRequest:
		if s.capabilities.Resources == nil {
			return fmt.Errorf("server does not support resources (required for %s)", r.Method)
		}
		return nil
	case *types.CallToolRequest:
	case *types.ListToolsRequest:
		if s.capabilities.Tools == nil {
			return fmt.Errorf("server does not support tools (required for %s)", r.Method)
		}
		return nil
	case *types.InitializeRequest:
	case *types.PingRequest:
		//No specific capability required for these methods
		return nil
	}
	return nil
}

//Server Methods

func (s *Server) onInitialize(request *types.InitializeRequest) (*types.InitializeResult, error) {
	requestedVersion := request.Params.ProtocolVersion
	s.clientCapabilities = &request.Params.Capabilities
	s.clientVersion = &request.Params.ClientInfo

	protocolVersion := requestedVersion
	_, okVersion := types.SUPPORTED_PROTOCOL_VERSIONS[requestedVersion]
	if !okVersion {
		protocolVersion = types.LATEST_PROTOCOL_VERSION
	}

	result := &types.InitializeResult{
		ProtocolVersion: protocolVersion,
		Capabilities:    s.getCapabilities(),
		ServerInfo:      s.serverInfo,
		Instructions:    s.instructions,
	}

	return result, nil
}

func (s *Server) getCapabilities() types.ServerCapabilities {
	return s.capabilities
}

//Registers new capabilities. This can only be called before connecting to a transport.
//
//The new capabilities will be merged with any existing capabilities previously given (e.g., at initialization).
func (s *Server) RegisterCapabilities(capabilities types.ServerCapabilities) error {
	if s.GetTransport() != nil {
		return fmt.Errorf("cannot register capabilities after connecting to transport")
	}
	s.capabilities = capabilities
	return nil
}

//After initialization has completed, this will be populated with the client's reported capabilities.
func (s *Server) GetClientCapabilities() *types.ClientCapabilities {
	return s.clientCapabilities
}

//After initialization has completed, this will be populated with information about the client's name and version.
func (s *Server) GetClientVersion() *types.Implementation {
	return s.clientVersion
}

func (s *Server) Ping() error {
	_, err := s.Protocol.Request(types.NewPingRequest(), nil)
	if err != nil {
		return fmt.Errorf("s.Protocol.Request %v", err)
	}
	return nil
}

func (s *Server) CreateMessage(params types.CreateMessageParams, opts *shared.RequestOptions) (types.ResultInterface, error) {
	result, err := s.Request(types.NewCreateMessageRequest(&params), opts)
	if err != nil {
		return nil, fmt.Errorf("s.Request, %v", err)
	}
	return result, nil
}

func (s *Server) ListRoots(params *types.BaseRequestParams, opts *shared.RequestOptions) (types.ResultInterface, error) {
	result, err := s.Request(types.NewListRootsRequest(params), opts)
	if err != nil {
		return nil, fmt.Errorf("s.Request, %v", err)
	}
	return result, nil
}

func (s *Server) SendLoggingMessage(params types.LoggingMessageNotificationParams) error {
	err := s.Notification(types.NewLoggingMessageNotification(&params), nil)
	if err != nil {
		return fmt.Errorf("s.Notification, %v", err)
	}
	return nil
}

func (s *Server) SendResourceUpdated(params types.ResourceUpdatedNotificationParams) error {
	err := s.Notification(types.NewResourceUpdatedNotification(&params), nil)
	if err != nil {
		return fmt.Errorf("s.Notification, %v", err)
	}
	return nil
}

func (s *Server) SendResourceListChanged() error {
	err := s.Notification(types.NewResourceListChangedNotification(nil), nil)
	if err != nil {
		return fmt.Errorf("s.Notification, %v", err)
	}
	return nil
}

func (s *Server) SendToolListChanged() error {
	err := s.Notification(types.NewToolListChangedNotification(nil), nil)
	if err != nil {
		return fmt.Errorf("s.Notification, %v", err)
	}
	return nil
}

func (s *Server) SendPromptListChanged() error {
	err := s.Notification(types.NewPromptListChangedNotification(nil), nil)
	if err != nil {
		return fmt.Errorf("s.Notification, %v", err)
	}
	return nil
}
