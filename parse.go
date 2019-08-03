package ini

import (
	"strings"

	"github.com/gookit/ini/v2/parser"
)

// parse and load data
func (c *Ini) parse(data string) (err error) {
	if strings.TrimSpace(data) == "" {
		return
	}

	p := parser.NewSimpled()
	p.Collector = c.valueCollector
	p.DefSection = c.opts.DefSection
	p.IgnoreCase = c.opts.IgnoreCase

	return p.ParseString(data)
}

// collect value form parser
func (c *Ini) valueCollector(section, key, val string, isSlice bool) {
	if c.opts.IgnoreCase {
		section = strings.ToLower(section)
		key = strings.ToLower(key)
	}

	// if opts.ParseEnv is true. will parse like: "${SHELL}"
	// parse on there, will export data error.
	// if c.opts.ParseEnv {
	// 	val = c.parseEnvValue(val)
	// }

	if sec, ok := c.data[section]; ok {
		sec[key] = val
		c.data[section] = sec
	} else {
		// create the section if it does not exist
		c.data[section] = Section{key: val}
	}
}

// parse var reference
func (c *Ini) parseVarReference(key, valStr string, sec Section) string {
	if c.opts.VarOpen != "" && strings.Index(valStr, c.opts.VarOpen) == -1 {
		return valStr
	}

	// http://%(host)s:%(port)s/Portal
	// %(section:key)s key in the section
	vars := c.varRegex.FindAllString(valStr, -1)
	if len(vars) == 0 {
		return valStr
	}

	var name string
	var oldNew []string
	for _, fVar := range vars {
		realVal := fVar
		name = fVar[2 : len(fVar)-2]

		if val, ok := sec[name]; ok && key != name {
			realVal = val
		} else if val, ok := c.GetValue(name); ok {
			realVal = val
		}

		oldNew = append(oldNew, fVar, realVal)
	}

	return strings.NewReplacer(oldNew...).Replace(valStr)
}
