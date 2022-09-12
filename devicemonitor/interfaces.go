package devicemonitor

import "onplugd/deviceevent"

// IDeviceMonitor is the interface that describe a monitor, that being an
// object that implement a device status poller engine.
type IDeviceMonitor interface {
	Start() error
	Stop() error
	AddCallback(func(deviceevent.IDeviceEvent) error)
}
