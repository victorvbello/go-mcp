package shared

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/victorvbello/gomcp/mcp/types"
	utils "github.com/victorvbello/gomcp/mcp/utils/logger"
)

const (
	//The default request timeout, in miliseconds.
	DEFAULT_REQUEST_TIMEOUT_MSEC = 60 * time.Millisecond
)

var ErrCancelFunNotFound = fmt.Errorf("cancel function not found")

type Protocol struct {
	transport            Transport
	requestMessageID     *muxRequestMessageID
	requestHandlerCancel *muxMapRequestHandlerCancel
	requestHandlers      *muxMapRequestHandlers
	responseHandlers     *muxMapResponseHandlers
	notificationHandlers *muxMapNotificationHandlers
	progressHandlers     *muxMapProgressHandlers
	timeoutInfo          *muxMapTimeoutConfig
	logger               utils.LogService
	options              *ProtocolOptions
	//Callback for when the connection is closed for any reason.
	//
	//This is invoked when close() is called as well.
	OnClose func() error
	//Callback for when an error occurs.
	//
	//Note that errors are not necessarily fatal; they are used for reporting any kind of exceptional condition out of band.
	OnError func(err error) error
	//A handler to invoke for any request types that do not have their own handler installed.
	FallbackRequestHandler requestHandler
	//A handler to invoke for any notification types that do not have their own handler installed.
	FallbackNotificationHandler notificationHandler
	//A method to check if a capability is supported by the remote side, for the given method to be called.
	//
	//This should be implemented by parent struct
	AssertCapabilityForMethod func(method string) error
	//A method to check if a notification is supported by the local side, for the given method to be sent.
	//
	//This should be implemented by parent struct
	AssertNotificationCapability func(method string) error
	//A method to check if a request handler is supported by the local side, for the given method to be handled.
	//
	//This should be implemented by parent struct
	AssertRequestHandlerCapability func(method string) error
}

func NewProtocol(opts *ProtocolOptions) *Protocol {
	newProtocol := &Protocol{
		requestMessageID:     new(muxRequestMessageID),
		requestHandlerCancel: new(muxMapRequestHandlerCancel),
		requestHandlers:      new(muxMapRequestHandlers),
		responseHandlers:     new(muxMapResponseHandlers),
		notificationHandlers: new(muxMapNotificationHandlers),
		progressHandlers:     new(muxMapProgressHandlers),
		logger:               utils.NewLoggerService(),
		options:              opts,
	}
	newProtocol.logger = utils.NewLoggerService()
	newProtocol.SetNotificationHandler(types.NewCancelledNotification(nil), func(notification types.NotificationInterface) error {
		notify := notification.(*types.CancelledNotification)
		CancelFunc, ok := newProtocol.requestHandlerCancel.Get(notify.Params.RequestID)
		if !ok {
			return ErrCancelFunNotFound
		}
		newProtocol.logger.Info(nil, notify.Params.Reason)
		CancelFunc()
		return nil
	})
	newProtocol.SetNotificationHandler(types.NewProgressNotification(nil), func(notification types.NotificationInterface) error {
		notify := notification.(*types.ProgressNotification)
		newProtocol.onProgress(notify)
		return nil
	})

	newProtocol.SetRequestHandler(types.NewPingRequest(), func(request types.RequestInterface, extra *RequestHandlerExtra) (types.ResultInterface, error) {
		// Automatic pong by default.
		return &types.EmptyResult{}, nil
	})

	return newProtocol
}

//Add timeout to timeoutInfo list by msg id
func (p *Protocol) setupTimeout(messageID int, timeout *timeoutConfig) {
	p.timeoutInfo.Set(messageID, timeout)
}

//Reset the timeout by validating the MaxTotalTimeout
func (p *Protocol) resetTimeout(messageID int) (bool, types.ErrorInterface) {
	timeout, ok := p.timeoutInfo.Get(messageID)
	if !ok {
		return false, nil
	}
	totalElapsed := time.Until(timeout.StartTime) * -1
	if timeout.MaxTotalTimeout > 0 && totalElapsed >= timeout.MaxTotalTimeout {
		return false, types.NewMcpError(
			types.ERROR_CODE_REQUEST_TIMEOUT,
			"Maximum total timeout exceeded", map[string]interface{}{
				"maxTotalTimeout": timeout.MaxTotalTimeout,
				"totalElapsed":    totalElapsed})
	}
	timeout.Clear()
	//Recreate timeout whit same timeout
	timeout.Start()
	return true, nil
}

func (p *Protocol) cleanupTimeout(messageID int) {
	timeout, ok := p.timeoutInfo.Get(messageID)
	if !ok {
		return
	}
	timeout.Clear()
	p.timeoutInfo.Delete(messageID)
}

