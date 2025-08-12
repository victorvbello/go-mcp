package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
)

type StreamableHTTPServerTransportOptions struct {
	//Function that generates a session ID for the transport.
	//The session ID SHOULD be globally unique and cryptographically secure (e.g., a securely generated UUID, a JWT, or a cryptographic hash)
	//
	//If is nil disable session management.
	SessionIDGenerator func() string
	//A callback for session initialization events
	//This is called when the server initializes a new session.
	//Useful in cases when you need to register multiple mcp sessions
	//and need to keep track of them.
	//sessionID(string) The generated session ID
	OnSessionInitialized func(sessionID string)
	//A callback for session close events
	//This is called when the server closes a session due to a DELETE request.
	//Useful in cases when you need to clean up resources associated with the session.
	//Note that this is different from the transport closing, if you are handling
	//HTTP requests from multiple nodes you might want to close each
	//StreamableHTTPServerTransport after a request is completed while still keeping the
	//session open/running.
	//sessionID(string) The generated session ID
	OnSessionClosed func(sessionID string)
	//If true, the server will return JSON responses instead of starting an SSE stream.
	//This can be useful for simple request/response scenarios without streaming.
	//Default is false (SSE streams are preferred).
	EnableJSONResponse *bool
	//Event store for resumability support
	//If provided, resumability will be enabled, allowing clients to reconnect and resume messages
	EventStore shared.EventStore
	//List of allowed host header values for DNS rebinding protection.
	//If not specified, host validation is disabled.
	AllowedHosts map[string]struct{}
	//List of allowed origin header values for DNS rebinding protection.
	//If not specified, origin validation is disabled.
	AllowedOrigins map[string]struct{}
	//Enable DNS rebinding protection (requires allowedHosts and/or allowedOrigins to be configured).
	//Default is false for backwards compatibility.
	EnableDNSRebindingProtection *bool
}

type ResponseWriter struct {
	writer http.ResponseWriter
}

func (r *ResponseWriter) Writer() http.ResponseWriter {
	return r.writer
}

func (r *ResponseWriter) SetWriter(w http.ResponseWriter) {
	r.writer = w
}

func (r *ResponseWriter) WriteJSON(code int, data interface{}) error {
	r.writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, err := json.Marshal(data)
	if err != nil {
		lErr := fmt.Errorf("json.Marshal payload: %v", err)
		r.writer.WriteHeader(http.StatusInternalServerError)
		r.writer.Write([]byte(lErr.Error()))
		return lErr
	}
	r.writer.WriteHeader(code)
	if code, err := r.writer.Write(b); err != nil {
		return fmt.Errorf("could not response - code: %d", code)
	}
	return nil
}

//muxMapStreamMapping
type muxMapStreamMapping struct {
	mu sync.RWMutex
	m  map[shared.StreamID]ResponseWriter
}

func (xm *muxMapStreamMapping) Clear() {
	xm.mu.Lock()
	xm.m = make(map[shared.StreamID]ResponseWriter)
	xm.mu.Unlock()
}

func (xm *muxMapStreamMapping) Get(key shared.StreamID) (ResponseWriter, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapStreamMapping) GetAll() map[shared.StreamID]ResponseWriter {
	xm.mu.RLock()
	clonedMap := make(map[shared.StreamID]ResponseWriter)
	for key, value := range xm.m {
		clonedMap[key] = value
	}
	xm.mu.RUnlock()
	return clonedMap
}

func (xm *muxMapStreamMapping) Set(key shared.StreamID, value ResponseWriter) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}
func (xm *muxMapStreamMapping) Delete(key shared.StreamID) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}

//muxMapRequestToStreamMapping
type muxMapRequestToStreamMapping struct {
	mu sync.RWMutex
	m  map[types.RequestID]shared.StreamID
}

func (xm *muxMapRequestToStreamMapping) Clear() {
	xm.mu.Lock()
	xm.m = make(map[types.RequestID]shared.StreamID)
	xm.mu.Unlock()
}

func (xm *muxMapRequestToStreamMapping) Get(key types.RequestID) (shared.StreamID, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapRequestToStreamMapping) GetAll() map[types.RequestID]shared.StreamID {
	xm.mu.RLock()
	clonedMap := make(map[types.RequestID]shared.StreamID)
	for key, value := range xm.m {
		clonedMap[key] = value
	}
	xm.mu.RUnlock()
	return clonedMap
}

func (xm *muxMapRequestToStreamMapping) Set(key types.RequestID, value shared.StreamID) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}
func (xm *muxMapRequestToStreamMapping) Delete(key types.RequestID) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}

//muxMapRequestResponseMap
type muxMapRequestResponseMap struct {
	mu sync.RWMutex
	m  map[types.RequestID]types.JSONRPCMessage
}

func (xm *muxMapRequestResponseMap) Clear() {
	xm.mu.Lock()
	xm.m = make(map[types.RequestID]types.JSONRPCMessage)
	xm.mu.Unlock()
}

func (xm *muxMapRequestResponseMap) Get(key types.RequestID) (types.JSONRPCMessage, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxMapRequestResponseMap) Set(key types.RequestID, value types.JSONRPCMessage) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}
func (xm *muxMapRequestResponseMap) Delete(key types.RequestID) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}
