package logger

import (
	"log"
	"os"
)

// Level represents log level
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

// Logger provides structured logging
type Logger struct {
	level  Level
	debug  *log.Logger
	info   *log.Logger
	warn   *log.Logger
	errLog *log.Logger
}

// New creates a new logger with the specified level
func New(levelStr string) *Logger {
	level := parseLevel(levelStr)

	return &Logger{
		level:  level,
		debug:  log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lshortfile),
		info:   log.New(os.Stdout, "[INFO]  ", log.LstdFlags),
		warn:   log.New(os.Stdout, "[WARN]  ", log.LstdFlags),
		errLog: log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile),
	}
}

// Debug logs debug messages
func (l *Logger) Debug(v ...interface{}) {
	if l.level <= DEBUG {
		l.debug.Println(v...)
	}
}

// Debugf logs formatted debug messages
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.debug.Printf(format, v...)
	}
}

// Info logs info messages
func (l *Logger) Info(v ...interface{}) {
	if l.level <= INFO {
		l.info.Println(v...)
	}
}

// Infof logs formatted info messages
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= INFO {
		l.info.Printf(format, v...)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(v ...interface{}) {
	if l.level <= WARN {
		l.warn.Println(v...)
	}
}

// Warnf logs formatted warning messages
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level <= WARN {
		l.warn.Printf(format, v...)
	}
}

// Error logs error messages
func (l *Logger) Error(v ...interface{}) {
	if l.level <= ERROR {
		l.errLog.Println(v...)
	}
}

// Errorf logs formatted error messages
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.errLog.Printf(format, v...)
	}
}

// Fatal logs error and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.errLog.Fatal(v...)
}

// Fatalf logs formatted error and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.errLog.Fatalf(format, v...)
}

func parseLevel(levelStr string) Level {
	switch levelStr {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}
