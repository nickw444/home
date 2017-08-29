package common

import (
	"fmt"
	"os"
	"sync"
)

// Logger represents a minimal levelled logger
type Logger interface {
	// Debugf handles debug level messages
	Debugf(format string, args ...interface{})
	// Infof handles info level messages
	Infof(format string, args ...interface{})
	// Warnf handles warn level messages
	Warnf(format string, args ...interface{})
	// Errorf handles error level messages
	Errorf(format string, args ...interface{})
	// Fatalf handles fatal level messages, and must exit the application
	Fatalf(format string, args ...interface{})
	// Panicf handles debug level messages, and must panic the application
	Panicf(format string, args ...interface{})
}

// StubLogger satisfies the Logger interface, and simply does nothing with
// received messages
type StubLogger struct{}

// Debugf handles debug level messages
func (l *StubLogger) Debugf(format string, args ...interface{}) {}

// Infof handles info level messages
func (l *StubLogger) Infof(format string, args ...interface{}) {}

// Warnf handles warn level messages
func (l *StubLogger) Warnf(format string, args ...interface{}) {}

// Errorf handles error level messages
func (l *StubLogger) Errorf(format string, args ...interface{}) {}

// Fatalf handles fatal level messages, exits the application
func (l *StubLogger) Fatalf(format string, args ...interface{}) {
	os.Exit(1)
}

// Panicf handles debug level messages, and panics the application
func (l *StubLogger) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

type logPrefixer struct {
	log Logger
	sync.Mutex
}

// Debugf handles debug level messages, prefixing them for golifx
func (l *logPrefixer) Debugf(format string, args ...interface{}) {
	l.Lock()
	l.log.Debugf(l.prefix(format), args...)
	l.Unlock()
}

// Infof handles info level messages, prefixing them for golifx
func (l *logPrefixer) Infof(format string, args ...interface{}) {
	l.Lock()
	l.log.Infof(l.prefix(format), args...)
	l.Unlock()
}

// Warnf handles warn level messages, prefixing them for golifx
func (l *logPrefixer) Warnf(format string, args ...interface{}) {
	l.Lock()
	l.log.Warnf(l.prefix(format), args...)
	l.Unlock()
}

// Errorf handles error level messages, prefixing them for golifx
func (l *logPrefixer) Errorf(format string, args ...interface{}) {
	l.Lock()
	l.log.Errorf(l.prefix(format), args...)
	l.Unlock()
}

// Fatalf handles fatal level messages, prefixing them for golifx
func (l *logPrefixer) Fatalf(format string, args ...interface{}) {
	l.Lock()
	l.log.Fatalf(l.prefix(format), args...)
	l.Unlock()
}

// Panicf handles debug level messages, prefixing them for golifx
func (l *logPrefixer) Panicf(format string, args ...interface{}) {
	l.Lock()
	l.log.Panicf(l.prefix(format), args...)
	l.Unlock()
}

func (l *logPrefixer) prefix(format string) string {
	return `[golifx] ` + format
}

var (
	// Log holds the global logger used by golifx, can be set via SetLogger() in
	// the golifx package
	Log Logger
)

func init() {
	Log = &logPrefixer{log: new(StubLogger)}
}

// SetLogger wraps the supplied logger with a logPrefixer to denote golifx logs
func SetLogger(logger Logger) {
	Log = &logPrefixer{log: logger}
}
