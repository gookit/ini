// Package dotenv provide load .env data to os ENV
package dotenv

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gookit/ini/v2/parser"
)

var (
	// UpperEnvKey change key to upper on set ENV
	UpperEnvKey = true

	// DefaultName default file name
	DefaultName = ".env"

	// OnlyLoadExists only load on file exists
	OnlyLoadExists bool

	// save original Env data
	// originalEnv []string

	// cache all loaded ENV data
	loadedData = map[string]string{}
)

// LoadedData get all loaded data by dontenv
func LoadedData() map[string]string {
	return loadedData
}

// Reset clear the previously set ENV value
func Reset() { ClearLoaded() }

// ClearLoaded clear the previously set ENV value
func ClearLoaded() {
	for key := range loadedData {
		_ = os.Unsetenv(key)
	}

	// reset
	loadedData = map[string]string{}
}

// DontUpperEnvKey don't change key to upper on set ENV
func DontUpperEnvKey() {
	UpperEnvKey = false
}

// Load parse .env file data to os ENV.
//
// Usage:
//
//	dotenv.Load("./", ".env")
func Load(dir string, filenames ...string) (err error) {
	if len(filenames) == 0 {
		filenames = []string{DefaultName}
	}

	for _, filename := range filenames {
		file := filepath.Join(dir, filename)
		if err = loadFile(file); err != nil {
			break
		}
	}
	return
}

// LoadExists only load on file exists
func LoadExists(dir string, filenames ...string) error {
	oldVal := OnlyLoadExists

	OnlyLoadExists = true
	err := Load(dir, filenames...)
	OnlyLoadExists = oldVal

	return err
}

// LoadFiles load ENV from given file
func LoadFiles(filePaths ...string) (err error) {
	for _, filePath := range filePaths {
		if err = loadFile(filePath); err != nil {
			break
		}
	}
	return
}

// LoadExistFiles load ENV from given files, only load exists
func LoadExistFiles(filePaths ...string) error {
	oldVal := OnlyLoadExists
	defer func() {
		OnlyLoadExists = oldVal
	}()

	OnlyLoadExists = true
	return LoadFiles(filePaths...)
}

// LoadFromMap load data from given string map
func LoadFromMap(kv map[string]string) (err error) {
	for key, val := range kv {
		if UpperEnvKey {
			key = strings.ToUpper(key)
		}

		err = os.Setenv(key, val)
		if err != nil {
			break
		}

		// cache it
		loadedData[key] = val
	}
	return
}

// Get get os ENV value by name
func Get(name string, defVal ...string) (val string) {
	if val, ok := getVal(name); ok {
		return val
	}

	if len(defVal) > 0 {
		val = defVal[0]
	}
	return
}

// Bool get a bool value by key
func Bool(name string, defVal ...bool) (val bool) {
	if str, ok := getVal(name); ok {
		val, err := strconv.ParseBool(str)
		if err == nil {
			return val
		}
	}

	if len(defVal) > 0 {
		val = defVal[0]
	}
	return
}

// Int get a int value by key
func Int(name string, defVal ...int) (val int) {
	if str, ok := getVal(name); ok {
		val, err := strconv.ParseInt(str, 10, 0)
		if err == nil {
			return int(val)
		}
	}

	if len(defVal) > 0 {
		val = defVal[0]
	}
	return
}

func getVal(name string) (val string, ok bool) {
	if UpperEnvKey {
		name = strings.ToUpper(name)
	}

	// cached
	if val = loadedData[name]; val != "" {
		ok = true
		return
	}

	// NOTICE: if is windows OS, os.Getenv() Key is not case-sensitive
	return os.LookupEnv(name)
}

// load and parse .env file data to os ENV
func loadFile(file string) (err error) {
	fd, err := os.Open(file)
	if err != nil {
		if OnlyLoadExists && os.IsNotExist(err) {
			return nil
		}
		return err
	}

	//noinspection GoUnhandledErrorResult
	defer fd.Close()

	// parse file contents
	p := parser.NewLite()
	if _, err = p.ParseFrom(bufio.NewScanner(fd)); err != nil {
		return
	}

	// set data to os ENV
	if mp := p.LiteSection(p.DefSection); len(mp) > 0 {
		err = LoadFromMap(mp)
	}
	return
}
