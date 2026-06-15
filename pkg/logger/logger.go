package logger

import (
	"fmt"
	"log"
	"os"
)

// Define our log levels
const (
	LevelDebug = iota
	LevelInfo
	LevelError
)

// Logger is our custom logging struct
type Logger struct {
	Level    int
	infoLog  *log.Logger
	debugLog *log.Logger
	errorLog *log.Logger
}

// Global Buffer
var LiveBuffer = NewBuffer(200)

// New creates a new TyPO logger instance
func New(level int) *Logger {
	return &Logger{
		Level:    level,
		infoLog:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		debugLog: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info prints standard operational messages
func (l *Logger) Info(format string, v ...interface{}) {
	if l.Level <= LevelInfo {
		LiveBuffer.Add(fmt.Sprintf(format, v...))
		l.infoLog.Output(2, fmt.Sprintf(format, v...))
	}
}

// Debug prints granular, white-box tracing details
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.Level <= LevelDebug {
		LiveBuffer.Add(fmt.Sprintf(format, v...))
		l.debugLog.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error prints failure states to Standard Error
func (l *Logger) Error(format string, v ...interface{}) {
	if l.Level <= LevelError {
		LiveBuffer.Add(fmt.Sprintf(format, v...))
		l.errorLog.Output(2, fmt.Sprintf(format, v...))
	}
}
