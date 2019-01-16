package ini

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

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

	key = c.formatKey(key)
	if key == "" {
		return
	}

	sep := c.opts.SectionSep
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

	return stringToArray(str, sep), ok
}

// StringMap get a section data map
func (c *Ini) StringMap(name string) (mp map[string]string, ok bool) {
	name = c.formatKey(name)
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
func (c *Ini) Section(name string) (Section, bool) {
	return c.StringMap(name)
}

/*************************************************************
 * config set
 *************************************************************/

// Set a value to the section by key.
// if section is empty, will set to default section
func (c *Ini) Set(key string, val interface{}, section ...string) (err error) {
	// if is readonly
	if c.opts.Readonly {
		return errSetInReadonly
	}

	c.ensureInit()
	c.lock.Lock()
	defer c.lock.Unlock()

	key = c.formatKey(key)
	if key == "" {
		return errEmptyKey
	}

	// section name
	name := c.opts.DefSection
	if len(section) > 0 {
		name = section[0]
	}

	strVal, isString := val.(string)
	if !isString {
		strVal = fmt.Sprint(val)
	}

	// allow section name is empty string ""
	name = c.formatKey(name)
	sec, ok := c.data[name]
	if ok {
		sec[key] = strVal
	} else {
		sec = Section{key: strVal}
	}

	c.data[name] = sec
	return
}

// SetInt set a int by key
func (c *Ini) SetInt(key string, value int, section ...string) error {
	return c.Set(key, fmt.Sprintf("%d", value), section...)
}

// SetBool set a bool by key
func (c *Ini) SetBool(key string, value bool, section ...string) error {
	valStr := "false"
	if value {
		valStr = "true"
	}

	return c.Set(key, valStr, section...)
}

// SetString set a string by key
func (c *Ini) SetString(key, val string, section ...string) error {
	return c.Set(key, val, section...)
}

/*************************************************************
 * config dump
 *************************************************************/

// PrettyJSON translate to pretty JSON string
func (c *Ini) PrettyJSON() string {
	if len(c.data) == 0 {
		return ""
	}

	out, _ := json.MarshalIndent(c.data, "", "    ")
	return string(out)
}

// WriteToFile write config data to a file
func (c *Ini) WriteToFile(file string) (int64, error) {
	// open file
	fd, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return 0, err
	}

	return c.WriteTo(fd)
}

// WriteTo out an INI File representing the current state to a writer.
func (c *Ini) WriteTo(out io.Writer) (n int64, err error) {
	n = 0
	counter := 0
	thisWrite := 0
	// section
	defaultSection := c.opts.DefSection
	orderedSections := make([]string, len(c.data))

	for section := range c.data {
		orderedSections[counter] = section
		counter++
	}

	sort.Strings(orderedSections)

	for _, section := range orderedSections {
		// don't add section title for DefSection
		if section != defaultSection {
			thisWrite, err = fmt.Fprintln(out, "["+section+"]")
			n += int64(thisWrite)
			if err != nil {
				return
			}
		}

		items := c.data[section]
		orderedStringKeys := make([]string, len(items))
		counter = 0
		for key := range items {
			orderedStringKeys[counter] = key
			counter++
		}

		sort.Strings(orderedStringKeys)
		for _, key := range orderedStringKeys {
			thisWrite, err = fmt.Fprintln(out, key, "=", items[key])
			n += int64(thisWrite)
			if err != nil {
				return
			}
		}

		thisWrite, err = fmt.Fprintln(out)
		n += int64(thisWrite)
		if err != nil {
			return
		}
	}
	return
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

	name = c.formatKey(name)

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
	name = c.formatKey(name)
	_, ok := c.data[name]
	return ok
}

// DelSection del section by name
func (c *Ini) DelSection(name string) (ok bool) {
	// if is readonly
	if c.opts.Readonly {
		return
	}

	name = c.formatKey(name)
	if _, ok = c.data[name]; ok {
		delete(c.data, name)
	}
	return
}
