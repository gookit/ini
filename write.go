package ini

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

// PrettyJSON translate to pretty JSON string
func (c *Ini) PrettyJSON() string {
	if len(c.data) == 0 {
		return ""
	}

	out, err := json.MarshalIndent(c.data, "", "    ")
	if err != nil {
		return ""
	}

	return string(out)
}

// WriteToFile write config data to a file
func (c *Ini) WriteToFile(file string) (n int64, err error) {
	// open file
	fd, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return
	}

	return c.WriteTo(fd)
}

// WriteTo out an INI File representing the current state to a writer.
func (c *Ini) WriteTo(out io.Writer) (n int64, err error) {
	n = 0
	counter := 0
	thisWrite := 0
	defSection := c.opts.DefSection
	orderedSections := make([]string, len(c.data))

	for section := range c.data {
		orderedSections[counter] = section
		counter++
	}

	sort.Strings(orderedSections)

	for _, section := range orderedSections {
		// don't add section title for DefSection
		if section != defSection {
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
