/*
Package ini is a ini config file/data manage implement

Source code and other details for the project are available at GitHub:

	https://github.com/gookit/ini

INI parser is: https://github.com/gookit/ini/parser

*/
package ini

import (
	"io/ioutil"
	"os"
	"regexp"
	"sync"
)

// some default constants
const (
	SepSection = "."
	DefSection = "__default"
	// EnvValue   = "${NAME}"
)

// Section in INI config
type Section map[string]string

// Options for config
type Options struct {
	// set to read-only mode
	Readonly bool
	// parse ENV var name. default True
	ParseEnv bool
	// parse variable reference. %(varName)s
	ParseVar bool

	// var left open char. default "%("
	VarOpen string
	// var right open char. default ")s"
	VarClose string

	// ignore the case of the key. default False
	IgnoreCase bool
	// default section name. default "__default"
	DefSection string
	// sep char for split key path. default ".", use like "section.subKey"
	SectionSep string
}

// Ini config data manager
type Ini struct {
	data map[string]Section
	opts *Options
	lock sync.RWMutex

	// when has data loaded, will change to true
	initialized bool
	varRegex    *regexp.Regexp
}

/*************************************************************
 * create config instance
 *************************************************************/

// New a instance
func New() *Ini {
	return &Ini{
		data: make(map[string]Section),
		opts: newDefaultOptions(),
	}
}

// NewWithOptions new a instance and with some options
// usage:
// ini.NewWithOptions(ini.ParseEnv, ini.Readonly)
func NewWithOptions(opts ...func(*Options)) *Ini {
	c := New()
	// apply options
	c.WithOptions(opts...)

	return c
}

/*************************************************************
 * quick use
 *************************************************************/

// LoadFiles load data from files
func LoadFiles(files ...string) (c *Ini, err error) {
	c = New()
	err = c.LoadFiles(files...)

	return
}

// LoadExists load files, will ignore not exists
func LoadExists(files ...string) (c *Ini, err error) {
	c = New()
	err = c.LoadExists(files...)

	return
}

// LoadStrings load data from strings
func LoadStrings(strings ...string) (c *Ini, err error) {
	c = New()
	err = c.LoadStrings(strings...)

	return
}

/*************************************************************
 * options func
 *************************************************************/

// newDefaultOptions create a new default Options
// Notice:
// Cannot use package var instead it. That will allow multiple instances to use the same Options
func newDefaultOptions() *Options {
	return &Options{
		ParseEnv: true,

		VarOpen:  "%(",
		VarClose: ")s",

		DefSection: DefSection,
		SectionSep: SepSection,
	}
}

// Readonly setting
// usage:
// ini.NewWithOptions(ini.Readonly)
func Readonly(opts *Options) {
	opts.Readonly = true
}

// ParseVar on get value
// usage:
// ini.WithOptions(ini.ParseVar)
func ParseVar(opts *Options) {
	opts.ParseVar = true
}

// ParseEnv on get value
// usage:
// ini.WithOptions(ini.ParseEnv)
func ParseEnv(opts *Options) {
	opts.ParseEnv = true
}

// IgnoreCase for get/set value by key
func IgnoreCase(opts *Options) {
	opts.IgnoreCase = true
}

// Options get
func (c *Ini) Options() *Options {
	return c.opts
}

// WithOptions apply some options
func (c *Ini) WithOptions(opts ...func(*Options)) {
	if c.initialized {
		panic("ini: Cannot set options after initialization is complete")
	}

	// apply options
	for _, opt := range opts {
		opt(c.opts)
	}
}

// DefSection get default section name
func (c *Ini) DefSection() string {
	return c.opts.DefSection
}

/*************************************************************
 * data load
 *************************************************************/

// LoadFiles load data from files
func (c *Ini) LoadFiles(files ...string) (err error) {
	c.ensureInit()

	for _, file := range files {
		err = c.loadFile(file, false)
		if err != nil {
			return
		}
	}

	if !c.initialized {
		c.initialized = true
	}
	return
}

// LoadExists load files, will ignore not exists
func (c *Ini) LoadExists(files ...string) (err error) {
	c.ensureInit()

	for _, file := range files {
		err = c.loadFile(file, true)
		if err != nil {
			return
		}
	}

	if !c.initialized {
		c.initialized = true
	}
	return
}

// LoadStrings load data from strings
func (c *Ini) LoadStrings(strings ...string) (err error) {
	c.ensureInit()

	for _, str := range strings {
		err = c.parse(str)
		if err != nil {
			return
		}
	}

	if !c.initialized {
		c.initialized = true
	}

	return
}

// LoadData load data map
func (c *Ini) LoadData(data map[string]Section) (err error) {
	c.ensureInit()
	if len(c.data) == 0 {
		c.data = data
	}

	// append or override setting data
	for name, sec := range data {
		err = c.SetSection(name, sec)
		if err != nil {
			return
		}
	}

	if !c.initialized {
		c.initialized = true
	}

	return
}

func (c *Ini) ensureInit() {
	if c.initialized {
		return
	}

	if c.data == nil {
		c.data = make(map[string]Section)
	}

	if c.opts == nil {
		c.opts = newDefaultOptions()
	}

	// build var regex. default is `%\(([\w-:]+)\)s`
	if c.opts.ParseVar && c.varRegex == nil {
		// regexStr := `%\([\w-:]+\)s`
		l := regexp.QuoteMeta(c.opts.VarOpen)
		r := regexp.QuoteMeta(c.opts.VarClose)

		// build like: `%\(([\w-:]+)\)s`
		regStr := l + `([\w-` + c.opts.SectionSep + `]+)` + r
		c.varRegex = regexp.MustCompile(regStr)
	}
}

func (c *Ini) loadFile(file string, loadExist bool) (err error) {
	// open file
	fd, err := os.Open(file)
	if err != nil {
		// skip not exist file
		if os.IsNotExist(err) && loadExist {
			return nil
		}

		return
	}
	defer fd.Close()

	// read file content
	bts, err := ioutil.ReadAll(fd)
	if err == nil {
		err = c.parse(string(bts))
		if err != nil {
			return
		}
	}

	return
}
