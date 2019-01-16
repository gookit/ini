/*
Package ini is a ini config file/data manage implement

Source code and other details for the project are available at GitHub:

	https://github.com/gookit/ini

INI parser is: https://github.com/gookit/ini/parser

*/
package ini

import (
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
)

// some default constants
const (
	SepSection = "."
	DefSection = "__default"
)

var (
	errEmptyKey      = errors.New("ini: key name cannot be empty")
	errSetInReadonly = errors.New("ini: config manager instance in 'readonly' mode")
	// default instance
	dc = New()
)

// Section in INI config
type Section map[string]string

// Options for config
type Options struct {
	// set to read-only mode. default False
	Readonly bool
	// parse ENV var name. default True
	ParseEnv bool
	// parse variable reference "%(varName)s". default False
	ParseVar bool

	// var left open char. default "%("
	VarOpen string
	// var right close char. default ")s"
	VarClose string

	// ignore key name case. default False
	IgnoreCase bool
	// default section name. default "__default", it's allow empty string.
	DefSection string
	// sep char for split key path. default ".", use like "section.subKey"
	SectionSep string
}

// Ini config data manager
type Ini struct {
	opts *Options
	lock sync.RWMutex
	data map[string]Section
	// regex for match user var
	varRegex *regexp.Regexp
}

/*************************************************************
 * config instance
 *************************************************************/

// New a config instance, with default options
func New() *Ini {
	return &Ini{
		data: make(map[string]Section),
		opts: newDefaultOptions(),
	}
}

// NewWithOptions new a instance and with some options
// Usage:
// ini.NewWithOptions(ini.ParseEnv, ini.Readonly)
func NewWithOptions(opts ...func(*Options)) *Ini {
	c := New()
	// apply options
	c.WithOptions(opts...)
	return c
}

// Default config instance
func Default() *Ini {
	return dc
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
// Usage:
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
	if !c.IsEmpty() {
		panic("ini: Cannot set options after data has been load")
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
func LoadFiles(files ...string) error { return dc.LoadFiles(files...) }

// LoadFiles load data from files
func (c *Ini) LoadFiles(files ...string) (err error) {
	c.ensureInit()

	for _, file := range files {
		err = c.loadFile(file, false)
		if err != nil {
			return
		}
	}
	return
}

// LoadExists load files, will ignore not exists
func LoadExists(files ...string) error { return dc.LoadExists(files...) }

// LoadExists load files, will ignore not exists
func (c *Ini) LoadExists(files ...string) (err error) {
	c.ensureInit()

	for _, file := range files {
		err = c.loadFile(file, true)
		if err != nil {
			return
		}
	}
	return
}

// LoadStrings load data from strings
func LoadStrings(strings ...string) error { return dc.LoadStrings(strings...) }

// LoadStrings load data from strings
func (c *Ini) LoadStrings(strings ...string) (err error) {
	c.ensureInit()

	for _, str := range strings {
		err = c.parse(str)
		if err != nil {
			return
		}
	}
	return
}

// LoadData load data map
func (c *Ini) LoadData(data map[string]Section) (err error) {
	c.ensureInit()

	if len(c.data) == 0 {
		c.data = data
		return
	}

	// append or override setting data
	for name, sec := range data {
		err = c.SetSection(name, sec)
		if err != nil {
			return
		}
	}
	return
}

func (c *Ini) ensureInit() {
	if !c.IsEmpty() {
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

// HasKey check
func (c *Ini) HasKey(key string) (ok bool) {
	_, ok = c.Get(key)
	return
}

// Delete value by key
func (c *Ini) Delete(key string) (ok bool) {
	if c.opts.Readonly {
		return
	}

	key = c.formatKey(key)
	if key == "" {
		return
	}

	sep := c.opts.SectionSep
	sec, key := c.splitSectionAndKey(key, sep)
	mp, ok := c.data[sec]
	if !ok {
		return
	}

	// key in a section
	if _, ok = mp[key]; ok {
		delete(mp, key)
		c.data[sec] = mp
	}
	return
}

// Reset all data for the default
func Reset() { dc.Reset() }

// Reset all data
func (c *Ini) Reset() {
	c.data = make(map[string]Section)
}

// IsEmpty config data is empty
func IsEmpty() bool { return len(dc.data) == 0 }

// IsEmpty config data is empty
func (c *Ini) IsEmpty() bool {
	return len(c.data) == 0
}

// Data get all data from default instance
func Data() map[string]Section { return dc.data }

// Data get all data
func (c *Ini) Data() map[string]Section {
	return c.data
}

/*************************************************************
 * helper methods
 *************************************************************/

func (c *Ini) splitSectionAndKey(key, sep string) (string, string) {
	// default find from default Section
	name := c.opts.DefSection

	// get val by path. eg "log.dir"
	if strings.Contains(key, sep) {
		ss := strings.SplitN(key, sep, 2)
		name, key = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
	}

	return name, key
}

// format key by some options
func (c *Ini) formatKey(key string) string {
	sep := c.opts.SectionSep
	key = strings.Trim(strings.TrimSpace(key), sep)

	if c.opts.IgnoreCase {
		key = strings.ToLower(key)
	}

	return key
}

// simple merge two string map
func mergeStringMap(src, dst map[string]string, ignoreCase bool) map[string]string {
	for k, v := range src {
		if ignoreCase {
			k = strings.ToLower(k)
		}

		dst[k] = v
	}
	return dst
}

func mapKeyToLower(src map[string]string) map[string]string {
	newMp := make(map[string]string)

	for k, v := range src {
		k = strings.ToLower(k)
		newMp[k] = v
	}
	return newMp
}

func stringToArray(str, sep string) (arr []string) {
	str = strings.TrimSpace(str)
	ss := strings.Split(str, sep)

	for _, val := range ss {
		if val = strings.TrimSpace(val); val != "" {
			arr = append(arr, val)
		}
	}
	return arr
}
