package confmonitor

// IConfMonitor is the interface that describe a configuration monitor.
type IConfMonitor interface {
	Start() error
	Stop()
	Events() <-chan FileEvent
}
