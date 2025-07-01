package shared

import (
	"github.com/victorvbello/gomcp/mcp/types"
)

type TransportSendOptions struct {
	//If present, `relatedRequestId` is used to indicate to the transport which incoming request to associate this outgoing message with.
	RelatedRequestID types.RequestID `json:"relatedRequestId,omitempty"`
	//The resumption token used to continue long-running requests that were interrupted.
	//
	//This allows clients to reconnect and continue from where they left off, if supported by the transport.
	ResumptionToken string `json:"resumptionToken,omitempty"`
	//A callback that is invoked when the resumption token changes, if supported by the transport.
	//
	//This allows clients to persist the latest token for potential reconnection.
	OnResumptionToken func(string) `json:"onresumptiontoken,omitempty"`
}

type Transport interface {
	//Starts processing messages on the transport, including any connection steps that might need to be taken.
	//
	//This method should only be called after callbacks are installed, or else messages may be lost.
	//
	//NOTE: This method should not be called explicitly when using Client, Server, or Protocol classes, as they will implicitly call start().
	Start() error
	//Sends a JSON-RPC message (request or response).
	//
	//If present, `relatedRequestId` is used to indicate to the transport which incoming request to associate this outgoing message with.
	Send(request types.JSONRPCMessage, options *TransportSendOptions) (*types.JSONRPCResponse, error)
	//Closes the connection.
	Close() error
	//Callback for when the connection is closed for any reason.
	//
	//This should be invoked when close() is called as well.
	//
	//Always execute in defer flow, if the prop globalOnClose if is defined
	OnClose() error
	//Callback for when an error occurs.
	//
	//Note that errors are not necessarily fatal; they are used for reporting any kind of exceptional condition out of band.
	//
	//Always execute in defer flow, if the prop globalOnError if is defined
	OnError(error)
	//Callback for when a message (request or response) is received over the connection.
	//
	//Includes the authInfo if the transport is authenticated.
	//
	//Always execute in defer flow, if the prop globalOnMessage if is defined
	OnMessage(message types.JSONRPCMessage, extra interface{})
	//Sets the protocol version used for the connection (called when the initialize response is received).
	SetProtocolVersion(version string)
	//Set this if globalOnClose is needed, this must be executed into OnClose Func first
	SetGlobalOnClose(func())
	//Set this if globalOnError is needed, this must be executed into OnError Func first
	SetGlobalOnError(func(err error))
	//Set this if globalOnMessage is needed, this must be executed into OnMessage Func first
	SetGlobalOnMessage(func(message types.JSONRPCMessage, extra interface{}))
	//Return true if the transport has already started
	IsStarted() bool
	//Return the session ID
	GetSessionID() string
}
