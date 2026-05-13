package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/op/go-logging"
)

const (
	maxLogBufferSize = 10240
	maxLogLineLength = 4096
)

var (
	logger      *logging.Logger
	loggerLevel logging.Level
	logBufferMu sync.RWMutex
	logBuffer   []struct {
		time  string
		level logging.Level
		log   string
	}
)

func InitLogger(level logging.Level) {
	newLogger := logging.MustGetLogger("s-ui")
	var err error
	var backend logging.Backend
	var format logging.Formatter

	_, inContainer := os.LookupEnv("container")
	if !inContainer {
		if _, statErr := os.Stat("/.dockerenv"); statErr == nil {
			inContainer = true
		}
	}
	if inContainer {
		backend = logging.NewLogBackend(os.Stderr, "", 0)
		format = logging.MustStringFormatter(`%{time:2006/01/02 15:04:05} %{level} - %{message}`)
	} else {
		backend, err = logging.NewSyslogBackend("")
		if err != nil {
			fmt.Println("Unable to use syslog: " + err.Error())
			backend = logging.NewLogBackend(os.Stderr, "", 0)
		}
		if err != nil {
			format = logging.MustStringFormatter(`%{time:2006/01/02 15:04:05} %{level} - %{message}`)
		} else {
			format = logging.MustStringFormatter(`%{level} - %{message}`)
		}
	}

	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(level, "s-ui")
	newLogger.SetBackend(backendLeveled)

	logger = newLogger
	loggerLevel = level
}

func GetLogger() *logging.Logger {
	return logger
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
	addToBuffer("DEBUG", fmt.Sprint(args...))
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
	addToBuffer("DEBUG", fmt.Sprintf(format, args...))
}

func Info(args ...interface{}) {
	logger.Info(args...)
	addToBuffer("INFO", fmt.Sprint(args...))
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
	addToBuffer("INFO", fmt.Sprintf(format, args...))
}

func Warning(args ...interface{}) {
	logger.Warning(args...)
	addToBuffer("WARNING", fmt.Sprint(args...))
}

func Warningf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
	addToBuffer("WARNING", fmt.Sprintf(format, args...))
}

func Error(args ...interface{}) {
	logger.Error(args...)
	addToBuffer("ERROR", fmt.Sprint(args...))
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
	addToBuffer("ERROR", fmt.Sprintf(format, args...))
}

func addToBuffer(level string, newLog string) {
	if len(newLog) > maxLogLineLength {
		newLog = newLog[:maxLogLineLength] + "...(truncated)"
	}
	t := time.Now()
	logLevel, _ := logging.LogLevel(level)
	if logLevel > loggerLevel {
		return
	}

	logBufferMu.Lock()
	defer logBufferMu.Unlock()

	if len(logBuffer) >= maxLogBufferSize {
		logBuffer = logBuffer[1:]
	}

	logBuffer = append(logBuffer, struct {
		time  string
		level logging.Level
		log   string
	}{
		time:  t.Format("2006/01/02 15:04:05"),
		level: logLevel,
		log:   newLog,
	})
}

func GetLogs(c int, level string) []string {
	if c <= 0 {
		return nil
	}
	var output []string
	logLevel, _ := logging.LogLevel(level)

	logBufferMu.RLock()
	defer logBufferMu.RUnlock()

	for i := len(logBuffer) - 1; i >= 0 && len(output) < c; i-- {
		entry := logBuffer[i]
		if entry.level <= logLevel {
			output = append(output, fmt.Sprintf("%s %s - %s", entry.time, entry.level, entry.log))
		}
	}
	return output
}
