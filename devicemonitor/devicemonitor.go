package devicemonitor

import (
	"context"
	"fmt"
	"strings"

	udev "github.com/jochenvg/go-udev"

	"onplugd/device"
	"onplugd/deviceevent"
	"onplugd/messagepipe"
)

// The Udev API has two entry points: Udev itself, and the kernel. We'll be
// using the Udev interface. This is manifested as the name of the netlink
// against which we open the DBUS connection.
const netlinkUdev = "udev"

// The name of the udev.Device attribute that carries the kernel event
// properties.
const ueventAttr = "uevent"

// For now, only monitor USB and input device events.
var monitoredSubsystems = [...]string{"usb", "input"}

// UdevDeviceMonitor is an udev-based implementation of IDeviceMonitor.
type UdevDeviceMonitor struct {
	callbacks []func(deviceevent.IDeviceEvent) error
	records   map[string]device.IDevice
	pipe      messagepipe.IMessagePipe
	done      chan bool
}

// Start starts this device monitoring engine.
func (m *UdevDeviceMonitor) Start() error {
	if m.done != nil {
		close(m.done)
	}

	m.records = make(map[string]device.IDevice)
	m.done = make(chan bool)

	udev := udev.Udev{}
	monitor := udev.NewMonitorFromNetlink(netlinkUdev)

	for _, subsystem := range monitoredSubsystems {
		err := monitor.FilterAddMatchSubsystem(subsystem)
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	devices, err := monitor.DeviceChan(ctx)
	if err != nil {
		cancel()
		return err
	}

	err = m.doColdPlug(&udev)
	if err != nil {
		cancel()
		return err
	}

	go func() {
	out:
		for {
			select {
			case device := <-devices:
				event := m.actionToEvent(device.Action())
				m.processEvent(event, device)

			case <-m.done:
				cancel()
				break out
			}
		}

		m.pipe.Debug("UdevDeviceMonitor stopped.")
	}()

	m.pipe.Debug("UdevDeviceMonitor started.")

	return nil
}

func (m *UdevDeviceMonitor) doColdPlug(udev *udev.Udev) error {

	enumerate := udev.NewEnumerate()

	// Only get devices for which udev has finished the initialization.
	err := enumerate.AddMatchIsInitialized()
	if err != nil {
		return err
	}

	for _, subsystem := range monitoredSubsystems {
		err = enumerate.AddMatchSubsystem(subsystem)
		if err != nil {
			return err
		}
	}

	devices, err := enumerate.Devices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		m.processEvent(deviceevent.Coldplug, device)
	}
	return nil

}

func (m UdevDeviceMonitor) processEvent(event deviceevent.Event, dev *udev.Device) {

	if event == deviceevent.Unknown {
		return
	}

	if !dev.IsInitialized() {
		return
	}

	err := checkSubsystem(dev)
	if err != nil {
		m.pipe.Error(err)
		return
	}

	path := dev.Devpath()
	d, found := m.records[path]
	if !found {
		d = device.New(path)

		if event != deviceevent.Unbind && event != deviceevent.Remove {
			m.records[path] = d
		}
	}

	// Note that this assumes Remove events fire after Unbind events. This holds
	// out empirically but there seems to be no explicit guarantee that this will
	// always be the case.
	if event == deviceevent.Remove {
		delete(m.records, d.Path())
	}

	attrs := attrsFromUdevDevice(dev)
	uevent := ueventFromAttrs(attrs)

	d.SetSubsystem(dev.Subsystem())
	d.SetType(dev.Devtype())
	d.SetDriver(dev.Driver())

	for k, v := range attrs {
		d.Attrs()[k] = v
	}

	for k, v := range uevent {
		d.Uevent()[k] = v
	}

	e := deviceevent.New(event, d)

	m.pipe.Info(e.String())
	m.pipe.Debug(e.Device().Debug())

	for _, callback := range m.callbacks {
		err := callback(e)
		if err != nil {
			m.pipe.Error(err)
		}
	}
}

// Stop stops this device monitoring engine. It is idempotent and can safely be
// called multiple times.
func (m *UdevDeviceMonitor) Stop() error {

	if m.done != nil {
		close(m.done)
		m.done = nil
	}

	return nil
}

// AddCallback adds a callback to the device monitoring engine, which will be
// called when an event happens to a device.
func (m *UdevDeviceMonitor) AddCallback(f func(deviceevent.IDeviceEvent) error) {
	m.callbacks = append(m.callbacks, f)
}

// actionToEvent takes an udev.Device action string and returns the
// corresponding Event.
func (m *UdevDeviceMonitor) actionToEvent(action string) deviceevent.Event {

	event, found := map[string]deviceevent.Event{
		"add":    deviceevent.Add,
		"remove": deviceevent.Remove,
		"bind":   deviceevent.Bind,
		"unbind": deviceevent.Unbind,
		"change": deviceevent.Change,
		"move":   deviceevent.Move,
	}[action]

	if !found {
		m.pipe.Error(fmt.Errorf("Unknown event type: '%v'", action))
		return deviceevent.Unknown
	}
	return event
}

// New returns a new UdevDeviceMonitor.
func New(pipe messagepipe.IMessagePipe) UdevDeviceMonitor {
	return UdevDeviceMonitor{pipe: pipe}
}

func checkSubsystem(device *udev.Device) error {

	for _, expectedSubsystem := range monitoredSubsystems {
		if device.Subsystem() == expectedSubsystem {
			return nil
		}
	}

	return fmt.Errorf("Unexpected subsystem. Expected one of %v; got: %s",
		monitoredSubsystems, device.Subsystem())
}

func attrsFromUdevDevice(device *udev.Device) map[string]string {
	attrs := make(map[string]string)
	for k := range device.Sysattrs() {
		attrs[k] = device.SysattrValue(k)
	}
	return attrs
}

func ueventFromAttrs(attrs map[string]string) map[string]string {
	uevents := make(map[string]string)
	u, found := attrs[ueventAttr]
	if !found {
		return uevents
	}
	delete(attrs, ueventAttr)

	for _, line := range strings.Split(u, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		splitLine := strings.SplitN(line, "=", 2)
		if len(splitLine) < 2 {
			continue
		}

		k := strings.TrimSpace(splitLine[0])
		v := strings.TrimSpace(splitLine[1])
		uevents[k] = v
	}

	return uevents
}
