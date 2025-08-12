package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	"github.com/victorvbello/gomcp/mcp/utils"
)

const (
	MAXIMUM_MESSAGE_SIZE = 4 * 1024 * 1024
)

//Server transport for Streamable HTTP: this implements the MCP Streamable HTTP transport specification.
//It supports both SSE streaming and direct HTTP responses.
//
//Usage example:
//
//Stateful mode - server sets the session ID
//statefulTransport := NewStreamableHTTPServerTransport(StreamableHTTPServerTransportOptions{
//   SessionIDGenerator: func(){return randomUUID()},
//});
//
//Stateless mode - explicitly set session ID to undefined
//statelessTransport := NewStreamableHTTPServerTransport(StreamableHTTPServerTransportOptions{
//   SessionIDGenerator: nil,
//});
//
//
//In stateful mode:
//- Session ID is generated and included in response headers
//- Session ID is always included in initialization responses
//- Requests with invalid session IDs are rejected with 404 Not Found
//- Non-initialization requests without a session ID are rejected with 400 Bad Request
//- State is maintained in-memory (connections, message history)
//
//In stateless mode:
//- No Session ID is included in any responses
//- No session validation is performed
type StreamableHTTPServerTransport struct {
	protocolVersion              string
	started                      bool
	globalOnClose                func()
	globalOnError                func(err error)
	globalOnMessage              func(message types.JSONRPCMessage, extra *shared.MessageExtraInfo)
	sessionIDGenerator           func() string
	streamMapping                *muxMapStreamMapping
	requestToStreamMapping       *muxMapRequestToStreamMapping
	requestResponseMap           *muxMapRequestResponseMap
	initialized                  bool
	enableJSONResponse           bool
	standaloneSseStreamID        shared.StreamID
	eventStore                   shared.EventStore
	onSessionInitialized         func(sessionID string)
	onSessionClosed              func(sessionID string)
	allowedHosts                 map[string]struct{}
	allowedOrigins               map[string]struct{}
	enableDNSRebindingProtection bool
	//The session ID generated for this connection.
	SessionID string
}

func NewStreamableHTTPServerTransport(opts StreamableHTTPServerTransportOptions) shared.Transport {
	nst := &StreamableHTTPServerTransport{
		sessionIDGenerator:   opts.SessionIDGenerator,
		eventStore:           opts.EventStore,
		onSessionInitialized: opts.OnSessionInitialized,
		onSessionClosed:      opts.OnSessionClosed,
		allowedHosts:         opts.AllowedHosts,
		allowedOrigins:       opts.AllowedOrigins,
	}
	if opts.EnableJSONResponse != nil {
		nst.enableJSONResponse = *opts.EnableJSONResponse
	}
	if opts.EnableDNSRebindingProtection != nil {
		nst.enableDNSRebindingProtection = *opts.EnableDNSRebindingProtection
	}
	return nst
}

//Starts processing messages on the transport, including any connection steps that might need to be taken.
//
//This method should only be called after callbacks are installed, or else messages may be lost.
//
//NOTE: This method should not be called explicitly when using Client, Server, or Protocol classes, as they will implicitly call start().
func (s *StreamableHTTPServerTransport) Start() error {
	if s.started {
		return fmt.Errorf("transport already started")
	}
	s.started = true
	return nil
}

//Validates request headers for DNS rebinding protection.
//returns Error message if validation fails, undefined if validation passes.
func (s *StreamableHTTPServerTransport) validateRequestHeaders(req *http.Request) error {
	//Skip validation if protection is not enabled
	if !s.enableDNSRebindingProtection {
		return nil
	}
	//Validate Host header if allowedHosts is configured
	if len(s.allowedHosts) > 0 {
		hostHeader := req.Host
		if _, ok := s.allowedHosts[hostHeader]; !ok {
			return fmt.Errorf("invalid Host header: %s", hostHeader)
		}
	}

	//Validate Origin header if allowedOrigins is configured
	if len(s.allowedOrigins) > 0 {
		originHeader := req.Header.Get("Origin")
		if _, ok := s.allowedOrigins[originHeader]; !ok {
			return fmt.Errorf("invalid Origin header: %s", originHeader)
		}
	}
	return nil
}

