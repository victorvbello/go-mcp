package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	utils "github.com/victorvbello/gomcp/mcp/utils/logger"
)

const (
	_MAX_STDIO_BUFFER_READ = 4096
)

//Server transport for stdio: this communicates with a MCP client by reading from the current process' stdin and writing to stdout.
//
//This transport is only available in Node.js environments.
type StdioServerTransport struct {
	mu                  sync.RWMutex
	protocolVersion     string
	globalOnClose       func()
	globalOnError       func(err error)
	globalOnMessage     func(message types.JSONRPCMessage, extra *shared.MessageExtraInfo)
	globalContext       context.Context
	globalContextCancel context.CancelFunc
	stdin               *bufio.Reader
	stdout              *bufio.Writer
	started             bool
	readBuffer          shared.ReadBuffer
	logger              utils.LogService
}

func NewStdioServerTransport(stdin io.Reader, stdout io.Writer) shared.Transport {
	nst := &StdioServerTransport{
		stdin:  bufio.NewReader(stdin),
		stdout: bufio.NewWriter(stdout),
		logger: utils.NewLoggerService(),
	}
	return nst
}

func (st *StdioServerTransport) processReadBuffer() {
	for {
		message, err := st.readBuffer.ReadMessage()
		if err != nil {
			st.OnError(fmt.Errorf("st.readBuffer.ReadMessage %v", err))
		}
		if message == nil {
			break
		}
		st.OnMessage(message, nil)
	}
}

func (st *StdioServerTransport) onData(chunk []byte) {
	st.readBuffer.Append(chunk)
	st.processReadBuffer()
}

func (st *StdioServerTransport) onError(err error) {
	st.OnError(err)
}

//Starts processing messages on the transport, including any connection steps that might need to be taken.
//
//This method should only be called after callbacks are installed, or else messages may be lost.
//
//NOTE: This method should not be called explicitly when using Client, Server, or Protocol classes, as they will implicitly call start().
func (st *StdioServerTransport) Start() error {
	if st.started {
		return fmt.Errorf("stdioServerTransport already started! If using Server class, note that connect() calls start() automatically")
	}

	st.globalContext, st.globalContextCancel = context.WithCancel(context.Background())

	go func() {
		buf := make([]byte, _MAX_STDIO_BUFFER_READ)
		for {
			select {
			case <-st.globalContext.Done():
				st.logger.Info(nil, "gracefully stop reading")
				return //gracefully stop reading
			default:
				n, err := st.stdin.Read(buf)
				if err != nil {
					st.onError(fmt.Errorf("st.stdin.Read %v", err))
					return
				}
				go func() {
					st.onData(buf[:n])
				}()
			}
		}
	}()

	st.started = true
	return nil
}

//Sends a JSON-RPC message (request or response).
//
//If present, `relatedRequestId` is used to indicate to the transport which incoming request to associate this outgoing message with.
func (st *StdioServerTransport) Send(request types.JSONRPCMessage, options *shared.TransportSendOptions) (*types.JSONRPCResponse, error) {
	msgJSON, err := shared.StdioSerializeMessage(request)
	if err != nil {
		return nil, fmt.Errorf("shared.StdioSerializeMessage %v", err)
	}
	st.mu.RLock()
	stdout := st.stdout
	st.mu.RUnlock()
	_, err = stdout.Write([]byte(msgJSON))
	if err != nil {
		return nil, fmt.Errorf("st.stdout.Write %v", err)
	}
	err = stdout.Flush()
	if err != nil {
		return nil, fmt.Errorf("st.stdout.Flush %v", err)
	}
	return nil, nil
}

//Closes the connection.
func (st *StdioServerTransport) Close() error {
	st.globalContextCancel()
	//Clear the buffer and notify closure
	st.readBuffer.Clear()
	err := st.OnClose()
	if err != nil {
		lErr := fmt.Errorf("OnClose Error %v", err)
		st.OnError(lErr)
		return lErr
	}
	return nil
}

//Callback for when the connection is closed for any reason.
//
//This should be invoked when close() is called as well.
//
//Always execute first the prop globalOnClose if is defined
func (st *StdioServerTransport) OnClose() error {
	if st.globalOnClose != nil {
		st.globalOnClose()
	}
	return nil
}

//Callback for when an error occurs.
//
//Note that errors are not necessarily fatal; they are used for reporting any kind of exceptional condition out of band.
//
//Always execute first the prop globalOnError if is defined
func (st *StdioServerTransport) OnError(err error) {
	if st.globalOnError != nil {
		st.globalOnError(err)
	}
}

//Callback for when a message (request or response) is received over the connection.
//
//Includes the authInfo if the transport is authenticated.
//
//Always execute first the prop globalOnMessage if is defined
func (st *StdioServerTransport) OnMessage(message types.JSONRPCMessage, extra *shared.MessageExtraInfo) {
	if st.globalOnMessage != nil {
		st.globalOnMessage(message, extra)
	}
}

//Sets the protocol version used for the connection (called when the initialize response is received).
func (st *StdioServerTransport) SetProtocolVersion(version string) {
	st.protocolVersion = version
}

//Return the session ID
func (st *StdioServerTransport) GetSessionID() string {
	return "mcp-session-id-stdio"
}

//Set this if globalOnClose is needed, this must be executed into OnClose Func first
func (st *StdioServerTransport) SetGlobalOnClose(fn func()) {
	st.globalOnClose = fn
}

//Set this if globalOnError is needed, this must be executed into OnError Func first
func (st *StdioServerTransport) SetGlobalOnError(fn func(err error)) {
	st.globalOnError = fn
}

//Set this if globalOnMessage is needed, this must be executed into OnMessage Func first
func (st *StdioServerTransport) SetGlobalOnMessage(fn func(message types.JSONRPCMessage, extra *shared.MessageExtraInfo)) {
	st.globalOnMessage = fn
}
