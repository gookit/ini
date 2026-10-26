package ini

import (
	"strings"

	"github.com/gookit/goutil/envutil"
	"github.com/gookit/ini/v2/parser"
)

// parse and load ini string
func (c *Ini) parse(str string) (err error) {
	if strings.TrimSpace(str) == "" {
		return
	}

	p := parser.NewLite()
	p.Collector = c.valueCollector
	p.IgnoreCase = c.opts.IgnoreCase
	p.DefSection = c.opts.DefSection

	err = p.ParseString(str)
	c.comments = p.Comments()
	p.Reset()
	return err
}

// collect value form parser
func (c *Ini) valueCollector(section, key, val string, _ bool) {
	if c.opts.IgnoreCase {
		key = strings.ToLower(key)
		section = strings.ToLower(section)
	}

	// backup value on contains var, use for export
	if strings.ContainsRune(val, '$') {
		c.rawBak[rawValueKey{section, key}] = val
		// if ParseEnv is true. will parse like: "${SHELL}".
		if c.opts.ParseEnv {
			val = envutil.ParseValue(val)
		}
	} else {
		delete(c.rawBak, rawValueKey{section, key})
	}

	if c.opts.ReplaceNl {
		val = strings.ReplaceAll(val, `\n`, "\n")
	}

	if sec, ok := c.data[section]; ok {
		sec[key] = val
		c.data[section] = sec
	} else {
		// create the section if it does not exist
		c.data[section] = Section{key: val}
	}
}

// maxVarRefDepth limits how deep the nested variable references are resolved, it also
// stops a circular reference(eg: "a = %(b)s" / "b = %(a)s") from recursing forever.
// The 10 levels limit is the same as the Python configparser.
const maxVarRefDepth = 10

// parseVarReference parse the var reference in the value. eg: "%(name)s", "%(section.key)s"
func (c *Ini) parseVarReference(key, valStr string, sec Section) string {
	return c.parseVarRefDepth(key, valStr, sec, 0)
}

// parseVarRefDepth parse the var reference in the value, with the current recursion depth.
func (c *Ini) parseVarRefDepth(key, valStr string, sec Section, depth int) string {
	if c.opts.VarOpen != "" && !strings.Contains(valStr, c.opts.VarOpen) {
		return valStr
	}

	// http://%(host)s:%(port)s/Portal
	// %(section.key)s key in the section
	vars := c.varRegex.FindAllString(valStr, -1)
	if len(vars) == 0 {
		return valStr
	}

	varOLen, varCLen := len(c.opts.VarOpen), len(c.opts.VarClose)

	var name string
	var oldNew []string
	for _, fVar := range vars {
		realVal := fVar
		name = fVar[varOLen : len(fVar)-varCLen]

		// first, find from current section
		if val, ok := sec[name]; ok && key != name {
			realVal = val
			if depth < maxVarRefDepth {
				realVal = c.parseVarRefDepth(name, val, sec, depth+1)
			}
		} else if val, ok = c.getValue(name); ok {
			realVal = val
			if depth < maxVarRefDepth {
				secName, refKey := c.splitSectionAndKey(c.formatKey(name))
				realVal = c.parseVarRefDepth(refKey, val, c.data[secName], depth+1)
			}
		}

		oldNew = append(oldNew, fVar, realVal)
	}

	return strings.NewReplacer(oldNew...).Replace(valStr)
}
