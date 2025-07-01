package utils

type LogFields map[string]interface{}
type LogService interface {
	AddFields(fields LogFields)
	Info(fields LogFields, msg string)
	Warning(fields LogFields, msg string)
	Error(fields LogFields, msg string)
	Fatal(fields LogFields, msg string)
	RemoveField(msg string)
}
