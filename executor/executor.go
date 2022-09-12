package executor

import (
	"context"
	"fmt"
	"os/exec"
	"path"

	"onplugd/messagepipe"
	"onplugd/utils"
)

// An Executor can safely run a command line in a given context.
type Executor struct {
	pipe    messagepipe.IMessagePipe
	context context.Context
}

// Exec runs the given command line with the given environment in a goroutine.
// It returns as soon as the command is started without waiting for it to
// complete.
func (e *Executor) Exec(cmdline string, env []string, prefix string) {

	cmdline = utils.Expand(cmdline)
	cmd := exec.CommandContext(e.context, "/bin/sh", "-c", cmdline)

	cmd.Env = env
	cmd.Dir = path.Dir("/")
	cmd.Stdin = nil // explicitly close stdin

	e.pipe.Info(fmt.Sprintf("Executing '%s'", cmdline))
	e.pipe.Debug(fmt.Sprintf("Environment: %v", env))

	go func() {

		stdout := &pipeWriter{prefix: "STDOUT (" + prefix + "):", pipe: e.pipe}
		stderr := &pipeWriter{prefix: "STDERR (" + prefix + "):", pipe: e.pipe}
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		defer stdout.Flush()
		defer stderr.Flush()

		if err := cmd.Run(); err != nil {
			e.pipe.Error(
				fmt.Errorf("Command '%s' failed with status %s", cmdline, err))
		}
	}()
}

// New returns a new executor, as well as the cleanup function to call when
// shutting down.
func New(pipe messagepipe.IMessagePipe) (*Executor, func()) {

	context, cancel := context.WithCancel(context.Background())

	return &Executor{
		pipe:    pipe,
		context: context,
	}, cancel
}
