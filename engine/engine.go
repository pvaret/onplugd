package engine

import (
	"onplugd/actionregistry"
	"onplugd/actionregistryupdater"
	"onplugd/confmonitor"
	"onplugd/deviceevent"
	"onplugd/devicemonitor"
	"onplugd/messagepipe"
)

// Engine runs the main loop of onplugd.
type Engine struct {
	started               bool
	deviceMonitor         devicemonitor.IDeviceMonitor
	confMonitor           confmonitor.IConfMonitor
	actionRegistry        actionregistry.IActionRegistry
	actionRegistryUpdater actionregistryupdater.ActionRegistryUpdater
	pipe                  messagepipe.IMessagePipe
	cleanups              []func()
}

// New instantiates and returns a new Engine.
func New(
	deviceMonitor devicemonitor.IDeviceMonitor,
	confMonitor confmonitor.IConfMonitor,
	actionRegistry actionregistry.IActionRegistry,
	messagePipe messagepipe.IMessagePipe) Engine {

	updater := actionregistryupdater.New(
		actionRegistry, confMonitor, messagePipe)

	e := Engine{
		deviceMonitor:         deviceMonitor,
		confMonitor:           confMonitor,
		actionRegistry:        actionRegistry,
		actionRegistryUpdater: updater,
		pipe:                  messagePipe,
	}

	deviceMonitor.AddCallback(e.onDeviceEvent)

	return e
}

// Start starts the loop, or restarts it if already running.
func (e *Engine) Start() error {

	// Ensure we aren't already running.
	e.Stop()

	// The order here matters: first we get ready to apply configurations, then we
	// start reading configurations, then we start waiting for devices.
	e.actionRegistryUpdater.Start()
	e.confMonitor.Start()
	e.deviceMonitor.Start()

	e.pipe.Debug("Engine started.")
	e.started = true

	return nil
}

// Stop stops the loop.
func (e *Engine) Stop() error {

	if e.started {
		e.deviceMonitor.Stop()
		e.confMonitor.Stop()
		e.actionRegistryUpdater.Stop()
		e.pipe.Debug("Engine stopped.")
		e.started = false

		for _, f := range e.cleanups {
			f()
		}
	}

	return nil
}

// onDeviceEvent provides the entry point for device events, to be used as a
// callback by IDeviceMonitor instances.
func (e *Engine) onDeviceEvent(event deviceevent.IDeviceEvent) error {
	e.actionRegistry.OnDeviceEvent(event)
	return nil
}

// AddCleanupCallback adds a callback to the engine that will be called at turndown time.
func (e *Engine) AddCleanupCallback(callback func()) {
	e.cleanups = append(e.cleanups, callback)
}
