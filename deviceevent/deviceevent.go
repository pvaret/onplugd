package deviceevent

import (
	"fmt"

	"onplugd/device"
)

// An Event is a noteworthy change that may happen to the status of a device.
type Event string

const (
	// Add is when a device is added.
	Add Event = "ADD"
	// Remove is when a device is removed.
	Remove Event = "REMOVE"
	// Bind is when a device is... bound I guess?
	Bind Event = "BIND"
	// Unbind is when a device is unbound.
	Unbind Event = "UNBIND"
	// Change is when something changes with the device, for instance a media is
	// inserted in an already plugged reader.
	Change Event = "CHANGE"
	// Move is when a device's path changes, for instance a network interface
	// gets renamed to a stable name.
	Move Event = "MOVE"
	// Coldplug is when a device is detected as already being there when
	// onplugd starts.
	Coldplug Event = "COLDPLUG"
	// Unknown is when we have no clue what happened with the device.
	Unknown Event = "?unknown event?"
)

// DeviceEvent is an implementation of IDeviceEvent.
type DeviceEvent struct {
	device device.IDevice
	event  Event
}

func (e DeviceEvent) String() string {
	return fmt.Sprintf("Event: %s; Device: %s", e.event, e.device)
}

// Device implements IDeviceEvent.Device for DeviceEvent.
func (e DeviceEvent) Device() device.IDevice {
	return e.device
}

// Event implements IDeviceEvent.Event for DeviceEvent.
func (e DeviceEvent) Event() Event {
	return e.event
}

// New creates a new DeviceEvent for the given event and device.
func New(event Event, dev device.IDevice) *DeviceEvent {
	return &DeviceEvent{
		event:  event,
		device: dev,
	}
}
