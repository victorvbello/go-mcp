package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	utils "github.com/victorvbello/gomcp/mcp/utils/logger"
)

const (
	_MAX_CLIENT_STDIO_BUFFER_READ = 4096
)

var _DEFAULT_INHERITED_ENV_VARS = []string{"HOME", "LOGNAME", "PATH", "SHELL", "TERM", "USER"}

type StdioServerParameters struct {
	//The executable to run to start the server.
	Command string
	//Command line arguments to pass to the executable.
	Args []string
	// The environment to use when spawning the process.
	//
	// If not specified, the result of getDefaultEnvironment() will be used.
	Env map[string]string
	//The working directory to use when spawning the process.
	//
	//If not specified, the current working directory will be inherited.
	CWD string
}

// Returns a default environment object including only environment variables deemed safe to inherit.
func GetDefaultEnvironment() map[string]string {
	env := make(map[string]string)

	for _, key := range _DEFAULT_INHERITED_ENV_VARS {
		currentEnv := os.Getenv(key)
		if currentEnv == "" {
			continue
		}
		env[key] = currentEnv
	}
	return env
}

// Client transport for stdio: this will connect to a server by spawning a process and communicating with it over stdin/stdout.
//
// This transport is only available in Node.js environments.
type StdioClientTransport struct {
	mu                  sync.RWMutex
	serverSTDin         *bufio.Writer
	serverSTDout        *bufio.Reader
	serverSTDerr        *bufio.Reader
	serverCommand       *exec.Cmd
	protocolVersion     string
	globalOnClose       func()
	globalOnError       func(err error)
	globalOnMessage     func(message types.JSONRPCMessage, extra *shared.MessageExtraInfo)
	globalContext       context.Context
	globalContextCancel context.CancelFunc
	readBuffer          shared.ReadBuffer
	readBufferNotify    shared.ReadBuffer
	serverParams        StdioServerParameters
	started             bool
	logger              utils.LogService
	showServerInfoLog   bool
}

func NewStdioClientTransport(sp StdioServerParameters, showServerInfoLog bool) shared.Transport {
	nct := &StdioClientTransport{
		serverParams: sp,
		logger:       utils.NewLoggerService("new-stdio-client-transport"),
	}
	return nct
}

