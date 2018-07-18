package ini

import "strings"

/*************************************************************
 * data get
 *************************************************************/

// Get
func (ini *Ini) Get(key string) (val string, ok bool) {
	key = strings.Trim(strings.TrimSpace(key), ".")
	if key == "" {
		return
	}

	// default find from default Section
	name := DefSection

	// get val by path. eg "log.dir"
	if strings.Contains(key, ".") {
		ss := strings.SplitN(key, ".", 2)
		name, key = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
	}

	// get section data
	if sec, ok := ini.data[name]; !ok {
		// find from section
		val, ok = sec[key]
	}

	return
}

// GetSection
func (ini *Ini) GetSection(name string) (sec map[string]string, ok bool) {
	sec, ok = ini.data[name]
	return
}

/*************************************************************
 * config set
 *************************************************************/

// Add
func (ini *Ini) Add(key, val string, section ...string) {
	name := DefSection
	if len(section) > 0 {
		name = section[0]
	}

	sec, ok := ini.data[name]
	if ok {
		sec[key] = val
	} else {
		sec = Section{key: val}
	}

	ini.data[name] = sec
}

// SetSection
func (ini *Ini) SetSection(name string, values map[string]string) {
	if old, ok := ini.data[name]; ok {
		ini.data[name] = mergeStringMap(values, old)
	} else {
		ini.data[name] = values
	}
}

// AddSection
func (ini *Ini) AddSection(name string, values map[string]string) {
	ini.data[name] = values
}

// MergeData
func (ini *Ini) MergeData(data map[string]Section) {
	ini.ensureInit()

	if len(ini.data) == 0 {
		ini.data = data
	}

	// append or override setting data
	for name, sec := range data {
		ini.SetSection(name, sec)
	}
}

/*************************************************************
 * helper methods
 *************************************************************/

// Reset all data
func (ini *Ini) Reset() {
	ini.data = make(map[string]Section)
}

// Data get all data
func (ini *Ini) Data() map[string]Section {
	return ini.data
}

// HasSection
func (ini *Ini) HasSection(name string) bool {
	_, ok := ini.data[name]
	return ok
}