//Attaches to the given transport, starts it, and starts listening for messages.
//
//The Protocol object assumes ownership of the Transport, replacing any callbacks that have already been set, and expects that it is the only user of the Transport instance going forward.
func (p *Protocol) Connect(transport Transport) {
	p.transport = transport
	p.transport.SetGlobalOnClose(func() {
		p.onClose()
	})
	p.transport.SetGlobalOnError(func(err error) {
		p.onError(err)
	})
	p.transport.SetGlobalOnMessage(func(message types.JSONRPCMessage, extra interface{}) {
		switch msg := message.(type) {
		case types.JSONRPCGeneralResponse:
			p.onResponse(msg)
		case *types.JSONRPCRequest:
			p.onRequest(msg, extra.(*RequestHandlerExtra))
		case *types.JSONRPCNotification:
			p.onNotification(msg)
		default:
			p.onError(fmt.Errorf("unknown message type %T", msg))
		}
	})

	err := p.transport.Start()
	if err != nil {
		p.onError(fmt.Errorf("transport.Start %v", err))
	}
}

func (p *Protocol) onClose() {
	responseHandlers := p.responseHandlers.GetAll()
	p.responseHandlers.Clear()
	p.progressHandlers.Clear()
	p.transport = nil
	p.OnClose()

	globalError := types.NewMcpError(types.ERROR_CODE_CONNECTION_CLOSED, "Connection closed", nil)
	jsonError := &types.JSONRPCError{Error: globalError}
	for _, handler := range responseHandlers {
		handler(jsonError)
	}
}

func (p *Protocol) onError(err error) {
	if p.OnError == nil {
		return
	}
	p.OnError(err)
}

func (p *Protocol) onNotification(notification *types.JSONRPCNotification) {
	handlerType := "notificationHandlers"
	handler, ok := p.notificationHandlers.Get(notification.Notification.GetNotification().Method)
	if !ok {
		handlerType = "fallbackNotificationHandler"
		handler = p.FallbackNotificationHandler
	}
	if handler == nil {
		return
	}
	err := handler(notification.Notification)
	if err != nil {
		p.onError(fmt.Errorf("uncaught error in notification handler[%s] %v %v", handlerType, err, notification))
	}
}

func (p *Protocol) onRequest(request *types.JSONRPCRequest, extra *RequestHandlerExtra) {
	handlerType := "requestHandlers"
	handler, ok := p.requestHandlers.Get(request.Method)
	if !ok {
		handlerType = "fallbackRequestHandler"
		handler = p.FallbackRequestHandler
	}
	if handler == nil {
		_, err := p.transport.Send(&types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			ID:      request.ID,
			Error: &types.Error{
				Code:    types.ERROR_CODE_METHOD_NOT_FOUND,
				Message: "Method not found",
			},
		}, nil)
		if err != nil {
			p.onError(fmt.Errorf("failed to send an error response %v", err))
		}
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	p.requestHandlerCancel.Set(request.ID, cancelFunc)
	defer func() {
		p.requestHandlerCancel.Delete(request.ID)
	}()
	extraRequestHandle := &RequestHandlerExtra{
		Context:   ctx,
		SessionID: p.transport.GetSessionID(),
		Meta:      request.Params.Metadata,
		AuthInfo:  extra.AuthInfo,
		RequestID: request.ID,
		SendNotification: func(notification types.NotificationInterface) {
			p.Notification(notification, &NotificationOptions{RelatedRequestID: request.ID})
		},
		SendRequest: func(req types.RequestInterface, opts RequestOptions) (types.ResultInterface, error) {
			opts.RelatedRequestID = request.ID
			return p.Request(req, opts)
		},
	}

	result, err := handler(request, extraRequestHandle)
	if ctx.Err() != nil {
		return
	}
	if err != nil {
		_, err := p.transport.Send(&types.JSONRPCError{
			JSONRPC: types.JSONRPC_VERSION,
			ID:      request.ID,
			Error: &types.Error{
				Code:    types.ERROR_CODE_INTERNAL_ERROR,
				Message: fmt.Sprintf("handler request error %s, [%v], %v ", handlerType, request, err),
			},
		}, nil)
		if err != nil {
			p.onError(fmt.Errorf("failed to send an error response %v", err))
			return
		}
	}

	_, err = p.transport.Send(&types.JSONRPCResponse{
		JSONRPC: types.JSONRPC_VERSION,
		ID:      request.ID,
		Result:  result,
	}, nil)
	if err != nil {
		p.onError(fmt.Errorf("failed to send result %v", err))
		return
	}

}

