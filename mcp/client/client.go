package client

import (
	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
)

type Client struct {
	*shared.Protocol
	onErrorCallBack func(err error)
}

//ProtocolInterface Methods

func (c *Client) onInitializeProtocolInterfaceType() int {
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
	return nil
}

//Add external Action on error
func (c *Client) SetOnErrorCallBack(fn func(err error)) {
	c.onErrorCallBack = fn
}

//A handler to invoke for any request types that do not have their own handler installed.
func (c *Client) FallbackRequestHandler() shared.RequestHandler {
	return func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
		return nil, nil
	}
}

//A handler to invoke for any notification types that do not have their own handler installed.
func (c *Client) FallbackNotificationHandler() shared.NotificationHandler {
	return func(notification types.NotificationInterface) error {
		return nil
	}
}

//A method to check if a capability is supported by the remote side, for the given method to be called.
//
//This should be implemented by parent struct
func (c *Client) AssertCapabilityForMethod(req types.RequestInterface) error {
	return nil
}

//A method to check if a notification is supported by the local side, for the given method to be sent.
//
//This should be implemented by parent struct
func (c *Client) AssertNotificationCapability(notify types.NotificationInterface) error {
	return nil
}

//A method to check if a request handler is supported by the local side, for the given method to be handled.
//
//This should be implemented by parent struct
func (c *Client) AssertRequestHandlerCapability(req types.RequestInterface) error {
	return nil
}

//Client Methods
