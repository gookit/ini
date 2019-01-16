package ini

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

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
