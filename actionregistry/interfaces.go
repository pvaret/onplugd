package actionregistry

import (
	"onplugd/action"
	"onplugd/deviceevent"
)

// IActionRegistry is the interface that describes the registry of device event
// actions.
type IActionRegistry interface {
	OnDeviceEvent(event deviceevent.IDeviceEvent)
	Update(name string, action action.IAction)
	Remove(name string)
}
