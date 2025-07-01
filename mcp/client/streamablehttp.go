package client

import (
	"context"

	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
)

type StreamableHTTPServerTransport struct {
	//The session ID generated for this connection.
	SessionID string `json:"sessionId,omitempty"`
}

//Starts processing messages on the transport, including any connection steps that might need to be taken.
//
//This method should only be called after callbacks are installed, or else messages may be lost.
func (t *StreamableHTTPServerTransport) Start(ctx context.Context) error {
	return nil
}

//Sends a JSON-RPC message (request or response).
//
//If present, `relatedRequestId` is used to indicate to the transport which incoming request to associate this outgoing message with.
func (t *StreamableHTTPServerTransport) Send(ctx context.Context, request types.JSONRPCMessage, options *shared.TransportSendOptions) (*types.JSONRPCResponse, error) {
	return nil, nil
}

//Closes the connection.
func (t *StreamableHTTPServerTransport) Close() error {
	return nil
}

//Callback for when the connection is closed for any reason.
//
//This should be invoked when close() is called as well.
func (t *StreamableHTTPServerTransport) OnClose() error {
	return nil
}

//Callback for when an error occurs.
//
//Note that errors are not necessarily fatal; they are used for reporting any kind of exceptional condition out of band.
func (t *StreamableHTTPServerTransport) OnError(error) {
	//Optional
}

//Callback for when a message (request or response) is received over the connection.
//
//Includes the authInfo if the transport is authenticated.
func (t *StreamableHTTPServerTransport) OnMessage(message types.JSONRPCMessage, extra interface{}) {
	//Optional
}

//Sets the protocol version used for the connection (called when the initialize response is received).
func (t *StreamableHTTPServerTransport) SetProtocolVersion(version string) {
	//Optional
}
