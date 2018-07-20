package ini

import (
	"strings"
	"fmt"
	"strconv"
	"os"
	"errors"
)

// if is readonly
var cannotSetInReadonly = errors.New("ini: The config manager instance in 'readonly' mode")

/*************************************************************
 * data get
 *************************************************************/

// Get a value by key string. you can use '.' split for get value in a special section
func (ini *Ini) Get(key string) (val string, ok bool) {
	// if not is readonly
	if !ini.opts.Readonly {
		ini.lock.Lock()
		defer ini.lock.Unlock()
	}

	key = formatKey(key)
	if key == "" {
		return
	}

	if ini.opts.IgnoreCase {
		key = strings.ToLower(key)
	}

	name, key := ini.splitSectionAndKey(key, SepSection)

	// get section data
	sec, ok := ini.data[name]
	if ok {
		val, ok = sec[key]
	}

	return
}

// Int get a int value
func (ini *Ini) Int(key string) (val int, ok bool) {
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
	if val, ok := ini.Int(key); ok {
		return val
	}

	return def
}

// MustInt get a int value, if not found return 0
func (ini *Ini) MustInt(key string) int {
	return ini.DefInt(key, 0)
}

// Bool Looks up a value for a key in this section and attempts to parse that value as a boolean,
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
func (ini *Ini) Bool(key string) (value bool, ok bool) {
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
	if value, ok := ini.Bool(key); ok {
		return value
	}

	return def
}

// MustBool get a string value, if not found return false
func (ini *Ini) MustBool(key string) bool {
	return ini.DefBool(key, false)
}

// GetString like Get method, but will parse ENV value.
func (ini *Ini) String(key string) (val string, ok bool) {
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
	if value, ok := ini.String(key); ok {
		return value
	}

	return def
}

// MustString get a string value, if not found return empty string
func (ini *Ini) MustString(key string) string {
	return ini.DefString(key, "")
}

// StringMap
func (ini *Ini) StringMap(name string) (mp map[string]string, ok bool) {
	if ini.opts.IgnoreCase {
		name = strings.ToLower(name)
	}

	mp, ok = ini.data[name]
	return
}

// Section
func (ini *Ini) Section(name string) (sec map[string]string, ok bool) {
	return ini.StringMap(name)
}

/*************************************************************
 * config set
 *************************************************************/

// Set a value to the section by key.
// if section is empty, will set to default section
func (ini *Ini) Set(key, val string, section ...string) (err error) {
	// if is readonly
	if ini.opts.Readonly {
		return cannotSetInReadonly
	} else {
		ini.lock.Lock()
		defer ini.lock.Unlock()
	}

	key = formatKey(key)
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
	return
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

/*************************************************************
 * section operate
 *************************************************************/

// SetSection if not exist, add new section. If exist, will merge to old section.
func (ini *Ini) SetSection(name string, values map[string]string) (err error) {
	// if is readonly
	if ini.opts.Readonly {
		return cannotSetInReadonly
	}

	if ini.opts.IgnoreCase {
		name = strings.ToLower(name)
	}

	if old, ok := ini.data[name]; ok {
		ini.data[name] = mergeStringMap(values, old, ini.opts.IgnoreCase)
	} else {
		if ini.opts.IgnoreCase {
			values = mapKeyToLower(values)
		}

		ini.data[name] = values
	}

	return
}

// NewSection add new section data, existed will be replace
func (ini *Ini) NewSection(name string, values map[string]string) (err error) {
	// if is readonly
	if ini.opts.Readonly {
		return cannotSetInReadonly
	}

	if ini.opts.IgnoreCase {
		name = strings.ToLower(name)
		ini.data[name] = mapKeyToLower(values)
	} else {
		ini.data[name] = values
	}
	return
}

// HasSection
func (ini *Ini) HasSection(name string) bool {
	_, ok := ini.data[name]
	return ok
}

// DelSection
func (ini *Ini) DelSection(name string) bool {
	// if is readonly
	if ini.opts.Readonly {
		return false
	}

	if ini.opts.IgnoreCase {
		name = strings.ToLower(name)
	}

	_, ok := ini.data[name]
	if ok {
		delete(ini.data, name)
	}

	return ok
}

/*************************************************************
 * helper methods
 *************************************************************/

// HasKey
func (ini *Ini) HasKey(key string) (ok bool) {
	_, ok = ini.Get(key)
	return
}

// Del
func (ini *Ini) Del(key string) (ok bool) {
	// if is readonly
	if ini.opts.Readonly {
		return
	}

	key = formatKey(key)
	if key == "" {
		return
	}

	if ini.opts.IgnoreCase {
		key = strings.ToLower(key)
	}

	sec, key := ini.splitSectionAndKey(key, SepSection)
	mp, ok := ini.data[sec]
	if !ok {
		return
	}

	if _, ok = mp[key]; ok {
		delete(mp, key)
		ini.data[sec] = mp
	}

	return
}

// Reset all data
func (ini *Ini) Reset() {
	ini.data = make(map[string]Section)
}

// Data get all data
func (ini *Ini) Data() map[string]Section {
	return ini.data
}

func (ini *Ini) splitSectionAndKey(key, sep string) (string, string) {
	// default find from default Section
	name := DefSection

	// get val by path. eg "log.dir"
	if strings.Contains(key, sep) {
		ss := strings.SplitN(key, sep, 2)
		name, key = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
	}

	return name, key
}

// format key
func formatKey(key string) string {
	return strings.Trim(strings.TrimSpace(key), ".")
}

// simple merge two string map
func mergeStringMap(src, dst map[string]string, ignoreCase bool) map[string]string {
	for k, v := range src {
		if ignoreCase {
			k = strings.ToLower(k)
		}

		dst[k] = v
	}

	return dst
}

func mapKeyToLower(src map[string]string) map[string]string {
	newMp := make(map[string]string)

	for k, v := range src {
		k = strings.ToLower(k)
		newMp[k] = v
	}

	return newMp
}
