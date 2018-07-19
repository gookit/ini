package ini

import (
	"github.com/gookit/ini/parser"
)

// parse and load data
func (ini *Ini) parse(data string) (err error) {
	p := parser.SimpleParser()

	if ini.opts.IgnoreCase {
		p.IgnoreCase = true
	}

	p.DefSection = DefSection
	p.Collector = ini.valueCollector

	err = p.ParseString(data)

	return
}

func (ini *Ini) valueCollector(section, key, val string, isArr bool) {
	if sec, ok := ini.data[section]; ok {
		sec[key] = val
		ini.data[section] = sec
	} else {
		// create the section if it does not exist
		ini.data[section] = Section{key: val}
	}
}
