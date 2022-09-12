package deviceevent

import (
	"fmt"

	"onplugd/device"
)

// IDeviceEvent is the interface for objects that describe an event that just
// happened to a device.
type IDeviceEvent interface {
	fmt.Stringer

	Device() device.IDevice
	Event() Event
}
