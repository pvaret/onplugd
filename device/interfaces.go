package device

import "fmt"

// IDevice is the interface that describes the properties of a device.
type IDevice interface {
	fmt.Stringer

	SetSubsystem(string)
	SetType(string)
	SetDriver(string)

	Path() string
	Subsystem() string
	Type() string
	Driver() string

	Attrs() map[string]string
	Uevent() map[string]string

	Debug() string
}