//Validates session ID for non-initialization requests
//Returns true if the session is valid, false otherwise
func (s *StreamableHTTPServerTransport) validateSession(res ResponseWriter, req *http.Request) bool {
	if s.sessionIDGenerator == nil {
		//If the sessionIDGenerator ID is not set, the session management is disabled
		//and we don't need to validate the session ID
		return true
	}
	if !s.initialized {
		//If the server has not been initialized yet, reject all requests
		err := res.WriteJSON(http.StatusBadRequest, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_CONNECTION_CLOSED,
				Message: "Bad Request: Server not initialized",
			},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON bad Request: Server not initialized %v", err))
			return false
		}
	}
	sessionID := req.Header.Get(shared.TRANSPORT_HEADER_SESSION_ID)
	if sessionID == "" {
		//Non-initialization requests without a session ID should return 400 Bad Request
		err := res.WriteJSON(http.StatusBadRequest, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_CONNECTION_CLOSED,
				Message: "Bad Request: Mcp-Session-Id header is required",
			},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Bad Request: Mcp-Session-Id header is required %v", err))
			return false
		}
		return false
	}

	if !utils.IsAlphanumeric(sessionID) {
		err := res.WriteJSON(http.StatusBadRequest, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_CONNECTION_CLOSED,
				Message: "Bad Request: mcp-session-id header must be a alphanumeric value",
			},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Bad Request: mcp-session-id header must be a alphanumeric value %v", err))
			return false
		}

	}

	if sessionID != s.SessionID {
		//Reject requests with invalid session ID with 404 Not Found
		err := res.WriteJSON(http.StatusFound, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_SESSION_ID_NOT_FOUND,
				Message: "Session not found",
			},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Session not found %v", err))
			return false
		}
	}

	return true
}

func (s *StreamableHTTPServerTransport) validateProtocolVersion(res ResponseWriter, req *http.Request) bool {
	protocolVersion := req.Header.Get("mcp-protocol-version")
	if protocolVersion == "" {
		protocolVersion = types.DEFAULT_NEGOTIATED_PROTOCOL_VERSION
	}
	if _, ok := types.SUPPORTED_PROTOCOL_VERSIONS[protocolVersion]; !ok {
		err := res.WriteJSON(http.StatusBadRequest, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_CONNECTION_CLOSED,
				Message: fmt.Sprintf("Bad Request: Unsupported protocol version (supported versions:%v)", types.SUPPORTED_PROTOCOL_VERSIONS),
			},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Session not found %v", err))
			return false
		}
		return false
	}
	return true
}

//Handles an incoming HTTP request, whether GET or POST
func (s *StreamableHTTPServerTransport) HandleRequest(res ResponseWriter, req *http.Request) {
	// Validate request headers for DNS rebinding protection
	validationError := s.validateRequestHeaders(req)
	if validationError != nil {
		err := res.WriteJSON(http.StatusForbidden, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error:   &types.Error{Code: types.ERROR_CODE_CONNECTION_CLOSED, Message: validationError.Error()},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON %v", err))
			return
		}
		s.OnError(fmt.Errorf("validateRequestHeaders %v", validationError))
		return
	}

	switch req.Method {
	case http.MethodPost:
		s.handlePostRequest(res, req)
	case http.MethodGet:
		s.handleGetRequest(res, req)
	case http.MethodDelete:
		s.handleDeleteRequest(res, req)
	default:
		s.handleUnsupportedRequest(res)
	}
}

