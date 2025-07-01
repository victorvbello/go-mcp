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

func (lg *Logger) Info(fields LogFields, msg string) {
	lg.Log.Printf("[INFO]: %s - %s\n", msg, lg.fieldsToString())
}

func (lg *Logger) Error(fields LogFields, msg string) {
	lg.Log.Printf("[ERROR]: %s - %s\n", msg, lg.fieldsToString())
}

func (lg *Logger) Warning(fields LogFields, msg string) {
	lg.Log.Printf("[WARNING]: %s - %s\n", msg, lg.fieldsToString())
}

func (lg *Logger) Fatal(fields LogFields, msg string) {
	lg.Log.Fatalf("[FATAL]: %s - %s\n", msg, lg.fieldsToString())
}

func (lg *Logger) AddFields(fields LogFields) {
	for k, v := range fields {
		lg.Fields[k] = v
	}
}

func (lg *Logger) RemoveField(key string) {
	delete(lg.Fields, key)
}
