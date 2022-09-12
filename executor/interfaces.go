package executor

// IExecutor describes a utility to run commands safely.
type IExecutor interface {
	Exec(cmdline string, env []string, prefix string)
}