//Replays events that would have been sent after the specified event ID
//Only used when resumability is enabled
func (s *StreamableHTTPServerTransport) replayEvents(lastEventID shared.EventID, res ResponseWriter) {
	if s.eventStore == nil {
		return
	}

	headers := map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache, no-transform",
		"Connection":    "keep-alive",
	}

	if s.SessionID != "" {
		headers["mcp-session-id"] = s.SessionID
	}

	for k, v := range headers {
		res.Writer().Header().Set(k, v)
	}
	//Sends headers to the client
	res.Writer().WriteHeader(http.StatusOK)
	if f, ok := res.Writer().(http.Flusher); ok {
		//Flushes the buffered data (including headers)
		f.Flush()
	}

	waitResponse := make(chan bool)

	streamID, err := s.eventStore.ReplayEventsAfter(lastEventID, func(eventID shared.EventID, msg types.JSONRPCMessage) {
		if !s.writeSSEEvent(res, msg, eventID) {
			s.OnError(fmt.Errorf("failed replay events"))
			waitResponse <- true
			return
		}
	})

	if err != nil {
		s.OnError(fmt.Errorf("s.eventStore.ReplayEventsAfter %v", err))
		return
	}
	<-waitResponse
	s.streamMapping.Set(streamID, res)
}

//Writes an event to the SSE stream with proper formatting
func (s *StreamableHTTPServerTransport) writeSSEEvent(res ResponseWriter, msg types.JSONRPCMessage, eventID shared.EventID) bool {
	eventData := "event: message\n"
	//Include event ID if provided - this is important for resumability
	if eventID != "" {
		eventData += fmt.Sprintf("id: %s\n", eventID)
	}
	eventData += fmt.Sprintf("data: %v\n\n", msg)

	res.Writer().Write([]byte(eventData))
	if code, err := res.Writer().Write([]byte(eventData)); err != nil {
		s.OnError(fmt.Errorf("res.Writer().Write code:%d  %v", code, err))
		return false
	}

	return true
}

//Handles GET requests for SSE stream
func (s *StreamableHTTPServerTransport) handleGetRequest(res ResponseWriter, req *http.Request) {
	//The client MUST include an Accept header, listing text/event-stream as a supported content type.
	acceptHeader := req.Header.Get("Accept")
	if !strings.Contains(acceptHeader, "text/event-stream") {
		err := res.WriteJSON(http.StatusNotAcceptable, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_CONNECTION_CLOSED,
				Message: "Not Acceptable: Client must accept text/event-stream"},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Not Acceptable: Client must accept text/event-stream %v", err))
			return
		}
		return
	}
	//If an mcp-session-id is returned by the server during initialization,
	//clients using the Streamable HTTP transport MUST include it
	//in the mcp-session-id header on all of their subsequent HTTP requests.
	if !s.validateSession(res, req) {
		return
	}
	if !s.validateProtocolVersion(res, req) {
		return
	}
	//Handle resumability: check for last-event-id header
	if s.eventStore != nil {
		lastEventID := shared.EventID(req.Header.Get(shared.TRANSPORT_HEADER_LAST_EVENT_ID))
		if lastEventID != "" {
			s.replayEvents(lastEventID, res)
			return
		}
	}

	//Check if there's already an active standalone SSE stream for this session
	if _, ok := s.streamMapping.Get(s.standaloneSseStreamID); ok {
		//Only one GET SSE stream is allowed per session
		err := res.WriteJSON(http.StatusConflict, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_CONNECTION_CLOSED,
				Message: "Conflict: Only one SSE stream is allowed per session"},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Conflict: Only one SSE stream is allowed per session %v", err))
			return
		}
		return
	}

	//The server MUST either return Content-Type: text/event-stream in response to this HTTP GET,
	//or else return HTTP 405 Method Not Allowed
	headers := map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache, no-transform",
		"Connection":    "keep-alive",
	}

	//After initialization, always include the session ID if we have one
	if s.SessionID != "" {
		headers["mcp-session-id"] = s.SessionID
	}

	//We need to send headers immediately as messages will arrive much later,
	//otherwise the client will just wait for the first message
	for k, v := range headers {
		res.Writer().Header().Set(k, v)
	}
	//Sends headers to the client
	res.Writer().WriteHeader(http.StatusOK)
	if f, ok := res.Writer().(http.Flusher); ok {
		//Flushes the buffered data (including headers)
		f.Flush()
	}

	//Assign the response to the standalone SSE stream
	s.streamMapping.Set(s.standaloneSseStreamID, res)
	//Set up close handler for client disconnects
	go func() {
		<-req.Context().Done()
		s.streamMapping.Delete(s.standaloneSseStreamID)
	}()
}

