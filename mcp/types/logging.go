package types

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

type SetLevelRequestParams struct {
	//The level of logging that the client wants to receive from the server. The server should send all logs at this level and higher (i.e., more severe) to the client as notifications/message.
	Level LoggingLevel `json:"level"`
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

type LoggingMessageNotificationParams struct {
	//The severity of this log message.
	Level LoggingLevel `json:"level"`
	//An optional name of the logger issuing this message.
	Logger string `json:"logger,omitempty"`
	//The data to be logged, such as a string message or an object. Any type is allowed here.
	Data interface{} `json:"data"`
}
