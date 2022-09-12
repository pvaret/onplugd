package action

import (
	"fmt"
	"os"
	"path"
	"strings"

	"gopkg.in/ini.v1"

	"onplugd/deviceevent"
	"onplugd/executor"
)

// Action is an IAction implementation where the details of the action are
// stored in an INI file.
type Action struct {
	name string

	events     []string
	paths      []string
	subsystems []string
	types      []string
	drivers    []string
	attrs      map[string][]string
	uevents    map[string][]string

	execs []string
}

// Match checks if a given IDeviceEvent matches this action.
func (a *Action) Match(event deviceevent.IDeviceEvent) bool {

	if !foundIn(string(event.Event()), a.events) {
		return false
	}

	if !foundIn(event.Device().Path(), a.paths) {
		return false
	}

	if !foundIn(event.Device().Subsystem(), a.subsystems) {
		return false
	}

	if !foundIn(event.Device().Type(), a.types) {
		return false
	}

	if !foundIn(event.Device().Driver(), a.drivers) {
		return false
	}

	for attribute, values := range a.attrs {
		deviceAttrValue, found := event.Device().Attrs()[attribute]
		if !found {
			return false
		}
		if !foundIn(deviceAttrValue, values) {
			return false
		}
	}

	for uevent, values := range a.uevents {
		deviceUeventValue, found := event.Device().Uevent()[uevent]
		if !found {
			return false
		}
		if !foundIn(deviceUeventValue, values) {
			return false
		}
	}

	return true
}

// Do executes the action for the given event.
func (a *Action) Do(event deviceevent.IDeviceEvent, executor executor.IExecutor) error {

	env := os.Environ()

	env = append(env,
		"ONPLUGD_EVENT="+strings.ToUpper(string(event.Event())))
	env = append(env, "ONPLUGD_PATH="+event.Device().Path())
	env = append(env, "ONPLUGD_SUBSYSTEM="+event.Device().Subsystem())

	if driver := event.Device().Driver(); driver != "" {
		env = append(env, "ONPLUGD_DRIVER="+driver)
	}

	if typ := event.Device().Type(); typ != "" {
		env = append(env, "ONPLUGD_TYPE="+typ)
	}

	for attr, attrValue := range event.Device().Attrs() {
		attrEnv := fmt.Sprintf("ONPLUGD_ATTR_%s", strings.ToUpper(attr))
		env = append(env, attrEnv+"="+attrValue)
	}

	for uevent, ueventValue := range event.Device().Uevent() {
		ueventEnv := fmt.Sprintf("ONPLUGD_UEVENT_%s", strings.ToUpper(uevent))
		env = append(env, ueventEnv+"="+ueventValue)
	}

	for _, cmdline := range a.execs {
		executor.Exec(cmdline, env, a.name)
	}

	return nil
}

// NewActionFromFile creates a new action from the given file path.
func NewActionFromFile(fullpath string) (*Action, error) {

	a := Action{name: path.Base(fullpath)}

	// ShadowLoad (instead of Load) lets us list keys multiple times.
	conf, err := ini.ShadowLoad(fullpath)
	if err != nil {
		return nil, err
	}

	a.events = loadSliceFromShadow(
		conf.Section("match").Key("event").ValueWithShadows())
	if len(a.events) == 0 {
		a.events = []string{"COLDPLUG", "ADD"}
	}

	a.paths = loadSliceFromShadow(
		conf.Section("match").Key("path").ValueWithShadows())
	a.subsystems = loadSliceFromShadow(
		conf.Section("match").Key("subsystem").ValueWithShadows())
	a.types = loadSliceFromShadow(
		conf.Section("match").Key("type").ValueWithShadows())
	a.drivers = loadSliceFromShadow(
		conf.Section("match").Key("driver").ValueWithShadows())

	a.attrs = make(map[string][]string)
	err = loadMapFromShadow(
		a.attrs, conf.Section("match").Key("attr").ValueWithShadows())
	if err != nil {
		return nil, err
	}

	a.uevents = make(map[string][]string)
	err = loadMapFromShadow(
		a.uevents, conf.Section("match").Key("uevent").ValueWithShadows())
	if err != nil {
		return nil, err
	}

	a.execs = conf.Section("action").Key("exec").ValueWithShadows()

	return &a, nil
}

func match(s1, s2 string) bool {
	return len(s1) > 0 && strings.ToLower(s1) == strings.ToLower(s2)
}

func foundIn(needle string, haystack []string) bool {
	if len(haystack) == 0 {
		return true
	}

	for _, hay := range haystack {
		if match(needle, hay) {
			return true
		}
	}

	return false
}

// Populate a map from a slice of entries that look like "k=v"
func loadMapFromShadow(m map[string][]string, shadow []string) error {
	for _, entry := range shadow {

		// Shadowloading adds at least one empty value to the shadow variable, let's
		// filter that out.
		if entry == "" {
			continue
		}

		keyvalue := strings.SplitN(entry, "=", 2)
		if len(keyvalue) != 2 {
			return fmt.Errorf(
				"Invalid attribute: expected something formatted as KEY=VALUE, got '%s'", entry)
		}
		k := strings.TrimSpace(keyvalue[0])
		v := strings.TrimSpace(keyvalue[1])
		m[k] = append(m[k], v)
	}

	return nil
}

func loadSliceFromShadow(shadow []string) []string {
	var val []string

	for _, item := range shadow {
		// Shadowloading adds at least one empty value to the shadow variable, let's
		// filter that out.
		if item != "" {
			val = append(val, item)
		}
	}

	return val
}
