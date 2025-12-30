package utils

const (
	LOG_LEVEL_INFO    = "[INFO]"
	LOG_LEVEL_ERROR   = "[ERROR]"
	LOG_LEVEL_WARNING = "[WARNING]"
	LOG_LEVEL_FATAL   = "[FATAL]"
)

var loggerKeyWordsSafe = map[string]string{
	LOG_LEVEL_INFO:    "info",
	LOG_LEVEL_ERROR:   "error",
	LOG_LEVEL_WARNING: "warning",
	LOG_LEVEL_FATAL:   "fatal",
}

type LogFields map[string]interface{}
type LogService interface {
	AddFields(fields LogFields)
	Info(fields LogFields, msg string)
	Warning(fields LogFields, msg string)
	Error(fields LogFields, msg string)
	Fatal(fields LogFields, msg string)
	RemoveField(msg string)
	LogMsgIsLevel(logMsg string, level string) bool
}
