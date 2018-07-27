package ini

import (
	"github.com/gookit/ini/parser"
	"os"
	"regexp"
	"strings"
)

// parse env value, eg: "${SHELL}" ${NotExist|defValue}
var envRegex = regexp.MustCompile(`\$\{([\w-| ]+)}`)

// parse and load data
func (c *Ini) parse(data string) (err error) {
	if strings.TrimSpace(data) == "" {
		return
	}

	p := parser.SimpleParser()
	p.DefSection = c.opts.DefSection
	p.Collector = c.valueCollector
	p.IgnoreCase = c.opts.IgnoreCase

	return p.ParseString(data)
}

// collect value form parser
func (c *Ini) valueCollector(section, key, val string, isArr bool) {
	if c.opts.IgnoreCase {
		section = strings.ToLower(section)
		key = strings.ToLower(key)
	}

	// if opts.ParseEnv is true. will parse like: "${SHELL}"
	if c.opts.ParseEnv {
		val = c.parseEnvValue(val)
	}

	if sec, ok := c.data[section]; ok {
		sec[key] = val
		c.data[section] = sec
	} else {
		// create the section if it does not exist
		c.data[section] = Section{key: val}
	}
}

// parse Env Value
func (c *Ini) parseEnvValue(val string) string {
	if strings.Index(val, "${") == -1 {
		return val
	}

	// nodes like: ${VAR} -> [${VAR}]
	// val = "${GOPATH}/${APP_ENV | prod}/dir" -> [${GOPATH} ${APP_ENV | prod}]
	vars := envRegex.FindAllString(val, -1)
	if len(vars) == 0 {
		return val
	}

	var oldNew []string
	var name, def string
	for _, fVar := range vars {
		ss := strings.SplitN(fVar[2:len(fVar)-1], "|", 2)

		// has default ${NotExist|defValue}
		if len(ss) == 2 {
			name, def = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
		} else {
			def = fVar
			name = ss[0]
		}

		envVal := os.Getenv(name)
		if envVal == "" {
			envVal = def
		}

		oldNew = append(oldNew, fVar, envVal)
	}

	return strings.NewReplacer(oldNew...).Replace(val)
}

// parse var reference
func (c *Ini) parseVarReference(valStr string, sec Section) string {
	if strings.Index(valStr, c.opts.VarOpen) == -1 {
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

		if val, ok := sec[name]; ok {
			realVal = val
		} else if val, ok := c.Get(name); ok {
			realVal = val
		}

		oldNew = append(oldNew, fVar, realVal)
	}

	return strings.NewReplacer(oldNew...).Replace(valStr)
}