//Handles POST requests containing JSON-RPC messages
func (s *StreamableHTTPServerTransport) handlePostRequest(res ResponseWriter, req *http.Request) {
	//Validate the Accept header
	acceptHeader := req.Header.Get("Accept")
	//The client MUST include an Accept header, listing both application/json and text/event-stream as supported content types.
	if !strings.Contains(acceptHeader, "application/json") || !strings.Contains(acceptHeader, "text/event-stream") {
		err := res.WriteJSON(http.StatusNotAcceptable, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_CONNECTION_CLOSED,
				Message: "Not Acceptable: Client must accept both application/json and text/event-stream"},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Not Acceptable: Client must accept both application/json and text/event-stream %v", err))
			return
		}
		return
	}

	ctHeader := req.Header.Get("content-type")
	if ctHeader == "" || !strings.Contains(ctHeader, "application/json") {
		err := res.WriteJSON(http.StatusUnsupportedMediaType, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_CONNECTION_CLOSED,
				Message: "Unsupported Media Type: Content-Type must be application/json"},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Unsupported Media Type: Content-Type must be application/json %v", err))
			return
		}
		return
	}

	authInfo := shared.GetAuthInfoRequest(req)
	requestInfo := shared.RequestInfo{
		Headers: req.Header,
	}
	var messages []types.RawMessage

	req.Body = http.MaxBytesReader(res.Writer(), req.Body, MAXIMUM_MESSAGE_SIZE)
	if err := json.NewDecoder(req.Body).Decode(&messages); err != nil {
		err := res.WriteJSON(http.StatusBadRequest, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_CONNECTION_CLOSED,
				Message: "Bad request: invalid body"},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Bad request: invalid body %v", err))
			return
		}
		return
	}

	//Check if this is an initialization request
	//https://spec.modelcontextprotocol.io/specification/2025-03-26/basic/lifecycle/
	if types.MessagesHasSomeInitializeRequest(messages) {
		//If it's a server with session management and the session ID is already set we should reject the request
		//to avoid re-initialization.
		if s.initialized && s.SessionID != "" {
			err := res.WriteJSON(http.StatusBadRequest, types.JSONRPCError{
				JSONRPC: types.JSONRPC_VERSION,
				Error: &types.Error{
					Code:    types.ERROR_CODE_INVALID_REQUEST,
					Message: "Invalid Request: Server already initialized"},
			})
			if err != nil {
				s.OnError(fmt.Errorf("res.WriteJSON Invalid Request: Server already initialized %v", err))
				return
			}
			return
		}
		if len(messages) > 0 {
			err := res.WriteJSON(http.StatusBadRequest, types.JSONRPCError{
				JSONRPC: types.JSONRPC_VERSION,
				Error: &types.Error{
					Code:    types.ERROR_CODE_INVALID_REQUEST,
					Message: "Invalid Request: Only one initialization request is allowed"},
			})
			if err != nil {
				s.OnError(fmt.Errorf("res.WriteJSON Invalid Request: Only one initialization request is allowed %v", err))
				return
			}
			return
		}
		s.SessionID = s.sessionIDGenerator()
		s.initialized = true

		//If we have a session ID and an onSessionInitialized handler, call it immediately
		//This is needed in cases where the server needs to keep track of multiple sessions
		if s.SessionID != "" && s.onSessionInitialized != nil {
			s.onSessionInitialized(s.SessionID)
		}

	} else {
		//If not a initialization request
		//
		//If an mcp-session-id is returned by the server during initialization,
		//clients using the Streamable HTTP transport MUST include it
		//in the mcp-session-id header on all of their subsequent HTTP requests.
		if !s.validateSession(res, req) {
			return
		}
		// mcp-protocol-version header is required for all requests after initialization.
		if !s.validateProtocolVersion(res, req) {
			return
		}
	}

	//check if it contains requests
	isJSONRPCRequest := types.MessagesHasSomeJSONRPCRequest(messages)

	parseError := func(gError error) {
		err := res.WriteJSON(http.StatusBadRequest, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_PARSE_ERROR,
				Message: "Parse error",
				Data:    gError,
			},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Parse globalError: %v error %v", gError, err))
		}
		s.OnError(gError)
	}

	if !isJSONRPCRequest {
		//if it only contains notifications or responses, return 202
		res.Writer().WriteHeader(http.StatusAccepted)
		//handle each message
		for _, msg := range messages {
			m, err := msg.ToJSONRPCMessage()
			if err != nil {
				parseError(fmt.Errorf("utils.MessageToJSONRPCMessage %v", err))
				return
			}
			s.OnMessage(m, &shared.MessageExtraInfo{AuthInfo: &authInfo, RequestInfo: &requestInfo})
		}
		return
	} else {
		//The default behavior is to use SSE streaming
		//but in some cases server will return JSON responses
		streamID := shared.StreamID(uuid.New().String())
		if s.enableJSONResponse {
			headers := map[string]string{
				"Content-Type":  "text/event-stream",
				"Cache-Control": "no-cache",
				"Connection":    "keep-alive",
			}
			//After initialization, always include the session ID if we have one
			if s.SessionID != "" {
				headers["mcp-session-id"] = s.SessionID
			}

			for k, v := range headers {
				res.Writer().Header().Set(k, v)
			}
			//Sends headers to the client
			res.Writer().WriteHeader(http.StatusOK)
		}
		//Store the response for this request to send messages back through this connection
		//We need to track by request ID to maintain the connection
		for _, msg := range messages {
			msgReq, err := msg.ToJSONRPCRequest()
			if err != nil {
				parseError(fmt.Errorf("utils.MessageToJSONRPCMessage %v", err))
				return
			}
			if msgReq == nil {
				continue //Skip if it's not a JSONRPCRequest
			}
			s.streamMapping.Set(streamID, res)
			s.requestToStreamMapping.Set(msgReq.ID, streamID)
		}

		//Set up close handler for client disconnects
		go func() {
			<-req.Context().Done()
			s.streamMapping.Delete(streamID)
		}()

		//handle each message
		for _, msg := range messages {
			jsonrpc, err := msg.ToJSONRPCMessage()
			if err != nil {
				parseError(fmt.Errorf("msg.ToJSONRPCMessage %v", err))
				return
			}
			s.OnMessage(jsonrpc, &shared.MessageExtraInfo{AuthInfo: &authInfo, RequestInfo: &requestInfo})
		}
		//The server SHOULD NOT close the SSE stream before sending all JSON-RPC responses
		//This will be handled by the send() method when responses are ready
	}

}

