package parser

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

// Encode golang data(map, struct) to INI string.
func Encode(v any) ([]byte, error) { return EncodeWithDefName(v) }

// EncodeWithDefName golang data(map, struct) to INI, can set default section name
func EncodeWithDefName(v any, defSection ...string) (out []byte, err error) {
	switch vd := v.(type) {
	case map[string]any: // from full mode
		return EncodeFull(vd, defSection...)
	case map[string]map[string]string: // from simple mode
		return EncodeSimple(vd, defSection...)
	default:
		err = errors.New("ini: invalid data to encode as INI")
	}
	return
}

// EncodeFull full mode data to INI, can set default section name
func EncodeFull(data map[string]any, defSection ...string) (out []byte, err error) {
	if len(data) == 0 {
		return
	}

	defSecName := ""
	if len(defSection) > 0 {
		defSecName = defSection[0]
	}

	// sort data
	counter := 0
	sections := make([]string, len(data))
	for section := range data {
		sections[counter] = section
		counter++
	}
	sort.Strings(sections)

	defBuf := &bytes.Buffer{}
	secBuf := &bytes.Buffer{}

	for _, key := range sections {
		item := data[key]
		switch tpData := item.(type) {
		case float32, float64, int, int32, int64, string, bool: // k-v of the default section
			_, _ = defBuf.WriteString(fmt.Sprintf("%s = %v\n", key, tpData))
		case []int:
		case []string: // array of the default section
			for _, v := range tpData {
				_, _ = defBuf.WriteString(fmt.Sprintf("%s[] = %v\n", key, v))
			}
		// case map[string]string: // is section
		case map[string]any: // is section
			if key != defSecName {
				secBuf.WriteString("[" + key + "]\n")
				buildSectionBuffer(tpData, secBuf)
			} else {
				buildSectionBuffer(tpData, defBuf)
			}
			secBuf.WriteByte('\n')
		}
	}

	defBuf.WriteByte('\n')
	defBuf.Write(secBuf.Bytes())
	out = defBuf.Bytes()
	secBuf = nil
	return
}

func buildSectionBuffer(data map[string]any, buf *bytes.Buffer) {
	for key, item := range data {
		switch tpData := item.(type) {
		case []int:
		case []string: // array of the default section
			for _, v := range tpData {
				_, _ = buf.WriteString(fmt.Sprintf("%s[] = %v\n", key, v))
			}
		default: // k-v of the section
			_, _ = buf.WriteString(fmt.Sprintf("%s = %v\n", key, tpData))
		}
	}
}

// EncodeSimple data to INI
func EncodeSimple(data map[string]map[string]string, defSection ...string) ([]byte, error) {
	return EncodeLite(data, defSection...)
}

// EncodeLite data to INI
func EncodeLite(data map[string]map[string]string, defSection ...string) (out []byte, err error) {
	if len(data) == 0 {
		return
	}

	buf := &bytes.Buffer{}
	counter := 0
	defSecName := ""
	sortedSections := make([]string, len(data))

	if len(defSection) > 0 {
		defSecName = defSection[0]
	}

	for section := range data {
		sortedSections[counter] = section
		counter++
	}

	sort.Strings(sortedSections)
	for _, section := range sortedSections {
		// don't add section title for DefSection
		if section != defSecName {
			_, _ = buf.WriteString("[" + section + "]\n")
		}

		counter = 0
		items := data[section]
		orderedStringKeys := make([]string, len(items))

		for key := range items {
			orderedStringKeys[counter] = key
			counter++
		}

		sort.Strings(orderedStringKeys)
		for _, key := range orderedStringKeys {
			_, _ = buf.WriteString(key + " = " + items[key] + "\n")
		}

		buf.WriteByte('\n')
	}

	out = buf.Bytes()
	return
}
