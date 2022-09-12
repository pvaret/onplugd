package messagepipe

import (
	"fmt"
	"log"
	"os"

	"onplugd/utils"
)

// Severity is the severity of a message.
type Severity uint8

const (
	// SeverityError is the severity for actual errors.
	SeverityError Severity = 1 << iota
	// SeverityInfo is the severity for informational messages.
	SeverityInfo
	// SeverityDebug is the severity for messages that are not normally
	// useful except when debugging things.
	SeverityDebug
)

// MessagePipe is an implementation of IMessagePipe.
type MessagePipe struct {
	handlers []func(Severity, string)
	logger   *log.Logger
}

// Error records an error message in the pipe.
func (m *MessagePipe) Error(e error) {
	m.feed(SeverityError, fmt.Sprint(e))
}

// Info records an informational message in the pipe.
func (m *MessagePipe) Info(msg string) {
	m.feed(SeverityInfo, msg)
}

// Debug records an informational message in the pipe.
func (m *MessagePipe) Debug(msg string) {
	m.feed(SeverityDebug, msg)
}

// AddHandler adds a handler on this MessagePipe.
func (m *MessagePipe) AddHandler(f func(Severity, string)) {
	m.handlers = append(m.handlers, f)
}

// ClearHandlers removes all the message handlers from this MessagePipe.
func (m *MessagePipe) ClearHandlers() {
	m.handlers = nil
}

func (m *MessagePipe) feed(severity Severity, message string) {
	for _, handler := range m.handlers {
		handler(severity, message)
	}
}

// New creates a new MessagePipe and populates a default handler.
func New(debug bool) MessagePipe {
	m := MessagePipe{}

	var flags int
	if utils.IsATerminal(os.Stdout) {
		// If running in a terminal, add the full date to log lines.
		flags = log.LstdFlags
	}

	m.logger = log.New(os.Stdout, "", flags)

	m.AddHandler(func(s Severity, msg string) {
		if s == SeverityError {
			m.logger.Println("ERROR:", msg)
		} else if s == SeverityInfo {
			m.logger.Println("INFO:", msg)
		} else if s == SeverityDebug && debug {
			m.logger.Println("DEBUG:", msg)
		}
	})

	return m
}
