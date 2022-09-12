package action

import (
	"onplugd/deviceevent"
	"onplugd/executor"
)

// IAction is the interface that describes an action that can be taken as the
// result of a device event.
type IAction interface {
	Match(deviceevent.IDeviceEvent) bool
	Do(deviceevent.IDeviceEvent, executor.IExecutor) error
}
