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
	logger               utils.LogService
	owner                ProtocolInterface
	transport            Transport
	requestMessageID     *muxRequestMessageID
	requestHandlerCancel *muxMapRequestHandlerCancel
	requestHandlers      *muxMapRequestHandlers
	responseHandlers     *muxMapResponseHandlers
	notificationHandlers *muxMapNotificationHandlers
	progressHandlers     *muxMapProgressHandlers
	timeoutInfo          *muxMapTimeoutConfig
	options              *ProtocolOptions
}

func NewProtocol(opts *ProtocolOptions, pi ProtocolInterface) *Protocol {
	newProtocol := &Protocol{
		owner:                pi,
		requestMessageID:     new(muxRequestMessageID),
		requestHandlerCancel: newMuxMapRequestHandlerCancel(),
		requestHandlers:      newMuxMapRequestHandlers(),
		responseHandlers:     newMuxMapResponseHandlers(),
		notificationHandlers: newMuxMapNotificationHandlers(),
		progressHandlers:     newMuxMapProgressHandlers(),
		timeoutInfo:          newMuxMapTimeoutConfig(),
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
		//Automatic pong by default.
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
	p.transport.SetGlobalOnMessage(func(message types.JSONRPCMessage, extra *MessageExtraInfo) {
		switch msg := message.(type) {
		case types.JSONRPCGeneralResponse:
			p.onResponse(msg)
		case *types.JSONRPCRequest:
			p.onRequest(msg, extra)
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
	p.owner.OnClose()

	globalError := types.NewMcpError(types.ERROR_CODE_CONNECTION_CLOSED, "Connection closed", nil)
	jsonError := &types.JSONRPCError{Error: globalError}
	for _, handler := range responseHandlers {
		handler(jsonError)
	}
}

func (p *Protocol) onError(err error) {
	p.owner.OnError(err)
}

func (p *Protocol) onNotification(notification *types.JSONRPCNotification) {
	handlerType := "notificationHandlers"
	handler, ok := p.notificationHandlers.Get(notification.NotificationInterface.GetNotification().Method)
	if !ok {
		handlerType = "fallbackNotificationHandler"
		handler = p.owner.FallbackNotificationHandler()
	}
	if handler == nil {
		return
	}
	err := handler(notification.NotificationInterface)
	if err != nil {
		p.onError(fmt.Errorf("uncaught error in notification handler[%s] %v %v", handlerType, err, notification))
	}
}

func (p *Protocol) onRequest(request *types.JSONRPCRequest, extra *MessageExtraInfo) {
	handlerType := "requestHandlers"
	handler, ok := p.requestHandlers.Get(request.GetRequest().Method)
	if !ok {
		handlerType = "fallbackRequestHandler"
		handler = p.owner.FallbackRequestHandler()
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
			return
		}
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	p.requestHandlerCancel.Set(request.ID, cancelFunc)
	defer func() {
		p.requestHandlerCancel.Delete(request.ID)
	}()
	safeRequestParams := &types.BaseRequestParams{}
	if request.GetRequest().Params != nil {
		safeRequestParams = request.GetRequest().Params
	}
	RhExMeta, err := types.NewMetadataRequestFromMetadata(safeRequestParams.Metadata)
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
	safeExtra := &MessageExtraInfo{}
	if extra != nil {
		safeExtra = extra
	}
	extraRequestHandle := &RequestHandlerExtra{
		Context:     ctx,
		SessionID:   p.transport.GetSessionID(),
		Meta:        RhExMeta,
		AuthInfo:    safeExtra.AuthInfo,
		RequestID:   request.ID,
		RequestInfo: safeExtra.RequestInfo,
		SendNotification: func(notification types.NotificationInterface) {
			p.Notification(notification, &NotificationOptions{RelatedRequestID: request.ID})
		},
		SendRequest: func(req types.RequestInterface, opts *RequestOptions) (types.ResultInterface, error) {
			opts.RelatedRequestID = request.ID
			return p.Request(req, opts)
		},
	}
	result, err := handler(request.RequestInterface, extraRequestHandle)
	if err := ctx.Err(); err != nil {
		p.logger.Info(nil, fmt.Sprintf("context for method %s was closed %v", request.GetRequest().Method, err))
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
func (p *Protocol) Request(request types.RequestInterface, opts *RequestOptions) (gResp types.ResultInterface, gErr error) {
	safeOpts := opts
	if safeOpts == nil {
		safeOpts = &RequestOptions{}
	}
	if p.transport == nil {
		gErr = fmt.Errorf("transport not connected")
		return
	}
	if p.options.EnforceStrictCapabilities != nil && *p.options.EnforceStrictCapabilities {
		p.owner.AssertCapabilityForMethod(request)
	}
	if safeOpts.Canceled() {
		gErr = fmt.Errorf("the request was canceled by an external close function")
		return
	}

	messageID := p.requestMessageID.Increase()
	jsonrpcRequest := &types.JSONRPCRequest{
		JSONRPC:          types.JSONRPC_VERSION,
		ID:               types.RequestID(messageID),
		RequestInterface: request,
	}
	if safeOpts.Onprogress != nil {
		p.progressHandlers.Set(messageID, safeOpts.Onprogress)
		jsonrpcRequest.RequestInterface = &types.Request{
			Method: request.GetRequest().Method,
			Params: &types.BaseRequestParams{
				Metadata: map[string]interface{}{
					"progressToken": types.ProgressToken(messageID),
				},
				AdditionalProperties: request.GetRequest().Params.AdditionalProperties,
			}}
	}
	type requestReturnChan struct {
		r types.ResultInterface
		e error
	}
	returnChan := make(chan requestReturnChan)
	defer func() {
		fmt.Println("---- 1")
		select {
		case resultReturn := <-returnChan:
			gResp = resultReturn.r
			if gErr != nil {
				gErr = resultReturn.e
			}
		}
		fmt.Println("---- 2")
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
				RelatedRequestID:  safeOpts.RelatedRequestID,
				ResumptionToken:   safeOpts.ResumptionToken,
				OnResumptionToken: safeOpts.OnResumptionToken,
			})
			if err != nil {
				p.onError(fmt.Errorf("failed to send cancellation: %v", err))
			}
			go func() {
				returnChan <- requestReturnChan{
					e: reason.ToError(),
				}
			}()
		})
	}

	if safeOpts.Context != nil {
		//If RequestOptions has a cancel context
		if _, ok := safeOpts.Context.Deadline(); ok {
			go func() {
				//When the cancel function was called
				<-safeOpts.Context.Done()
				cancelFlow(&types.Error{Message: "context was canceled from outside"})
			}()
		}
	}

	p.responseHandlers.Set(messageID, func(response types.JSONRPCGeneralResponse) error {
		if safeOpts.Canceled() {
			return nil
		}
		if err, ok := response.(*types.JSONRPCError); ok {
			lErr := err.Error.ToError()
			go func() {
				returnChan <- requestReturnChan{
					e: lErr,
				}
			}()
			return lErr
		}
		if res, ok := response.(*types.JSONRPCResponse); ok {
			go func() {
				returnChan <- requestReturnChan{
					r: res.Result,
				}
			}()
			return nil
		}
		err := fmt.Errorf("invalid response type")
		go func() {
			returnChan <- requestReturnChan{
				e: err,
			}
		}()
		return err
	})

	timeout := safeOpts.Timeout
	if timeout == 0 {
		timeout = DEFAULT_REQUEST_TIMEOUT_MSEC
	}

	timeoutHandler := func() {
		cancelFlow(types.NewMcpError(
			types.ERROR_CODE_REQUEST_TIMEOUT,
			"request timed out", map[string]interface{}{"timeout": timeout}))

	}

	var resetTimeoutOnProgress bool
	if safeOpts.ResetTimeoutOnProgress != nil {
		resetTimeoutOnProgress = *safeOpts.ResetTimeoutOnProgress
	}
	p.setupTimeout(messageID, &timeoutConfig{
		Timeout:                time.Duration(timeout),
		MaxTotalTimeout:        safeOpts.MaxTotalTimeout,
		OnTimeout:              timeoutHandler,
		ResetTimeoutOnProgress: resetTimeoutOnProgress,
	})
	_, err := p.transport.Send(jsonrpcRequest,
		&TransportSendOptions{
			RelatedRequestID:  safeOpts.RelatedRequestID,
			ResumptionToken:   safeOpts.ResumptionToken,
			OnResumptionToken: safeOpts.OnResumptionToken,
		})
	if err != nil {
		p.cleanupTimeout(messageID)
		go func() {
			returnChan <- requestReturnChan{
				e: err,
			}
		}()
		return
	}
	return
}

//Emits a notification, which is a one-way message that does not expect a response.
func (p *Protocol) Notification(notification types.NotificationInterface, opts *NotificationOptions) error {
	safeOpts := opts
	if safeOpts == nil {
		safeOpts = &NotificationOptions{}
	}
	if p.transport == nil {
		return fmt.Errorf("transport not connected")
	}
	err := p.owner.AssertNotificationCapability(notification)
	if err != nil {
		return fmt.Errorf("assertNotificationCapability error: %v", err)
	}
	jsonrpcNotification := &types.JSONRPCNotification{
		JSONRPC:               types.JSONRPC_VERSION,
		NotificationInterface: notification,
	}

	_, err = p.transport.Send(jsonrpcNotification, &TransportSendOptions{RelatedRequestID: safeOpts.RelatedRequestID})
	if err != nil {
		return fmt.Errorf("transport.Send error: %v", err)
	}
	return nil
}

//Registers a handler to invoke when this protocol object receives a request with the given method.
//
//Note that this will replace any previous request handler for the same method.
func (p *Protocol) SetRequestHandler(request types.RequestInterface, handler RequestHandler) {
	method := request.GetRequest().Method
	err := p.owner.AssertRequestHandlerCapability(request)
	if err != nil {
		detailErr := fmt.Errorf("assertRequestHandlerCapability %v", err)
		p.onError(detailErr)
	}

	if err := p.AssertCanSetRequestHandler(method); err != nil {
		p.onError(fmt.Errorf("p.AssertCanSetRequestHandler, %v", err))
	}

	p.requestHandlers.Set(method, handler)
}

//Removes the request handler for the given method.
func (p *Protocol) RemoveRequestHandler(method string) {
	p.requestHandlers.Delete(method)
}

//Asserts that a request handler has not already been set for the given method, in preparation for a new one being automatically installed.
func (p *Protocol) AssertCanSetRequestHandler(method string) error {
	_, ok := p.requestHandlers.Get(method)
	if ok {
		return fmt.Errorf("method %s has been registered", method)
	}
	return nil
}

//Registers a handler to invoke when this protocol object receives a notification with the given method.
//
//Note that this will replace any previous notification handler for the same method.
func (p *Protocol) SetNotificationHandler(notification types.NotificationInterface, handler NotificationHandler) {
	p.notificationHandlers.Set(notification.GetNotification().Method, handler)
}

//Removes the notification handler for the given method.
func (p *Protocol) RemoveNotificationHandler(method string) {
	p.notificationHandlers.Delete(method)
}