func (p *Protocol) onProgress(progressNotify *types.ProgressNotification) {
	messageID := progressNotify.Params.ProgressToken.(int)
	progressHandler, okProgressHandle := p.progressHandlers.Get(messageID)
	if !okProgressHandle {
		p.onError(fmt.Errorf("received a progress notification for an unknown token, progressHandlers not found: %v", progressNotify))
		return
	}

	responseHandler, okResponseHandler := p.responseHandlers.Get(messageID)
	timeout, okTimeoutInfo := p.timeoutInfo.Get(messageID)
	if okTimeoutInfo && okResponseHandler && timeout.ResetTimeoutOnProgress {
		_, err := p.resetTimeout(messageID)
		if err != nil {
			jsonError := &types.JSONRPCError{Error: err}
			responseHandler(jsonError)
			return
		}
	}

	err := progressHandler(progressNotify.Params.Progress)
	if err != nil {
		p.onError(fmt.Errorf("progressHandler %v %v", progressNotify, err))
	}
}

func (p *Protocol) onResponse(response types.JSONRPCGeneralResponse) {
	messageID := int(response.GetRequestID())
	responseHandler, okResponseHandler := p.responseHandlers.Get(messageID)
	if !okResponseHandler {
		p.onError(fmt.Errorf("received a response for an unknown message ID %v", response))
		return
	}
	p.responseHandlers.Delete(messageID)
	p.progressHandlers.Delete(messageID)
	p.cleanupTimeout(messageID)

	var err error
	var handlerType string
	switch resp := response.(type) {
	case *types.JSONRPCResponse:
		handlerType = fmt.Sprintf("%T", resp)
		err = responseHandler(resp)
	case *types.JSONRPCError:
		handlerType = fmt.Sprintf("%T", resp)
		globalError := types.NewMcpError(
			resp.Error.GetErrorCode(),
			resp.Error.GetErrorMessage(),
			resp.Error.GetErrorData(),
		)
		err = responseHandler(&types.JSONRPCError{Error: globalError})
	}
	if err != nil {
		p.onError(fmt.Errorf("responseHandler error in %s %v", handlerType, err))
	}
}

func (p *Protocol) GetTransport() Transport {
	return p.transport
}

func (p *Protocol) Close() {
	err := p.transport.Close()
	if err != nil {
		p.onError(fmt.Errorf("transport.Close %v", err))
	}
}

