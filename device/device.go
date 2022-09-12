package device

import (
	"fmt"
	"sort"
)

// Device is an implementation of IDevice.
type Device struct {
	path      string
	subsystem string
	typ       string
	driver    string
	attrs     map[string]string
	uevent    map[string]string
}

func (d Device) String() string {
	str := fmt.Sprintf("[%s] (%s)",
		d.path, d.subsystem)
	if d.typ != "" {
		str += fmt.Sprintf(" type:%s", d.typ)
	}
	if d.driver != "" {
		str += fmt.Sprintf(" driver:%s", d.driver)
	}

	return str
}

// Debug dumps a detailed description of the device.
func (d Device) Debug() string {
	str := d.String() + "\n"

	attrs := []string{}
	for attr := range d.attrs {
		attrs = append(attrs, attr)
	}

	str += "ATTRS:\n"
	sort.Strings(attrs)
	for _, attr := range attrs {
		str += fmt.Sprintf("  %s=%s\n", attr, d.attrs[attr])
	}

	uevents := []string{}
	for uevent := range d.uevent {
		uevents = append(uevents, uevent)
	}

	str += "UEVENT:\n"
	sort.Strings(uevents)
	for _, uevent := range uevents {
		str += fmt.Sprintf("  %s=%s\n", uevent, d.uevent[uevent])
	}

	return str
}

// SetSubsystem sets the udev subsystem of the device.
func (d *Device) SetSubsystem(subsystem string) { d.subsystem = subsystem }

// SetType sets the udev type of the device.
func (d *Device) SetType(typ string) { d.typ = typ }

// SetDriver sets the Linux driver associated with the device.
func (d *Device) SetDriver(driver string) { d.driver = driver }

// Path returns the udev path of the device.
func (d *Device) Path() string { return d.path }

// Subsystem returns the udev subsystem of the device.
func (d *Device) Subsystem() string { return d.subsystem }

// Type returns the udev type of the device.
func (d *Device) Type() string { return d.typ }

// Driver returns the Linux driver associated with the device.
func (d *Device) Driver() string { return d.driver }

// Attrs returns the device's attribute map as exported by udev.
func (d *Device) Attrs() map[string]string { return d.attrs }

// Uevent returns the device's uevent map as exported by udev.
func (d *Device) Uevent() map[string]string { return d.uevent }

// New creates a new device for the given path.
func New(path string) *Device {
	return &Device{
		path:   path,
		attrs:  make(map[string]string),
		uevent: make(map[string]string),
	}
}
