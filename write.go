package ini

import (
	"io"
	"sort"
	"fmt"
	"os"
	"encoding/json"
	"bytes"
)

// String
func (ini *Ini) String() string {
	buf := &bytes.Buffer{}

	_, err := ini.WriteTo(buf)
	if err != nil {
		return ""
	}

	return buf.String()
}

// WriteToFile get pretty Json string
func (ini *Ini) PrettyJson() string{
	out, err := json.MarshalIndent(ini.data, "", "    ")
	if err != nil {
		return ""
	}

	return string(out)
}

// WriteToFile
func (ini *Ini) WriteToFile(file string) (n int64, err error) {
	// open file
	fd, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return
	}

	return ini.WriteTo(fd)
}

// Write out an INI File representing the current state to a writer.
func (ini *Ini) WriteTo(out io.Writer) (n int64, err error) {
	n = 0
	counter := 0
	thisWrite := 0
	orderedSections := make([]string, len(ini.data))

	for section := range ini.data {
		orderedSections[counter] = section
		counter++
	}

	sort.Strings(orderedSections)

	for _, section := range orderedSections {
		if section != DefSection {
			thisWrite, err = fmt.Fprintln(out, "["+section+"]")
			n += int64(thisWrite)
			if err != nil {
				return
			}
		}

		items := ini.data[section]
		orderedStringKeys := make([]string, len(items))
		counter = 0
		for key, _ := range items {
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

		// orderedArrayKeys := make([]string, len(options.arrayValues))
		// counter = 0
		// for key, _ := range options.arrayValues {
		// 	orderedArrayKeys[counter] = key
		// 	counter++
		// }
		// sort.Strings(orderedArrayKeys)
		// for _, key := range orderedArrayKeys {
		// 	for _, value := range options.arrayValues[key] {
		// 		thisWrite, err = fmt.Fprintln(out, key, "[]=", value)
		// 		n += int64(thisWrite)
		// 		if (err) != nil {
		// 			return
		// 		}
		// 	}
		// }
		thisWrite, err = fmt.Fprintln(out)
		n += int64(thisWrite)
		if err != nil {
			return
		}
	}
	return
}
