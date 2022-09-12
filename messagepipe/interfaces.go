package messagepipe

// IMessagePipe provides a mean to emit errors and info messages from deep
// parts of the stack and decide on how to handle them further up.
type IMessagePipe interface {
	Error(error)
	Info(string)
	Debug(string)
	AddHandler(func(Severity, string))
	ClearHandlers()
}
