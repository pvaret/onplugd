package executor

import (
	"fmt"

	"onplugd/messagepipe"
)

// pipeWriter is a io.Writer implementation that can be used to replace
// cmd.Std{out,err} and write to the given IMessagePipe instead.
type pipeWriter struct {
	prefix string
	pipe   messagepipe.IMessagePipe
	buffer []byte
}

func (w *pipeWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if b == '\n' {
			w.pipe.Info(fmt.Sprint(w.prefix, " ", string(w.buffer)))
			w.buffer = w.buffer[:0]
		} else {
			w.buffer = append(w.buffer, b)
		}
	}

	return len(p), nil
}

func (w *pipeWriter) Flush() {
	if len(w.buffer) > 0 {
		w.pipe.Info(fmt.Sprint(w.prefix, " ", string(w.buffer)))
	}
}