//Sends a request and wait for a response.
//
//Do not use this method to emit notifications! Use notification() instead.
func (p *Protocol) Request(request types.RequestInterface, opts RequestOptions) (gResp types.ResultInterface, gErr error) {
	if p.transport == nil {
		gErr = fmt.Errorf("transport not connected")
		return
	}
	if p.options.EnforceStrictCapabilities != nil && *p.options.EnforceStrictCapabilities {
		p.AssertCapabilityForMethod(request.GetRequest().Method)
	}
	if opts.Canceled() {
		gErr = fmt.Errorf("the request was canceled by an external close function")
		return
	}
	messageID := p.requestMessageID.Increase()
	jsonrpcRequest := &types.JSONRPCRequest{
		JSONRPC: types.JSONRPC_VERSION,
		ID:      types.RequestID(messageID),
		Request: request.GetRequest(),
	}
	if opts.Onprogress != nil {
		p.progressHandlers.Set(messageID, opts.Onprogress)
		jsonrpcRequest.Request.Params = &types.RequestParams{
			Metadata: &types.MetadataRequest{
				ProgressToken: messageID,
			},
			AdditionalProperties: request.GetRequest().Params.AdditionalProperties,
		}
	}
	type requestReturnChan struct {
		r types.ResultInterface
		e error
	}
	returnChan := make(chan requestReturnChan)
	defer func() {
		resultReturn := <-returnChan
		if gResp == nil {
			gResp = resultReturn.r
		}
		if gErr == nil {
			gErr = resultReturn.e
		}
	}()

	cancelFlow := func(reason types.ErrorInterface) {
		var once sync.Once
		once.Do(func() {
			p.responseHandlers.Delete(messageID)
			p.progressHandlers.Delete(messageID)
			p.cleanupTimeout(messageID)

			_, err := p.transport.Send(types.NewCancelledNotification(&types.CancelledNotificationParams{
				RequestID: types.RequestID(messageID),
				Reason:    reason.ToError().Error(),
			}), &TransportSendOptions{
				RelatedRequestID:  opts.RelatedRequestID,
				ResumptionToken:   opts.ResumptionToken,
				OnResumptionToken: opts.OnResumptionToken,
			})
			if err != nil {
				p.onError(fmt.Errorf("failed to send cancellation: %v", err))
			}
			returnChan <- requestReturnChan{
				e: reason.ToError(),
			}
		})
	}

	if opts.Context != nil {
		//If RequestOptions has a cancel context
		if _, ok := opts.Context.Deadline(); ok {
			go func() {
				//When the cancel function was called
				<-opts.Context.Done()
				cancelFlow(&types.Error{Message: "context was canceled from outside"})
			}()
		}
	}

	p.responseHandlers.Set(messageID, func(response types.JSONRPCGeneralResponse) error {
		if opts.Canceled() {
			return nil
		}
		if err, ok := response.(*types.JSONRPCError); ok {
			lErr := err.Error.ToError()
			returnChan <- requestReturnChan{
				e: lErr,
			}
			return lErr
		}
		if res, ok := response.(*types.JSONRPCResponse); ok {
			returnChan <- requestReturnChan{
				r: res,
			}
			return nil
		}
		err := fmt.Errorf("invalid response type")
		returnChan <- requestReturnChan{
			e: err,
		}
		return err
	})

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DEFAULT_REQUEST_TIMEOUT_MSEC
	}

	timeoutHandler := func() {
		cancelFlow(types.NewMcpError(
			types.ERROR_CODE_REQUEST_TIMEOUT,
			"request timed out", map[string]interface{}{"timeout": timeout}))

	}

	var resetTimeoutOnProgress bool
	if opts.ResetTimeoutOnProgress != nil {
		resetTimeoutOnProgress = *opts.ResetTimeoutOnProgress
	}

	p.setupTimeout(messageID, &timeoutConfig{
		Timeout:                time.Duration(timeout),
		MaxTotalTimeout:        opts.MaxTotalTimeout,
		OnTimeout:              timeoutHandler,
		ResetTimeoutOnProgress: resetTimeoutOnProgress,
	})

	res, err := p.transport.Send(jsonrpcRequest,
		&TransportSendOptions{
			RelatedRequestID:  opts.RelatedRequestID,
			ResumptionToken:   opts.ResumptionToken,
			OnResumptionToken: opts.OnResumptionToken,
		})

	if err != nil {
		p.cleanupTimeout(messageID)
		returnChan <- requestReturnChan{
			e: err,
		}
		return
	}

	returnChan <- requestReturnChan{
		r: res,
	}
	return
}

//Emits a notification, which is a one-way message that does not expect a response.
func (p *Protocol) Notification(notification types.NotificationInterface, opts *NotificationOptions) error {
	if p.transport == nil {
		return fmt.Errorf("transport not connected")
	}
	err := p.AssertNotificationCapability(notification.GetNotification().Method)
	if err != nil {
		return fmt.Errorf("assertNotificationCapability error: %v", err)
	}
	jsonrpcNotification := &types.JSONRPCNotification{
		JSONRPC:      types.JSONRPC_VERSION,
		Notification: notification,
	}

	_, err = p.transport.Send(jsonrpcNotification, &TransportSendOptions{RelatedRequestID: opts.RelatedRequestID})
	if err != nil {
		return fmt.Errorf("transport.Send error: %v", err)
	}
	return nil
}

//Registers a handler to invoke when this protocol object receives a request with the given method.
//
//Note that this will replace any previous request handler for the same method.
func (p *Protocol) SetRequestHandler(request types.RequestInterface, handler requestHandler) {
	method := request.GetRequest().Method
	err := p.AssertRequestHandlerCapability(method)
	if err != nil {
		detailErr := fmt.Errorf("assertRequestHandlerCapability %v", err)
		p.onError(detailErr)
	}

	if !p.AssertCanSetRequestHandler(method) {
		detailErr := fmt.Errorf("method %s has been registered", method)
		p.onError(detailErr)
	}

	p.requestHandlers.Set(method, handler)
}

//Removes the request handler for the given method.
func (p *Protocol) RemoveRequestHandler(method string) {
	p.requestHandlers.Delete(method)
}

//Asserts that a request handler has not already been set for the given method, in preparation for a new one being automatically installed.
func (p *Protocol) AssertCanSetRequestHandler(method string) bool {
	_, ok := p.requestHandlers.Get(method)
	return !ok
}

//Registers a handler to invoke when this protocol object receives a notification with the given method.
//
//Note that this will replace any previous notification handler for the same method.
func (p *Protocol) SetNotificationHandler(notification types.NotificationInterface, handler notificationHandler) {
	p.notificationHandlers.Set(notification.GetNotification().Method, handler)
}

//Removes the notification handler for the given method.
func (p *Protocol) RemoveNotificationHandler(method string) {
	p.notificationHandlers.Delete(method)
}
