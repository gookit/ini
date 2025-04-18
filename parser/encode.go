package parser

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/gookit/goutil/structs"
	"github.com/gookit/goutil/timex"
)

type EncodeOptions struct {
	// DefSection name
	DefSection string
	// AddExportDate add export date to head of file. default: true
	AddExportDate bool
	// Comments comments map, key is `section +"_"+ key`, value is comment.
	Comments map[string]string
	// RawValueMap raw value map, key is `section +"_"+ key`, value is raw value.
	//
	// TIP: if you want to set raw value to INI file, you can use this option. see `rawBak` in ini.Ini
	RawValueMap map[string]string
}

func newEncodeOptions(defSection []string) *EncodeOptions {
	opts := &EncodeOptions{AddExportDate: true}
	if len(defSection) > 0 {
		opts.DefSection = defSection[0]
	}
	return opts
}

// EncodeWith golang data(map, struct) to INI string, can with options.
func EncodeWith(v any, opts *EncodeOptions) ([]byte, error) {
	if opts == nil {
		opts = &EncodeOptions{AddExportDate: true}
	}

	switch vd := v.(type) {
	case map[string]any: // from full mode
		return encodeFull(vd, opts)
	case map[string]map[string]string: // from lite mode
		return encodeLite(vd, opts)
	default:
		if vd != nil {
			// as struct data, use structs.ToMap convert
			anyMap, err := structs.StructToMap(vd)
			if err != nil {
				return nil, err
			}
			return encodeFull(anyMap, opts)
		}
		return nil, errors.New("ini: invalid data to encode as INI")
	}
}

// Encode golang data(map, struct) to INI string.
func Encode(v any) ([]byte, error) { return EncodeWithDefName(v) }

// EncodeWithDefName golang data(map, struct) to INI, can set default section name
func EncodeWithDefName(v any, defSection ...string) (out []byte, err error) {
	return EncodeWith(v, newEncodeOptions(defSection))
}

// EncodeFull full mode data to INI, can set default section name
func EncodeFull(data map[string]any, defSection ...string) (out []byte, err error) {
	return encodeFull(data, newEncodeOptions(defSection))
}

// EncodeSimple data to INI
func EncodeSimple(data map[string]map[string]string, defSection ...string) ([]byte, error) {
	return encodeLite(data, newEncodeOptions(defSection))
}

// EncodeLite data to INI
func EncodeLite(data map[string]map[string]string, defSection ...string) (out []byte, err error) {
	return encodeLite(data, newEncodeOptions(defSection))
}

// EncodeFull full mode data to INI, can set default section name
func encodeFull(data map[string]any, opts *EncodeOptions) (out []byte, err error) {
	ln := len(data)
	if ln == 0 {
		return
	}

	defSecName := opts.DefSection
	sortedGroups := make([]string, 0, ln)
	for section := range data {
		sortedGroups = append(sortedGroups, section)
	}

	buf := &bytes.Buffer{}
	buf.Grow(ln * 4)
	if opts.AddExportDate {
		buf.WriteString("; exported at " + timex.Now().Datetime() + "\n\n")
	}

	sort.Strings(sortedGroups)
	maxLn := len(sortedGroups) - 1
	secBuf := &bytes.Buffer{}

	for idx, section := range sortedGroups {
		item := data[section]
		switch tpData := item.(type) {
		case []int:
		case []string: // array of the default section
			for _, v := range tpData {
				buf.WriteString(fmt.Sprintf("%s[] = %v\n", section, v))
			}
		// case map[string]string: // is section
		case map[string]any: // is section
			if section != defSecName {
				secBuf.WriteString("[" + section + "]\n")
				writeAnyMap(secBuf, tpData)
			} else {
				writeAnyMap(buf, tpData)
			}

			if idx < maxLn {
				secBuf.WriteByte('\n')
			}
		default: // k-v of the default section
			buf.WriteString(fmt.Sprintf("%s = %v\n", section, tpData))
		}
	}

	buf.WriteByte('\n')
	buf.Write(secBuf.Bytes())
	out = buf.Bytes()
	secBuf = nil
	return
}

func writeAnyMap(buf *bytes.Buffer, data map[string]any) {
	for key, item := range data {
		switch tpData := item.(type) {
		case []int:
		case []string: // array of the default section
			for _, v := range tpData {
				buf.WriteString(key + "[] = ")
				buf.WriteString(fmt.Sprint(v))
				buf.WriteByte('\n')
			}
		default: // k-v of the section
			buf.WriteString(key + " = ")
			buf.WriteString(fmt.Sprint(tpData))
			buf.WriteByte('\n')
		}
	}
}

func encodeLite(data map[string]map[string]string, opts *EncodeOptions) (out []byte, err error) {
	ln := len(data)
	if ln == 0 {
		return
	}

	defSecName := opts.DefSection
	sortedGroups := make([]string, 0, ln)
	for section := range data {
		// don't add section title for default section
		if section != defSecName {
			sortedGroups = append(sortedGroups, section)
		}
	}

	buf := &bytes.Buffer{}
	buf.Grow(ln * 4)
	if opts.AddExportDate {
		buf.WriteString("; exported at " + timex.Now().Datetime() + "\n\n")
	}

	// first, write default section values
	if defSec, ok := data[defSecName]; ok {
		writeStrMap(buf, defSec, defSecName, opts)
		buf.WriteByte('\n')
	}

	sort.Strings(sortedGroups)
	maxLn := len(sortedGroups) - 1
	for idx, section := range sortedGroups {
		// comments for section
		if s, ok := opts.Comments[section]; ok {
			buf.WriteString(s + "\n")
		}

		buf.WriteString("[" + section + "]\n")
		writeStrMap(buf, data[section], section, opts)

		if idx < maxLn {
			buf.WriteByte('\n')
		}
	}

	out = buf.Bytes()
	return
}

func writeStrMap(buf *bytes.Buffer, strMap map[string]string, section string, opts *EncodeOptions) {
	sortedKeys := make([]string, 0, len(strMap))
	for key := range strMap {
		sortedKeys = append(sortedKeys, key)
	}

	sort.Strings(sortedKeys)
	for _, key := range sortedKeys {
		value := strMap[key]
		keyPath := section + "_" + key
		// add comments
		if s, ok := opts.Comments[keyPath]; ok {
			buf.WriteString(s + "\n")
		}

		if val1, ok := opts.RawValueMap[keyPath]; ok {
			value = val1
		}
		buf.WriteString(key + " = " + value + "\n")
	}
}
