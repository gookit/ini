/*
Ini parse by golang. ini config file/data manage

Source code and other details for the project are available at GitHub:

	https://github.com/gookit/ini

INI parser is: https://github.com/gookit/ini/parser

 */
package ini

import (
	"os"
	"io/ioutil"
)

const (
	SepSection = "."
	DefSection = "__default"
)

// section in ini data
type Section map[string]string

// Options
type Options struct {
	Readonly   bool
	ParseEnv   bool
	IgnoreCase bool
}

// Ini data manager
type Ini struct {
	data map[string]Section
	opts *Options

	inited bool
}

// DefOptions
var DefOptions = &Options{ParseEnv: true}

// New
func New() *Ini {
	return &Ini{
		data: make(map[string]Section),
		opts: DefOptions,
	}
}

// NewWithOptions
// usage:
// ini.NewWithOptions(ini.ParseEnv, ini.Readonly)
func NewWithOptions(opts ...func(*Options)) *Ini {
	ini := &Ini{
		data: make(map[string]Section),
		opts: &Options{},
	}

	// apply options
	for _, opt := range opts {
		opt(ini.opts)
	}

	return ini
}

/*************************************************************
 * quick use
 *************************************************************/


// LoadFiles
func LoadFiles(files ...string) (ini *Ini, err error) {
	ini = New()
	err = ini.LoadFiles(files...)

	return
}

// LoadExists
func LoadExists(files ...string) (ini *Ini, err error) {
	ini = New()
	err = ini.LoadExists(files...)

	return
}

// LoadStrings
func LoadStrings(strings ...string) (ini *Ini, err error) {
	ini = New()
	err = ini.LoadStrings(strings...)

	return
}

/*************************************************************
 * options func
 *************************************************************/

// Readonly
// usage:
// ini.NewWithOptions(ini.Readonly)
func Readonly(opts *Options) {
	opts.Readonly = true
}

// ParseEnv
// usage:
// ini.NewWithOptions(ini.ParseEnv)
func ParseEnv(opts *Options) {
	opts.ParseEnv = true
}

// IgnoreCase
func IgnoreCase(opts *Options) {
	opts.IgnoreCase = true
}

// Options
func (ini *Ini) Options() *Options {
	return ini.opts
}

// WithOptions
func (ini *Ini) WithOptions(opts ...func(*Options)) {
	if ini.inited {
		panic("ini: Cannot set options after initialization is complete")
	}

	// apply options
	for _, opt := range opts {
		opt(ini.opts)
	}
}

/*************************************************************
 * data load
 *************************************************************/

// LoadFiles
func (ini *Ini) LoadFiles(files ...string) (err error) {
	ini.ensureInit()

	for _, file := range files {
		err = ini.loadFile(file, false)
		if err != nil {
			return
		}
	}

	return
}

// LoadExists
func (ini *Ini) LoadExists(files ...string) (err error) {
	ini.ensureInit()

	for _, file := range files {
		err = ini.loadFile(file, true)
		if err != nil {
			return
		}
	}

	return
}

// LoadStrings
func (ini *Ini) LoadStrings(strings ...string) (err error) {
	ini.ensureInit()

	for _, str := range strings {
		err = ini.parse(str)
		if err != nil {
			return
		}
	}

	return
}

// LoadData
func (ini *Ini) LoadData(data map[string]Section) {
	ini.ensureInit()
	if len(ini.data) == 0 {
		ini.data = data
	}

	// append or override setting data
	for name, sec := range data {
		ini.SetSection(name, sec)
	}
}

func (ini *Ini) ensureInit() {
	if ini.data == nil {
		ini.data = make(map[string]Section)
	}

	if !ini.inited {
		ini.inited = true
	}
}

func (ini *Ini) loadFile(file string, loadExist bool) (err error) {
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
		err = ini.parse(string(bts))
		if err != nil {
			return
		}
	}

	return
}
