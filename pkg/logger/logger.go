package logger

import (
	"log"
	"os"
	"strings"
)

// Level represents logging level
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger provides structured logging
type Logger struct {
	level       Level
	serviceName string
	logger      *log.Logger
}

// New creates a new Logger instance
func New(serviceName, level string) *Logger {
	return &Logger{
		level:       parseLevel(level),
		serviceName: serviceName,
		logger:      log.New(os.Stdout, "", log.LstdFlags),
	}
}

// parseLevel converts string level to Level
func parseLevel(level string) Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level <= DEBUG {
		l.log("DEBUG", msg, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level <= INFO {
		l.log("INFO", msg, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.level <= WARN {
		l.log("WARN", msg, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	if l.level <= ERROR {
		l.log("ERROR", msg, args...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log("FATAL", msg, args...)
	os.Exit(1)
}

// log is the internal logging method
func (l *Logger) log(level, msg string, args ...interface{}) {
	prefix := "[" + l.serviceName + "] [" + level + "] "
	if len(args) > 0 {
		l.logger.Printf(prefix+msg, args...)
	} else {
		l.logger.Println(prefix + msg)
	}
}

// With creates a new logger with additional context
func (l *Logger) With(key, value string) *Logger {
	newLogger := *l
	newLogger.serviceName = l.serviceName + " " + key + "=" + value
	return &newLogger
}
