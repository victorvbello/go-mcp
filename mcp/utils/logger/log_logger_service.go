package utils

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	Log    *log.Logger
	Fields map[string]interface{}
}

func NewLoggerService() LogService {
	lg := &Logger{
		Log:    log.New(os.Stdout, "LoggerService: ", log.Ldate|log.Ltime),
		Fields: make(map[string]interface{}),
	}
	return lg
}

func (lg *Logger) fieldsToString() string {
	var result string
	for k, v := range lg.Fields {
		result += fmt.Sprintf("%s: %v ", k, v)
	}
	return result
}

func (lg *Logger) formatLog(flag string, fields LogFields, msg string) string {
	data := lg.fieldsToString()
	step := ""
	if data != "" {
		step = "-"
	}
	return fmt.Sprintf("%s: %s %s %s\n", flag, msg, step, data)
}

func (lg *Logger) Info(fields LogFields, msg string) {
	lg.Log.Printf(lg.formatLog("[INFO]", fields, msg))
}

func (lg *Logger) Error(fields LogFields, msg string) {
	lg.Log.Printf(lg.formatLog("[ERROR]", fields, msg))
}

func (lg *Logger) Warning(fields LogFields, msg string) {
	lg.Log.Printf(lg.formatLog("[WARNING]", fields, msg))
}

func (lg *Logger) Fatal(fields LogFields, msg string) {
	lg.Log.Fatalf(lg.formatLog("[FATAL]", fields, msg))
}

func (lg *Logger) AddFields(fields LogFields) {
	for k, v := range fields {
		lg.Fields[k] = v
	}
}

func (lg *Logger) RemoveField(key string) {
	delete(lg.Fields, key)
}
