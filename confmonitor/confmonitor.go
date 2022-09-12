package confmonitor

import (
	"os"
	"path"
	"path/filepath"

	"github.com/fsnotify/fsnotify"

	"onplugd/messagepipe"
)

// FileEventType characterizes an event that can happen on a file.
type FileEventType uint8

const confPattern = "*.conf"

const (
	// FileCreate is the event that indicates a file creation.
	FileCreate FileEventType = iota
	// FileChange is the event that indicates a file modification.
	FileChange
	// FileDelete is the event that indicates a file deletion.
	FileDelete
)

// FileEvent captures an event on a file.
type FileEvent struct {
	Event FileEventType
	Name  string
}

// ConfMonitor is an fsnotify-based implementation of IConfMonitor.
type ConfMonitor struct {
	path    string
	events  chan FileEvent
	done    chan bool
	started bool
	pipe    messagepipe.IMessagePipe
}

// New creates a new ConfMonitor.
func New(
	path string, pipe messagepipe.IMessagePipe) ConfMonitor {

	return ConfMonitor{path: path, pipe: pipe}
}

// Start sets up the monitor and starts the monitoring goroutine.
func (m *ConfMonitor) Start() error {
	if m.started {
		return nil
	}

	if _, err := path.Match("", confPattern); err != nil {
		// Invalid pattern.
		return err
	}

	m.Stop()

	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		err = os.MkdirAll(m.path, 0755)
		if err != nil {
			return err
		}
	}

	m.done = make(chan bool)
	m.events = make(chan FileEvent)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	watcher.Add(m.path)

	go func() {

		// Populate existing files.
		files, _ := filepath.Glob(path.Join(m.path, confPattern))
		for _, f := range files {
			m.events <- FileEvent{Event: FileCreate, Name: f}
		}

	out:
		for {
			select {
			case <-m.done:
				break out

			case event, ok := <-watcher.Events:
				if !ok {
					break out
				}

				// Only match events on paths that have the proper suffix.
				if matching, err := path.Match(
					confPattern, path.Base(event.Name)); !matching || err != nil {
					continue
				}

				if event.Op&fsnotify.Create != 0 {
					m.events <- FileEvent{Event: FileCreate, Name: event.Name}
				} else if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					m.events <- FileEvent{Event: FileDelete, Name: event.Name}
				} else if event.Op&fsnotify.Write != 0 {
					m.events <- FileEvent{Event: FileChange, Name: event.Name}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					break out
				}

				m.pipe.Error(err)
			}
		}

		watcher.Close()
		close(m.events)
		m.started = false

		m.pipe.Debug("ConfMonitor stopped.")
	}()

	m.pipe.Debug("ConfMonitor started.")
	m.started = true
	return nil
}

// Stop stops the monitoring goroutine.
func (m *ConfMonitor) Stop() {

	if m.done != nil {
		close(m.done)
		m.done = nil
	}
}

// Events returns the configuration monitor file event channel. It starts the
// monitoring if needed.
func (m *ConfMonitor) Events() <-chan FileEvent {
	m.Start()
	return m.events
}