func (st *StdioClientTransport) processReadBuffer() {
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

func (st *StdioClientTransport) processReadBufferNotifcation() {
	for {
		messageStr, err := st.readBufferNotify.ReadStringData()
		if err != nil {
			st.OnError(fmt.Errorf("st.readBuffer.ReadStringData %v", err))
		}
		if messageStr == "" {
			break
		}
		if json.Valid([]byte(messageStr)) {
			message, err := st.readBufferNotify.StringDataToMessage(messageStr)
			if err != nil {
				st.OnError(fmt.Errorf("st.readBuffer.StringDataToMessage msg:%s err:%v", messageStr, err))
				break
			}
			st.OnMessage(message, nil)
			break
		}
		if !st.logger.LogMsgIsLevel(messageStr, utils.LOG_LEVEL_ERROR) && st.showServerInfoLog {
			st.logger.Info(nil, messageStr)
			break
		}
		st.onError(fmt.Errorf(messageStr))
	}
}

func (st *StdioClientTransport) onData(chunk []byte) {
	st.readBuffer.Append(chunk)
	st.processReadBuffer()
}

func (st *StdioClientTransport) onNotificationData(chunk []byte) {
	st.readBufferNotify.Append(chunk)
	st.processReadBufferNotifcation()
}

func (st *StdioClientTransport) onError(err error) {
	st.OnError(err)
}

// Starts the server process and prepares to communicate with it.
func (st *StdioClientTransport) Start(ctx context.Context) error {
	if st.started {
		return fmt.Errorf("stdioClientTransport already started! If using Client class, note that connect() calls start() automatically")
	}

	st.globalContext, st.globalContextCancel = context.WithCancel(ctx)

	env := GetDefaultEnvironment()
	if len(st.serverParams.Env) > 0 {
		env = st.serverParams.Env
	}
	var finalEnv []string
	for k, v := range env {
		finalEnv = append(finalEnv, k+"="+v)
	}

	st.serverCommand = exec.CommandContext(
		st.globalContext,
		st.serverParams.Command,
		st.serverParams.Args...)
	st.serverCommand.Env = append(os.Environ(), finalEnv...)
	st.serverCommand.Dir = st.serverParams.CWD

	serverStdin, err := st.serverCommand.StdinPipe()
	if err != nil {
		return fmt.Errorf("cmd.StdoutPipe %v", err)
	}
	defer func() {
		if err == nil {
			return
		}
		serverStdin.Close()
	}()
	st.serverSTDin = bufio.NewWriter(serverStdin)

	serverStdout, err := st.serverCommand.StdoutPipe()
	if err != nil {
		return fmt.Errorf("cmd.StdoutPipe %v", err)
	}
	defer func() {
		if err == nil {
			return
		}
		serverStdout.Close()
	}()
	st.serverSTDout = bufio.NewReader(serverStdout)

	serverStderr, err := st.serverCommand.StderrPipe()
	if err != nil {
		return fmt.Errorf("cmd.StderrPipe %v", err)
	}
	defer func() {
		if err == nil {
			return
		}
		serverStderr.Close()
	}()
	st.serverSTDerr = bufio.NewReader(serverStderr)

	if err := st.serverCommand.Start(); err != nil {
		return fmt.Errorf("cmd.Start %v", err)
	}

	go st.handlerSTDIOReader("StdoutPipe", st.serverSTDout, func(data []byte) error {
		st.onData(data)
		return nil
	})
	go st.handlerSTDIOReader("StderrPipe", st.serverSTDerr, func(data []byte) error {
		st.onNotificationData(data)
		return nil
	})

	go func() {
		if err := st.serverCommand.Wait(); err != nil {
			cmdError := fmt.Errorf("cmd.Wait %v", err)
			st.onError(cmdError)
		}
	}()
	st.started = true
	return nil
}

func (st *StdioClientTransport) handlerSTDIOReader(readerType string, reader *bufio.Reader, porcessData func([]byte) error) {
	buf := make([]byte, _MAX_CLIENT_STDIO_BUFFER_READ)
	for {
		select {
		case <-st.globalContext.Done():
			st.logger.Info(nil, fmt.Sprintf("%s gracefully stop reading", readerType))
			return //gracefully stop reading
		default:
			n, err := reader.Read(buf)
			if st.serverCommand == nil {
				st.logger.Info(nil, fmt.Sprintf("%s gracefully stop serverCommand", readerType))
				return
			}
			if st.serverCommand.ProcessState != nil && st.serverCommand.ProcessState.Exited() {
				st.logger.Info(nil, fmt.Sprintf("%s gracefully stop serverCommand %v", readerType, st.serverCommand.ProcessState))
				return
			}
			if err != nil {
				st.onError(fmt.Errorf("%s.Read in start %v", readerType, err))
				return
			}
			data := make([]byte, n)
			copy(data, buf[:n])
			go func() {
				err := porcessData(data)
				if err != nil {
					st.onError(fmt.Errorf("porcessData in handlerSTDIOReader for %s %v", readerType, err))
					return
				}
			}()
		}
	}
}

// Sends a JSON-RPC message (request or response).
//
// If present, `relatedRequestId` is used to indicate to the transport which incoming request to associate this outgoing message with.
func (st *StdioClientTransport) Send(request types.JSONRPCMessage, options *shared.TransportSendOptions) (*types.JSONRPCResponse, error) {
	if st.serverSTDin == nil {
		return nil, fmt.Errorf("server not connected")
	}
	msgJSON, err := shared.StdioSerializeMessage(request)
	if err != nil {
		return nil, fmt.Errorf("shared.StdioSerializeMessage %v", err)
	}
	st.mu.RLock()
	_, err = st.serverSTDin.Write([]byte(msgJSON))
	if err != nil {
		return nil, fmt.Errorf("st.serverSTDin.Write in Send %v", err)
	}
	err = st.serverSTDin.Flush()
	if err != nil {
		return nil, fmt.Errorf("serverStdoutBuff.Flush in Send %v", err)
	}
	st.mu.RUnlock()
	return nil, nil
}

// Closes the connection.
func (st *StdioClientTransport) Close() error {
	st.serverSTDin = nil
	st.serverSTDout = nil
	st.serverSTDerr = nil
	st.serverCommand = nil
	//Clear the buffers and notify closure
	st.readBuffer.Clear()
	st.readBufferNotify.Clear()
	st.globalContextCancel()
	err := st.OnClose()
	if err != nil {
		lErr := fmt.Errorf("OnClose Error %v", err)
		st.OnError(lErr)
		return lErr
	}
	st.OnError(nil)
	time.Sleep(1 * time.Second)
	return nil
}

// Callback for when the connection is closed for any reason.
//
// This should be invoked when close() is called as well.
//
// Always execute first the prop globalOnClose if is defined
func (st *StdioClientTransport) OnClose() error {
	if st.globalOnClose != nil {
		st.globalOnClose()
	}
	return nil
}

// Callback for when an error occurs.
//
// Note that errors are not necessarily fatal; they are used for reporting any kind of exceptional condition out of band.
//
// Always execute first the prop globalOnError if is defined
func (st *StdioClientTransport) OnError(err error) {
	if st.globalOnError != nil {
		st.globalOnError(err)
	}
}

// Callback for when a message (request or response) is received over the connection.
//
// Includes the authInfo if the transport is authenticated.
//
// Always execute first the prop globalOnMessage if is defined
func (st *StdioClientTransport) OnMessage(message types.JSONRPCMessage, extra *shared.MessageExtraInfo) {
	if st.globalOnMessage != nil {
		st.globalOnMessage(message, extra)
	}
}

// Sets the protocol version used for the connection (called when the initialize response is received).
func (st *StdioClientTransport) SetProtocolVersion(version string) {
	st.protocolVersion = version
}

// Return the session ID
func (st *StdioClientTransport) GetSessionID() string {
	return "mcp-session-id-stdio"
}

// Set this if globalOnClose is needed, this must be executed into OnClose Func first
func (st *StdioClientTransport) SetGlobalOnClose(fn func()) {
	st.globalOnClose = fn
}

// Set this if globalOnError is needed, this must be executed into OnError Func first
func (st *StdioClientTransport) SetGlobalOnError(fn func(err error)) {
	st.globalOnError = fn
}

// Set this if globalOnMessage is needed, this must be executed into OnMessage Func first
func (st *StdioClientTransport) SetGlobalOnMessage(fn func(message types.JSONRPCMessage, extra *shared.MessageExtraInfo)) {
	st.globalOnMessage = fn
}