//Handles DELETE requests to terminate sessions
func (s *StreamableHTTPServerTransport) handleDeleteRequest(res ResponseWriter, req *http.Request) {
	if !s.validateSession(res, req) {
		return
	}
	if !s.validateProtocolVersion(res, req) {
		return
	}
	s.onSessionClosed(s.SessionID)
	err := s.Close()
	if err != nil {
		err := res.WriteJSON(http.StatusInternalServerError, types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			Error: &types.Error{
				Code:    types.ERROR_CODE_INTERNAL_ERROR,
				Message: "Error on close",
			},
		})
		if err != nil {
			s.OnError(fmt.Errorf("res.WriteJSON Error on close %v", err))
		}
		return
	}
	res.Writer().WriteHeader(http.StatusOK)
}

//Handles unsupported requests (PUT, PATCH, etc.)
func (s *StreamableHTTPServerTransport) handleUnsupportedRequest(res ResponseWriter) {
	res.Writer().Header().Set("Allow", "GET, POST, DELETE")
	err := res.WriteJSON(http.StatusMethodNotAllowed, types.JSONRPCError{
		JSONRPC: types.JSONRPC_VERSION,
		Error: &types.Error{
			Code:    types.ERROR_CODE_METHOD_NOT_ALLOWED,
			Message: "HTTP method not allowed",
		},
	})
	if err != nil {
		s.OnError(fmt.Errorf("res.WriteJSON HTTP method not allowed %v", err))
	}
}

