// Package dotnev provide load .env data to os ENV.
//
// Deprecated: pkg name is error, please use dotenv
package dotnev

import (
	"github.com/gookit/ini/v2/dotenv"
)

// LoadedData get all loaded data by dontenv
func LoadedData() map[string]string {
	return dotenv.LoadedData()
}

// ClearLoaded clear the previously set ENV value
func ClearLoaded() {
	dotenv.ClearLoaded()
}

// DontUpperEnvKey don't change key to upper on set ENV
func DontUpperEnvKey() {
	dotenv.DontUpperEnvKey()
}

// Load parse .env file data to os ENV.
//
// Usage:
// 	dotenv.Load("./", ".env")
func Load(dir string, filenames ...string) (err error) {
	return dotenv.Load(dir, filenames...)
}

// LoadExists only load on file exists
func LoadExists(dir string, filenames ...string) error {
	return dotenv.LoadExists(dir, filenames...)
}

// LoadFromMap load data from given string map
func LoadFromMap(kv map[string]string) (err error) {
	return dotenv.LoadFromMap(kv)
}

// Get get os ENV value by name
func Get(name string, defVal ...string) (val string) {
	return dotenv.Get(name, defVal...)
}

// Bool get a bool value by key
func Bool(name string, defVal ...bool) (val bool) {
	return dotenv.Bool(name, defVal...)
}

// Int get a int value by key
func Int(name string, defVal ...int) (val int) {
	return dotenv.Int(name, defVal...)
}
