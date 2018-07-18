package ini

import (
	"os"
	"io/ioutil"
	"io"
	"bufio"
)

const (
	DefSection = "__default"
)

// section in ini data
type Section map[string]string

// type ArrSection map[string][]string

// Options
type Options struct {
	ParseEnv   bool
	IgnoreCase bool
}

// Ini data manager
type Ini struct {
	data   map[string]Section
	opts   Options

	inited bool
}

// DefOptions
var DefOptions = Options{ParseEnv: true}

// New
func New() *Ini {
	return &Ini{
		data: make(map[string]Section),
		opts: DefOptions,
	}
}

// NewWithOptions
func NewWithOptions(opts Options) *Ini {
	return &Ini{
		data: make(map[string]Section),
		opts: opts,
	}
}

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

// SetOptions
func (ini *Ini) SetOptions(opts Options) {
	if ini.inited {
		panic("cannot setting options after inited")
	}

	ini.opts = opts
}

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
		p := newParser()
		err = p.parseString(str)
		if err != nil {
			return
		}

		// load and merge data
		ini.MergeData(p.mapping)
	}

	return
}

// ReadFrom Loads INI data from a reader and stores the data in the manager.
func (ini *Ini) ReadFrom(in io.Reader) (n int64, err error) {
	n = 0
	scanner := bufio.NewScanner(in)

	p := newParser()
	n, err = p.parse(scanner)

	return
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
		p := newParser()
		err = p.parseBytes(bts)
		if err != nil {
			return
		}

		// load and merge data
		ini.MergeData(p.mapping)
	}

	return
}