//Sends a JSON-RPC message (request or response).
//
//If present, `relatedRequestId` is used to indicate to the transport which incoming request to associate this outgoing message with.
func (s *StreamableHTTPServerTransport) Send(msg types.JSONRPCMessage, opts *shared.TransportSendOptions) (*types.JSONRPCResponse, error) {
	requestID := opts.RelatedRequestID
	_, isJSONRPCResponse := msg.(*types.JSONRPCResponse)
	_, isJSONRPCError := msg.(*types.JSONRPCError)
	switch msgRes := msg.(type) {
	case *types.JSONRPCResponse:
	case *types.JSONRPCError:
		//If the message is a response, use the request ID from the message
		requestID = msgRes.GetRequestID()
	}

	//Check if this message should be sent on the standalone SSE stream (no request ID)
	//Ignore notifications from tools (which have relatedRequestId set)
	//Those will be sent via dedicated response SSE streams
	if requestID == 0 {
		//For standalone SSE streams, we can only send requests and notifications
		if isJSONRPCResponse || isJSONRPCError {
			return nil, fmt.Errorf("cannot send a response on a standalone SSE stream unless resuming a previous client request")
		}
		standaloneSSEResp, okStandaloneSSE := s.streamMapping.Get(s.standaloneSseStreamID)
		if !okStandaloneSSE {
			//The spec says the server MAY send messages on the stream, so it's ok to discard if no stream
			return nil, nil
		}
		//Generate and store event ID if event store is provided
		var eventID shared.EventID
		var err error
		if s.eventStore != nil {
			//Stores the event and gets the generated event ID
			eventID, err = s.eventStore.StoreEvent(s.standaloneSseStreamID, msg)
			if err != nil {
				return nil, fmt.Errorf("s.eventStore.StoreEvent %v", err)
			}
		}
		//Send the message to the standalone SSE stream
		if !s.writeSSEEvent(standaloneSSEResp, msg, eventID) {
			return nil, fmt.Errorf("s.writeSSEEvent return false eventID: %v", eventID)
		}
		return nil, nil
	}

	//Get the response for this request
	streamID, okStream := s.requestToStreamMapping.Get(requestID)
	if !okStream {
		return nil, fmt.Errorf("no connection established for request ID: %v", requestID)
	}
	responseW, okResponseW := s.streamMapping.Get(streamID)
	if !okResponseW {
		return nil, fmt.Errorf("response writer not found for streamID: %v", streamID)
	}

	if !s.enableJSONResponse {
		//For SSE responses, generate event ID if event store is provided
		var eventID shared.EventID
		var err error
		if s.eventStore != nil {
			//Stores the event and gets the generated event ID
			eventID, err = s.eventStore.StoreEvent(streamID, msg)
			if err != nil {
				return nil, fmt.Errorf("s.eventStore.StoreEvent %v", err)
			}
		}
		//Write the event to the response stream
		if !s.writeSSEEvent(responseW, msg, eventID) {
			return nil, fmt.Errorf("s.writeSSEEvent return false eventID: %v", eventID)
		}
	}
	if isJSONRPCResponse || isJSONRPCError {
		var allRequestID []types.RequestID
		var countRequest int
		s.requestResponseMap.Set(requestID, msg)
		allRequestToStream := s.requestToStreamMapping.GetAll()
		for rID, sID := range allRequestToStream {
			if sID != streamID {
				continue
			}
			allRequestID = append(allRequestID, rID)
			_, okRes := s.requestResponseMap.Get(rID)
			if !okRes {
				continue
			}
			countRequest++
		}

		allResponsesReady := len(allRequestID) == countRequest
		if !allResponsesReady {
			return nil, nil
		}
		if s.enableJSONResponse {
			//All responses ready, send as JSON
			headers := map[string]string{
				"Content-Type": "text/event-stream",
			}
			if s.SessionID != "" {
				headers["mcp-session-id"] = s.SessionID
			}

			var respMsg []types.JSONRPCMessage
			for _, reqID := range allRequestID {
				res, ok := s.requestResponseMap.Get(reqID)
				if !ok {
					continue
				}
				respMsg = append(respMsg, res)
			}

			if len(respMsg) == 1 {
				responseW.WriteJSON(http.StatusOK, respMsg[0])
			} else {
				responseW.WriteJSON(http.StatusOK, respMsg)
			}
		} else {
			//End the SSE stream
			if f, ok := responseW.Writer().(http.Flusher); ok {
				//Flushes the buffered data (including headers)
				f.Flush()
			}
		}
		//Clean up
		for _, rID := range allRequestID {
			s.requestResponseMap.Delete(rID)
			s.requestToStreamMapping.Delete(rID)
		}
	}
	return nil, nil
}

