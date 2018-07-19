package ini

import (
	"github.com/gookit/ini/parser"
	"strings"
)

// parse and load data
func (ini *Ini) parse(data string) (err error) {
	if strings.TrimSpace(data) == "" {
		return
	}

	p := parser.SimpleParser()
	p.DefSection = DefSection
	p.Collector = ini.valueCollector
	p.IgnoreCase = ini.opts.IgnoreCase

	return p.ParseString(data)
}

// collect value form parser
func (ini *Ini) valueCollector(section, key, val string, isArr bool) {
	if ini.opts.IgnoreCase {
		section = strings.ToLower(section)
		key = strings.ToLower(key)
	}

	if sec, ok := ini.data[section]; ok {
		sec[key] = val
		ini.data[section] = sec
	} else {
		// create the section if it does not exist
		ini.data[section] = Section{key: val}
	}
}
