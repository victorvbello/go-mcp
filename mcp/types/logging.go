package types

import (
	"encoding/json"
	"fmt"

	"github.com/victorvbello/gomcp/mcp/methods"
)

const (
	LOGGING_LEVEL_DEBUG     LoggingLevel = "debug"
	LOGGING_LEVEL_INFO      LoggingLevel = "info"
	LOGGING_LEVEL_NOTICE    LoggingLevel = "notice"
	LOGGING_LEVEL_WARNING   LoggingLevel = "warning"
	LOGGING_LEVEL_ERROR     LoggingLevel = "error"
	LOGGING_LEVEL_CRITICAL  LoggingLevel = "critical"
	LOGGING_LEVEL_ALERT     LoggingLevel = "alert"
	LOGGING_LEVEL_EMERGENCY LoggingLevel = "emergency"
)

//The severity of a log message.
//
//These map to syslog message severities, as specified in RFC-5424:
//https://datatracker.ietf.org/doc/html/rfc5424#section-6.2.1
type LoggingLevel string

//A request from the client to the server, to enable or adjust logging.
//
//Only method: METHOD_REQUEST_SET_LEVEL_LOGGING
type SetLevelRequest struct {
	Request
	Params SetLevelRequestParams `json:"params"`
}

func (sl *SetLevelRequest) TypeOfClientRequest() int { return SET_LEVEL_REQUEST_CLIENT_REQUEST_TYPE }

func (sl *SetLevelRequest) TypeOfRequestInterface() int {
	return SET_LEVEL_REQUEST_REQUEST_INTERFACE_TYPE
}
func (sl *SetLevelRequest) GetRequest() Request {
	return sl.Request
}

func NewSetLevelRequest(params *SetLevelRequestParams) *SetLevelRequest {
	slr := new(SetLevelRequest)
	slr.Method = methods.METHOD_REQUEST_SET_LEVEL_LOGGING
	if params != nil {
		slr.Params = *params
	}
	return slr
}

type SetLevelRequestParams struct {
	BaseRequestParams
	//The level of logging that the client wants to receive from the server. The server should send all logs at this level and higher (i.e., more severe) to the client as notifications/message.
	Level LoggingLevel `json:"level"`
}

func (slrp *SetLevelRequestParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		Level LoggingLevel `json:"level"`
	}{
		Level: slrp.Level,
	}
	knownFields, err := json.Marshal(&aux)
	if err != nil {
		return nil, fmt.Errorf("marshal known fields: %w", err)
	}
	//Marshal knownFields to map
	baseMap := make(map[string]interface{})
	if err := json.Unmarshal(knownFields, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal known fields to map: %w", err)
	}
	//Marshal base.BaseRequestParams
	baseExtra, err := slrp.BaseRequestParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (slrp *SetLevelRequestParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		Level LoggingLevel `json:"level"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	slrp.Level = aux.Level

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "level")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := slrp.BaseRequestParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("baseRequestParams.UnmarshalJSON: %w", err)
	}
	return nil
}

//Notification of a log message passed from server to client. If no logging/setLevel request has been sent from the client, the server MAY decide which messages to send automatically.
//
//Only method: METHOD_NOTIFICATION_MESSAGE
type LoggingMessageNotification struct {
	Notification
	Params LoggingMessageNotificationParams `json:"params"`
}

func (lmn *LoggingMessageNotification) TypeOfServerNotification() int {
	return LOGGING_MESSAGE_NOTIFICATION_SERVER_NOTIFICATION_TYPE
}
func (lmn *LoggingMessageNotification) TypeOfNotification() int {
	return LOGGING_MESSAGE_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
}
func (lmn *LoggingMessageNotification) GetNotification() Notification {
	return lmn.Notification
}

func NewLoggingMessageNotification(params *LoggingMessageNotificationParams) *LoggingMessageNotification {
	lmn := new(LoggingMessageNotification)
	lmn.Method = methods.METHOD_NOTIFICATION_MESSAGE
	if params != nil {
		lmn.Params = *params
	}
	return lmn
}

type LoggingMessageNotificationParams struct {
	BaseNotificationParams
	//The severity of this log message.
	Level LoggingLevel `json:"level"`
	//An optional name of the logger issuing this message.
	Logger string `json:"logger,omitempty"`
	//The data to be logged, such as a string message or an object. Any type is allowed here.
	Data interface{} `json:"data"`
}

func (lmnp *LoggingMessageNotificationParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		Level  LoggingLevel `json:"level"`
		Logger string       `json:"logger,omitempty"`
		Data   interface{}  `json:"data"`
	}{
		Level:  lmnp.Level,
		Logger: lmnp.Logger,
		Data:   lmnp.Data,
	}
	knownFields, err := json.Marshal(&aux)
	if err != nil {
		return nil, fmt.Errorf("marshal known fields: %w", err)
	}
	//Marshal knownFields to map
	baseMap := make(map[string]interface{})
	if err := json.Unmarshal(knownFields, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal known fields to map: %w", err)
	}
	//Marshal base.BaseNotificationParams
	baseExtra, err := lmnp.BaseNotificationParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (lmnp *LoggingMessageNotificationParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		Level  LoggingLevel `json:"level"`
		Logger string       `json:"logger,omitempty"`
		Data   interface{}  `json:"data"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	lmnp.Level = aux.Level
	lmnp.Logger = aux.Logger
	lmnp.Data = aux.Data

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "level")
	delete(raw, "logger")
	delete(raw, "data")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := lmnp.BaseNotificationParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("BaseNotificationParams.UnmarshalJSON: %w", err)
	}
	return nil
}
