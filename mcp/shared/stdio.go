package shared

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/victorvbello/gomcp/mcp/types"
)

type ReadBuffer struct {
	buffer bytes.Buffer
}

//Append adds new data to the buffer.
func (rb *ReadBuffer) Append(chunk []byte) (int, error) {
	b, err := rb.buffer.Write(chunk)
	if err != nil {
		return 0, fmt.Errorf("rb.buffer.Write %v", err)
	}
	return b, nil
}

//ReadMessage reads the next JSON-RPC message from the buffer if a full line is available.
func (rb *ReadBuffer) ReadMessage() (types.JSONRPCMessage, error) {
	//Use a buffered reader over the current buffer contents
	reader := bufio.NewReader(&rb.buffer)
	//Peek to see if we have a complete line (without consuming it)
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil //No full line yet
		}
		return nil, fmt.Errorf("reader.ReadString %v", err)
	}
	//Remove optional trailing \r
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		//Return if empty
		return nil, nil
	}

	//Consume the line from the buffer
	//Note: bufio.ReadString also advances the underlying reader, so the buffer is now clean
	var msg types.RawMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil, fmt.Errorf("json.Unmarshal %v", err)
	}
	finalMsg, err := msg.ToJSONRPCMessage()
	if err != nil {
		return nil, fmt.Errorf("msg.ToJSONRPCMessage %v", err)
	}
	return finalMsg, nil
}

//Clear resets the internal buffer.
func (rb *ReadBuffer) Clear() {
	rb.buffer.Reset()
}

func StdioSerializeMessage(msg types.JSONRPCMessage) (string, error) {
	data, err := types.JSONRPCMessageMarshalJSON(msg)
	if err != nil {
		return "", fmt.Errorf("json.Marshal %v", err)
	}
	return string(data) + "\n", nil
}

func StdioDeserializeMessage(line string) (*types.JSONRPCMessage, error) {
	var msg types.JSONRPCMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
