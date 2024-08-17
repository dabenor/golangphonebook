// Standardize logging across the application
package internal

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Declare the Logger as a global variable so it can be used directly once the package is imported
var Logger *ConsoleLogger

func init() {
	Logger = NewConsoleLogger()
}

type ConsoleLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		debug: log.New(os.Stdout, "DEBUG: ", log.LstdFlags),
		info:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		warn:  log.New(os.Stdout, "WARN: ", log.LstdFlags),
		error: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}
}

func (l *ConsoleLogger) Debug(v ...interface{}) {
	l.debug.Println(v...)
}

func (l *ConsoleLogger) Info(v ...interface{}) {
	l.info.Println(v...)
}

func (l *ConsoleLogger) Warn(v ...interface{}) {
	l.warn.Println(v...)
}

func (l *ConsoleLogger) Error(v ...interface{}) {
	l.error.Println(v...)
}

func Timer(name string) func() {
	start := time.Now()
	return func() {
		Logger.Info(fmt.Sprintf("%s took %v\n", name, time.Since(start)))
	}
}
