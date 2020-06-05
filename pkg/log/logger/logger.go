package logger

import (
	"errors"
)

const (
	//Debug has verbose message
	Debug = "debug"
	//Info is default logger level
	Info = "info"
	//Warn is for logging messages about possible issues
	Warn = "warn"
	//Error is for logging errors
	Error = "error"
	//Fatal is for logging fatal messages. The system shutsdown after logging the message.
	Fatal = "fatal"
)

const (
	InstanceZapLogger int = iota
)

//Logger is our contract for the logger
type Logger interface {
	Debugf(format string, args ...interface{})

	Infof(format string, args ...interface{})

	Warnf(format string, args ...interface{})

	Errorf(format string, args ...interface{})

	Fatalf(format string, args ...interface{})

	Panicf(format string, args ...interface{})
}

// Configuration stores the config for the logger
type Configuration struct {
	EnableConsole     bool
	ConsoleJSONFormat bool
	ConsoleLevel      string
	EnableFile        bool
	FileJSONFormat    bool
	FileLevel         string
	FileLocation      string
}

//New returns an instance of logger
func New(config Configuration, loggerInstance int) (Logger, error) {
	switch loggerInstance {
	case InstanceZapLogger:
		logger, err := newZapLogger(config)
		if err != nil {
			return nil, err
		}

		return logger, nil

	default:
		return nil, errors.New("invalid logger instance")
	}
}
