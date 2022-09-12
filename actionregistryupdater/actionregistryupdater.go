package actionregistryupdater

import (
	"fmt"
	"path"

	"onplugd/action"
	"onplugd/actionregistry"
	"onplugd/confmonitor"
	"onplugd/messagepipe"
)

// ActionRegistryUpdater applies configuration updates from an IConfMonitor to
// an IActionRegistry.
type ActionRegistryUpdater struct {
	registry actionregistry.IActionRegistry
	monitor  confmonitor.IConfMonitor
	pipe     messagepipe.IMessagePipe

	done chan bool
}

// New creates a new ActionRegistryUpdater.
func New(
	registry actionregistry.IActionRegistry, monitor confmonitor.IConfMonitor,
	pipe messagepipe.IMessagePipe) ActionRegistryUpdater {
	aru := ActionRegistryUpdater{
		registry: registry,
		monitor:  monitor,
		pipe:     pipe,
	}
	return aru
}

// Start starts the ActionRegistryUpdater loop. Yeah.
func (aru *ActionRegistryUpdater) Start() error {
	err := aru.monitor.Start()
	if err != nil {
		return err
	}

	aru.done = make(chan bool)

	go func() {
		events := aru.monitor.Events()
	out:
		for {

			select {
			case <-aru.done:
				break out

			case event, ok := <-events:
				if !ok {
					break out
				}

				name := path.Base(event.Name)

				if event.Event == confmonitor.FileDelete {
					aru.pipe.Info(fmt.Sprint("Conf file removed: ", name))
					aru.registry.Remove(name)

				} else { // Create or Update
					action, err := action.NewActionFromFile(event.Name)

					if err != nil {
						aru.pipe.Error(fmt.Errorf(
							"Error while reading %s: %s", event.Name, err))
						continue
					}

					if event.Event == confmonitor.FileCreate {
						aru.pipe.Info(fmt.Sprint("Conf file added: ", name))
					} else { // Event is FileChange
						aru.pipe.Info(fmt.Sprint("Conf file modified: ", name))
					}

					aru.registry.Update(name, action)
				}
			}
		}

		aru.monitor.Stop()
	}()

	return nil
}

// Stop stops the ActionRegistryUpdater loop.
func (aru *ActionRegistryUpdater) Stop() {

	if aru.done != nil {
		close(aru.done)
		aru.done = nil
	}
}
