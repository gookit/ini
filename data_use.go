package ini

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// if is readonly
var errSetInReadonly = errors.New("ini: The config manager instance in 'readonly' mode")

/*************************************************************
 * data get
 *************************************************************/

// Get a value by key string. you can use '.' split for get value in a special section
func (c *Ini) Get(key string) (val string, ok bool) {
	// if not is readonly
	if !c.opts.Readonly {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	sep := c.opts.SectionSep
	key = formatKey(key, sep)
	if key == "" {
		return
	}

	if c.opts.IgnoreCase {
		key = strings.ToLower(key)
	}

	// get section data
	name, key := c.splitSectionAndKey(key, sep)
	strMap, ok := c.data[name]
	if !ok {
		return
	}

	val, ok = strMap[key]

	// if enable parse var
	if c.opts.ParseVar {
		// must close lock. because parseVarReference() maybe loop call Get()
		if !c.opts.Readonly {
			c.lock.Unlock()
			val = c.parseVarReference(key, val, strMap)
			c.lock.Lock()
		} else {
			val = c.parseVarReference(key, val, strMap)
		}
	}

	return
}

// Int get a int value
func (c *Ini) Int(key string) (val int, ok bool) {
	rawVal, ok := c.Get(key)
	if !ok {
		return
	}

	if val, err := strconv.Atoi(rawVal); err == nil {
		return val, true
	}

	ok = false
	return
}

// DefInt get a int value, if not found return default value
func (c *Ini) DefInt(key string, def int) (val int) {
	if val, ok := c.Int(key); ok {
		return val
	}

	return def
}

// MustInt get a int value, if not found return 0
func (c *Ini) MustInt(key string) int {
	return c.DefInt(key, 0)
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
func (c *Ini) Bool(key string) (value bool, ok bool) {
	rawVal, ok := c.Get(key)
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
func (c *Ini) DefBool(key string, def bool) bool {
	if value, ok := c.Bool(key); ok {
		return value
	}

	return def
}

// MustBool get a string value, if not found return false
func (c *Ini) MustBool(key string) bool {
	return c.DefBool(key, false)
}

// GetString like Get method
func (c *Ini) String(key string) (val string, ok bool) {
	return c.Get(key)
}

// DefString get a string value, if not found return default value
func (c *Ini) DefString(key string, def string) string {
	if value, ok := c.String(key); ok {
		return value
	}

	return def
}

// MustString get a string value, if not found return empty string
func (c *Ini) MustString(key string) string {
	return c.DefString(key, "")
}

// Strings get a string array, by split a string
func (c *Ini) Strings(key, sep string) (ss []string, ok bool) {
	str, ok := c.Get(key)
	if !ok {
		return
	}

	return strings.Split(str, sep), ok
}

// StringMap get a section data map
func (c *Ini) StringMap(name string) (mp map[string]string, ok bool) {
	if c.opts.IgnoreCase {
		name = strings.ToLower(name)
	}

	// empty name, return default section
	if name == "" {
		name = c.opts.DefSection
	}

	mp, ok = c.data[name]

	// parser Var ref
	if c.opts.ParseVar {
		for k, v := range mp {
			mp[k] = c.parseVarReference(k, v, mp)
		}
	}

	return
}

// MustMap must return a string map
func (c *Ini) MustMap(name string) map[string]string {
	if mp, ok := c.StringMap(name); ok {
		return mp
	}

	// empty map
	return map[string]string{}
}

// Section get a section data map
func (c *Ini) Section(name string) (sec Section, ok bool) {
	return c.StringMap(name)
}

/*************************************************************
 * config set
 *************************************************************/

// Set a value to the section by key.
// if section is empty, will set to default section
func (c *Ini) Set(key, val string, section ...string) (err error) {
	// if is readonly
	if c.opts.Readonly {
		return errSetInReadonly
	}

	c.ensureInit()

	// open lock
	c.lock.Lock()
	defer c.lock.Unlock()

	sep := c.opts.SectionSep
	key = formatKey(key, sep)
	if key == "" {
		return
	}

	name := c.opts.DefSection
	if len(section) > 0 {
		name = section[0]
	}

	if c.opts.IgnoreCase {
		key = strings.ToLower(key)
		name = strings.ToLower(name)
	}

	sec, ok := c.data[name]
	if ok {
		sec[key] = val
	} else {
		sec = Section{key: val}
	}

	c.data[name] = sec
	return
}

// SetInt set a int by key
func (c *Ini) SetInt(key string, val int, section ...string) {
	c.Set(key, fmt.Sprintf("%d", val), section...)
}

// SetBool set a bool by key
func (c *Ini) SetBool(key string, val bool, section ...string) {
	valStr := "false"
	if val {
		valStr = "true"
	}

	c.Set(key, valStr, section...)
}

// SetString set a string by key
func (c *Ini) SetString(key, val string, section ...string) {
	c.Set(key, val, section...)
}

/*************************************************************
 * section operate
 *************************************************************/

// SetSection if not exist, add new section. If exist, will merge to old section.
func (c *Ini) SetSection(name string, values map[string]string) (err error) {
	// if is readonly
	if c.opts.Readonly {
		return errSetInReadonly
	}

	if c.opts.IgnoreCase {
		name = strings.ToLower(name)
	}

	if old, ok := c.data[name]; ok {
		c.data[name] = mergeStringMap(values, old, c.opts.IgnoreCase)
	} else {
		if c.opts.IgnoreCase {
			values = mapKeyToLower(values)
		}

		c.data[name] = values
	}

	return
}

// NewSection add new section data, existed will be replace
func (c *Ini) NewSection(name string, values map[string]string) (err error) {
	// if is readonly
	if c.opts.Readonly {
		return errSetInReadonly
	}

	if c.opts.IgnoreCase {
		name = strings.ToLower(name)
		c.data[name] = mapKeyToLower(values)
	} else {
		c.data[name] = values
	}
	return
}

// HasSection has section
func (c *Ini) HasSection(name string) bool {
	if c.opts.IgnoreCase {
		name = strings.ToLower(name)
	}

	_, ok := c.data[name]
	return ok
}

// DelSection del section by name
func (c *Ini) DelSection(name string) bool {
	// if is readonly
	if c.opts.Readonly {
		return false
	}

	if c.opts.IgnoreCase {
		name = strings.ToLower(name)
	}

	_, ok := c.data[name]
	if ok {
		delete(c.data, name)
	}

	return ok
}

/*************************************************************
 * helper methods
 *************************************************************/

// HasKey check
func (c *Ini) HasKey(key string) (ok bool) {
	_, ok = c.Get(key)
	return
}

// Del key
func (c *Ini) Del(key string) (ok bool) {
	// if is readonly
	if c.opts.Readonly {
		return
	}

	sep := c.opts.SectionSep
	key = formatKey(key, sep)
	if key == "" {
		return
	}

	if c.opts.IgnoreCase {
		key = strings.ToLower(key)
	}

	sec, key := c.splitSectionAndKey(key, sep)
	mp, ok := c.data[sec]
	if !ok {
		return
	}

	if _, ok = mp[key]; ok {
		delete(mp, key)
		c.data[sec] = mp
	}

	return
}

// Reset all data
func (c *Ini) Reset() {
	c.data = make(map[string]Section)
}

// Data get all data
func (c *Ini) Data() map[string]Section {
	return c.data
}

func (c *Ini) splitSectionAndKey(key, sep string) (string, string) {
	// default find from default Section
	name := c.opts.DefSection

	// get val by path. eg "log.dir"
	if strings.Contains(key, sep) {
		ss := strings.SplitN(key, sep, 2)
		name, key = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
	}

	return name, key
}

// format key
func formatKey(key, sep string) string {
	return strings.Trim(strings.TrimSpace(key), sep)
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
