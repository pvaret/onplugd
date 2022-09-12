package actionregistry

import (
	"fmt"
	"sync"

	"onplugd/action"
	"onplugd/deviceevent"
	"onplugd/executor"
	"onplugd/messagepipe"
)

// ActionRegistry contains the actions that we know about.
type ActionRegistry struct {
	actions  map[string]action.IAction
	executor executor.IExecutor
	lock     sync.RWMutex
	pipe     messagepipe.IMessagePipe
}

// New creates and returns an ActionRegistry instance.
func New(
	pipe messagepipe.IMessagePipe, executor executor.IExecutor) *ActionRegistry {
	ar := ActionRegistry{
		actions:  make(map[string]action.IAction),
		executor: executor,
		pipe:     pipe,
	}

	return &ar
}

// OnDeviceEvent calls all the matching actions when a new device event
// arrives.
func (ar *ActionRegistry) OnDeviceEvent(event deviceevent.IDeviceEvent) {
	ar.lock.RLock()

	var actions []action.IAction
	for name, action := range ar.actions {
		if action.Match(event) {
			actions = append(actions, action)
			ar.pipe.Debug(fmt.Sprint("Match found: ", name))
		}
	}

	ar.lock.RUnlock()

	for _, a := range actions {
		go func(action action.IAction) {
			err := action.Do(event, ar.executor)
			if err != nil {
				ar.pipe.Error(err)
			}
		}(a)
	}
}

// Update updates an IAction in the registry, by name.
func (ar *ActionRegistry) Update(name string, action action.IAction) {
	ar.lock.Lock()
	defer ar.lock.Unlock()
	ar.actions[name] = action
}

// Remove removes an IAction from the registry, by name.
func (ar *ActionRegistry) Remove(name string) {
	ar.lock.Lock()
	defer ar.lock.Unlock()
	delete(ar.actions, name)
}
