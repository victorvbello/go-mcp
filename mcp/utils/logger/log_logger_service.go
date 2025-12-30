package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger struct {
	Log    *log.Logger
	Fields map[string]interface{}
}

func NewLoggerService(name string) LogService {
	lg := &Logger{
		Log:    log.New(os.Stderr, "LoggerService["+name+"]: ", log.Ldate|log.Ltime),
		Fields: make(map[string]interface{}),
	}
	return lg
}

func (lg *Logger) fieldsToString(inlineFields LogFields) string {
	var result string
	finalFields := make(LogFields)
	for k, v := range lg.Fields {
		finalFields[k] = v
	}
	for k, v := range inlineFields {
		finalFields[k] = v
	}
	for k, v := range finalFields {
		result += fmt.Sprintf("%s: %v ", k, v)
	}
	return result
}

func (lg *Logger) formatLog(flag string, fields LogFields, msg string) string {
	data := lg.fieldsToString(fields)
	step := ""
	if data != "" {
		step = "-"
	}
	for s, r := range loggerKeyWordsSafe {
		msg = strings.ReplaceAll(msg, s, r)
	}
	return fmt.Sprintf("%s: %s %s %s\n", flag, msg, step, data)
}

func (lg *Logger) Info(fields LogFields, msg string) {
	lg.Log.Printf(lg.formatLog(LOG_LEVEL_INFO, fields, msg))
}

func (lg *Logger) Error(fields LogFields, msg string) {
	lg.Log.Printf(lg.formatLog(LOG_LEVEL_ERROR, fields, msg))
}

func (lg *Logger) Warning(fields LogFields, msg string) {
	lg.Log.Printf(lg.formatLog(LOG_LEVEL_WARNING, fields, msg))
}

func (lg *Logger) Fatal(fields LogFields, msg string) {
	lg.Log.Fatalf(lg.formatLog(LOG_LEVEL_FATAL, fields, msg))
}

func (lg *Logger) AddFields(fields LogFields) {
	for k, v := range fields {
		lg.Fields[k] = v
	}
}

func (lg *Logger) RemoveField(key string) {
	delete(lg.Fields, key)
}

func (lg *Logger) LogMsgIsLevel(logMsg string, level string) bool {
	return strings.Contains(logMsg, level)
}
