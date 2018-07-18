package ini

import (
	"strings"
	"fmt"
	"strconv"
	"os"
)

/*************************************************************
 * data get
 *************************************************************/

// Get a value by key string. you can use '.' split for get value in a special section
func (ini *Ini) Get(key string) (val string, ok bool) {
	key = strings.Trim(strings.TrimSpace(key), ".")
	if key == "" {
		return
	}

	// default find from default Section
	name := DefSection

	// get val by path. eg "log.dir"
	if strings.Contains(key, ".") {
		ss := strings.SplitN(key, ".", 2)
		name, key = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
	}

	// get section data
	sec, ok := ini.data[name]
	if ok {
		// find from section
		val, ok = sec[key]
	}

	return
}

// GetInt get a int value
func (ini *Ini) GetInt(key string) (val int, ok bool) {
	rawVal, ok := ini.Get(key)
	if !ok {
		return
	}

	if val, err := strconv.Atoi(rawVal); err == nil {
		return val, true
	}

	return
}

// DefInt get a int value, if not found return default value
func (ini *Ini) DefInt(key string, def int) (val int) {
	if val, ok := ini.GetInt(key); ok {
		return val
	}

	return def
}

// MustInt get a int value, if not found return 0
func (ini *Ini) MustInt(key string) int {
	return ini.DefInt(key, 0)
}

// GetBool Looks up a value for a key in this section and attempts to parse that value as a boolean,
// along with a boolean result similar to a map lookup.
// of following(case insensitive):
//  - true
//  - false
//  - yes
//  - no
//  - off
//  - on
//  - 0
//  - 1
// The `ok` boolean will be false in the event that the value could not be parsed as a bool
func (ini *Ini) GetBool(key string) (value bool, ok bool) {
	rawVal, ok := ini.Get(key)
	if !ok {
		return
	}

	lowerCase := strings.ToLower(rawVal)
	switch lowerCase {
	case "", "0", "false", "no", "off":
		value = false
	case "1", "true", "yes", "on":
		value = true
	default:
		ok = false
	}

	return
}

// DefBool get a bool value, if not found return default value
func (ini *Ini) DefBool(key string, def bool) bool {
	if value, ok := ini.GetBool(key); ok {
		return value
	}

	return def
}

// MustBool get a string value, if not found return false
func (ini *Ini) MustBool(key string) bool {
	return ini.DefBool(key, false)
}

// GetString like Get method, but will parse ENV value.
func (ini *Ini) GetString(key string) (val string, ok bool) {
	val, ok = ini.Get(key)
	if !ok {
		return
	}

	// if opts.ParseEnv is true. will parse like: "${SHELL}"
	if ini.opts.ParseEnv && strings.Index(val, "${") == 0 {
		var name, def string
		str := strings.Trim(strings.TrimSpace(val), "${}")
		ss := strings.SplitN(str, "|", 2)

		// ${NotExist|defValue}
		if len(ss) == 2 {
			name, def = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
			// ${SHELL}
		} else {
			name = ss[0]
		}

		val = os.Getenv(name)
		if val == "" {
			val = def
		}
	}

	return
}

// DefString get a string value, if not found return default value
func (ini *Ini) DefString(key string, def string) string {
	if value, ok := ini.GetString(key); ok {
		return value
	}

	return def
}

// MustString get a string value, if not found return empty string
func (ini *Ini) MustString(key string) string {
	return ini.DefString(key, "")
}

// GetSection
func (ini *Ini) GetSection(name string) (sec map[string]string, ok bool) {
	sec, ok = ini.data[name]
	return
}

/*************************************************************
 * config set
 *************************************************************/

// Set a value to the section by key.
// if section is empty, will set to default section
func (ini *Ini) Set(key, val string, section ...string) {
	key = strings.Trim(strings.TrimSpace(key), ".")
	if key == "" {
		return
	}

	name := DefSection
	if len(section) > 0 {
		name = section[0]
	}

	if ini.opts.IgnoreCase {
		key = strings.ToLower(key)
		name = strings.ToLower(name)
	}

	sec, ok := ini.data[name]
	if ok {
		sec[key] = val
	} else {
		sec = Section{key: val}
	}

	ini.data[name] = sec
}

// SetInt
func (ini *Ini) SetInt(key string, val int, section ...string) {
	ini.Set(key, fmt.Sprintf("%d", val), section...)
}

// SetBool
func (ini *Ini) SetBool(key string, val bool, section ...string) {
	valStr := "false"
	if val {
		valStr = "true"
	}

	ini.Set(key, valStr, section...)
}

// SetString
func (ini *Ini) SetString(key, val string, section ...string) {
	ini.Set(key, val, section...)
}

// SetSection
func (ini *Ini) SetSection(name string, values map[string]string) {
	if old, ok := ini.data[name]; ok {
		if ini.opts.IgnoreCase {
			name = strings.ToLower(name)
		}

		ini.data[name] = mergeStringMap(values, old, ini.opts.IgnoreCase)
	} else {
		ini.AddSection(name, values)
	}
}

// AddSection
func (ini *Ini) AddSection(name string, values map[string]string) {
	if ini.opts.IgnoreCase {
		name = strings.ToLower(name)

		ini.data[name] = mapKeyToLower(values)
	} else {
		ini.data[name] = values
	}
}

// MergeData
func (ini *Ini) MergeData(data map[string]Section) {
	ini.ensureInit()
	if len(ini.data) == 0 {
		ini.data = data
	}

	// append or override setting data
	for name, sec := range data {
		ini.SetSection(name, sec)
	}
}

/*************************************************************
 * helper methods
 *************************************************************/

// Reset all data
func (ini *Ini) Reset() {
	ini.data = make(map[string]Section)
}

// Data get all data
func (ini *Ini) Data() map[string]Section {
	return ini.data
}

// HasSection
func (ini *Ini) HasSection(name string) bool {
	_, ok := ini.data[name]
	return ok
}
