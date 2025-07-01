package shared

import (
	"context"
	"sync"
	"time"

	"github.com/victorvbello/gomcp/mcp/types"
)

type notificationHandler func(notification types.NotificationInterface) error
type responseHandler func(response types.JSONRPCGeneralResponse) error

// func(request types.RequestInterface, extra *RequestHandlerExtra) (types.ResultInterface, error)
type requestHandler func(request types.RequestInterface, extra *RequestHandlerExtra) (types.ResultInterface, error)

type RequestOptions struct {
	TransportSendOptions
	//If set, requests progress notifications from the remote end (if supported). When progress notifications are received, this callback will be invoked.
	Onprogress types.ProgressCallback
	//This only can be a WithCancel context
	Context context.Context
	//This is required, if not set by default create new context
	ContextCancelFunc context.CancelFunc
	//A timeout (in milliseconds) for this request. If exceeded, an McpError with code `RequestTimeout` will be raised from request().
	//
	//If not specified, `DEFAULT_REQUEST_TIMEOUT_MSEC` will be used as the timeout.
	Timeout time.Duration
	//If true, receiving a progress notification will reset the request timeout.
	//This is useful for long-running operations that send periodic progress updates.
	//Default: false
	ResetTimeoutOnProgress *bool
	//Maximum total time (in milliseconds) to wait for a response.
	//If exceeded, an McpError with code `RequestTimeout` will be raised, regardless of progress notifications.
	//If not specified, there is no maximum total timeout.
	MaxTotalTimeout time.Duration
}

//Check if context was canceled
func (ro *RequestOptions) Canceled() bool {
	if ro.Context == nil {
		return true
	}
	return ro.Context.Err() == context.Canceled
}

type RequestHandlerExtra struct {
	//This only can be a WithCancel context
	Context context.Context
	//Information about a validated access token, provided to request handlers.
	AuthInfo types.AuthInfo
	//The session ID from the transport, if available.
	SessionID string
	//Metadata from the original request.
	Meta *types.MetadataRequest
	//The JSON-RPC ID of the request being handled.
	//This can be useful for tracking or logging purposes.
	RequestID types.RequestID
	//Sends a notification that relates to the current request being handled.
	//
	//This is used by certain transports to correctly associate related messages.
	SendNotification func(notification types.NotificationInterface)
	//Sends a request that relates to the current request being handled.
	//
	//This is used by certain transports to correctly associate related messages.
	SendRequest func(request types.RequestInterface, opts RequestOptions) (types.ResultInterface, error)
}

//Check if context was canceled
func (rh *RequestHandlerExtra) Canceled() bool {
	return rh.Context.Err() != nil
}

type ProtocolOptions struct {
	//Whether to restrict emitted requests to only those that the remote side has indicated that they can handle, through their advertised capabilities.
	//
	//Note that this DOES NOT affect checking of _local_ side capabilities, as it is considered a logic error to mis-specify those.
	//
	//Currently this defaults to false, for backwards compatibility with SDK versions that did not advertise capabilities correctly. In future, this will default to true.
	EnforceStrictCapabilities *bool
}

//Options that can be given per notification.
type NotificationOptions struct {
	//May be used to indicate to the transport which incoming request to associate this outgoing notification with.
	RelatedRequestID types.RequestID
}

//Information about a request's timeout state
type timeoutConfig struct {
	StartTime              time.Time
	Timeout                time.Duration
	MaxTotalTimeout        time.Duration
	ResetTimeoutOnProgress bool
	OnTimeout              func()
	ctx                    context.Context
	contextCancel          context.CancelFunc
}

//Start timeout using context with cancel
func (t *timeoutConfig) Start() {
	t.ctx, t.contextCancel = context.WithCancel(context.Background())
	go func() {
		select {
		case <-t.ctx.Done():
			return
		case <-time.After(t.Timeout):
			t.OnTimeout()
		}
	}()
}

//Call the cancel func of context
func (t *timeoutConfig) Clear() {
	t.contextCancel()
}