//Closes the connection.
func (s *StreamableHTTPServerTransport) Close() error {
	//Close all SSE connections
	sm := s.streamMapping.GetAll()
	for _, r := range sm {
		if f, ok := r.Writer().(http.Flusher); ok {
			//Flushes the buffered data (including headers)
			f.Flush()
		}
	}
	//Clear any pending responses
	s.streamMapping.Clear()

	err := s.OnClose()
	if err != nil {
		s.OnError(fmt.Errorf("OnClose Error %v", err))
	}
	return nil
}

//Callback for when the connection is closed for any reason.
//
//This should be invoked when close() is called as well.
//
//Always execute first the prop globalOnClose if is defined
func (s *StreamableHTTPServerTransport) OnClose() error {
	if s.globalOnClose != nil {
		s.globalOnClose()
	}
	return nil
}

//Callback for when an error occurs.
//
//Note that errors are not necessarily fatal; they are used for reporting any kind of exceptional condition out of band.
//
//Always execute first the prop globalOnError if is defined
func (s *StreamableHTTPServerTransport) OnError(err error) {
	if s.globalOnError != nil {
		s.globalOnError(err)
	}
}

//Callback for when a message (request or response) is received over the connection.
//
//Includes the authInfo if the transport is authenticated.
//
//Always execute first prop globalOnMessage if is defined
func (s *StreamableHTTPServerTransport) OnMessage(message types.JSONRPCMessage, extra *shared.MessageExtraInfo) {
	if s.globalOnMessage != nil {
		s.globalOnMessage(message, extra)
	}
}

//Return true if the transport has already started
func (s *StreamableHTTPServerTransport) IsStarted() bool {
	return s.started
}

//Sets the protocol version used for the connection (called when the initialize response is received).
func (s *StreamableHTTPServerTransport) SetProtocolVersion(version string) {
	s.protocolVersion = version
}

//Return the session ID
func (s *StreamableHTTPServerTransport) GetSessionID() string {
	return s.SessionID
}

//Set this if globalOnClose is needed, this must be executed into OnClose Func first
func (s *StreamableHTTPServerTransport) SetGlobalOnClose(f func()) {
	s.globalOnClose = f
}

//Set this if globalOnError is needed, this must be executed into OnError Func first
func (s *StreamableHTTPServerTransport) SetGlobalOnError(f func(err error)) {
	s.globalOnError = f
}

//Set this if globalOnMessage is needed, this must be executed into OnMessage Func first
func (s *StreamableHTTPServerTransport) SetGlobalOnMessage(f func(message types.JSONRPCMessage, extra *shared.MessageExtraInfo)) {
	s.globalOnMessage = f
}