//muxRequestMessageID
type muxRequestMessageID struct {
	mu sync.RWMutex
	i  int
}

func (xi *muxRequestMessageID) Get() int {
	xi.mu.RLock()
	val := xi.i
	xi.mu.RUnlock()
	return val
}

func (xi *muxRequestMessageID) Increase() int {
	xi.mu.RLock()
	xi.i += 1
	val := xi.i
	xi.mu.RUnlock()
	return val
}

//muxMapRequestHandlerCancel
type muxMapRequestHandlerCancel struct {
	mu sync.RWMutex
	m  map[types.RequestID]context.CancelFunc
}

func (xm *muxMapRequestHandlerCancel) Clear() {
	xm.mu.Lock()
	xm.m = make(map[types.RequestID]context.CancelFunc)
	xm.mu.Unlock()
}

func (xm *muxMapRequestHandlerCancel) Get(key types.RequestID) (context.CancelFunc, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapRequestHandlerCancel) Set(key types.RequestID, value context.CancelFunc) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}

func (xm *muxMapRequestHandlerCancel) Delete(key types.RequestID) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}

//muxMapRequestHandlers
type muxMapRequestHandlers struct {
	mu sync.RWMutex
	m  map[string]requestHandler
}

func (xm *muxMapRequestHandlers) Clear() {
	xm.mu.Lock()
	xm.m = make(map[string]requestHandler)
	xm.mu.Unlock()
}

func (xm *muxMapRequestHandlers) Get(key string) (requestHandler, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapRequestHandlers) Set(key string, value requestHandler) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}

func (xm *muxMapRequestHandlers) Delete(key string) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}

//muxMapResponseHandlers
type muxMapResponseHandlers struct {
	mu sync.RWMutex
	m  map[int]responseHandler
}

func (xm *muxMapResponseHandlers) Clear() {
	xm.mu.Lock()
	xm.m = make(map[int]responseHandler)
	xm.mu.Unlock()
}

func (xm *muxMapResponseHandlers) Get(key int) (responseHandler, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapResponseHandlers) Set(key int, value responseHandler) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}

func (xm *muxMapResponseHandlers) GetAll() map[int]responseHandler {
	xm.mu.RLock()
	val := xm.m
	xm.mu.RUnlock()
	return val
}

func (xm *muxMapResponseHandlers) Delete(key int) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}

//muxMapNotificationHandlers
type muxMapNotificationHandlers struct {
	mu sync.RWMutex
	m  map[string]notificationHandler
}

func (xm *muxMapNotificationHandlers) Clear() {
	xm.mu.Lock()
	xm.m = make(map[string]notificationHandler)
	xm.mu.Unlock()
}

func (xm *muxMapNotificationHandlers) Get(key string) (notificationHandler, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapNotificationHandlers) Set(key string, value notificationHandler) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}

func (xm *muxMapNotificationHandlers) Delete(key string) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}

//muxMapProgressHandlers
type muxMapProgressHandlers struct {
	mu sync.RWMutex
	m  map[int]types.ProgressCallback
}

func (xm *muxMapProgressHandlers) Clear() {
	xm.mu.Lock()
	xm.m = make(map[int]types.ProgressCallback)
	xm.mu.Unlock()
}

func (xm *muxMapProgressHandlers) Get(key int) (types.ProgressCallback, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapProgressHandlers) Set(key int, value types.ProgressCallback) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}

func (xm *muxMapProgressHandlers) Delete(key int) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}

//muxMapTimeoutConfig
type muxMapTimeoutConfig struct {
	mu sync.RWMutex
	m  map[int]*timeoutConfig
}

func (xm *muxMapTimeoutConfig) Clear() {
	xm.mu.Lock()
	xm.m = make(map[int]*timeoutConfig)
	xm.mu.Unlock()
}

func (xm *muxMapTimeoutConfig) Get(key int) (*timeoutConfig, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapTimeoutConfig) Set(key int, value *timeoutConfig) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}
func (xm *muxMapTimeoutConfig) Delete(key int) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}
