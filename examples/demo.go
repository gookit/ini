package main

import (
	"github.com/gookit/ini"
	"fmt"
)

// go run ./examples/demo.go
func main() {
	// config, err := ini.LoadFiles("testdata/tesdt.ini")
	// LoadExists will ignore not exists file
	config, err := ini.LoadExists("testdata/test.ini", "not-exist.ini")
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%v\n", config.Data())

	// load more, will override prev data by key
	config.LoadStrings(`
age = 100
[sec1]
newK = newVal
some = change val
`)
	// fmt.Printf("%v\n", config.Data())

	iv, ok := config.GetInt("age")
	fmt.Printf("- get int\n ok: %v, val: %v\n", ok, iv)

	bv, ok := config.GetBool("debug")
	fmt.Printf("- get bool\n ok: %v, val: %v\n", ok, bv)

	name, ok := config.GetString("name")
	fmt.Printf("- get string\n ok: %v, val: %v\n", ok, name)

	sec1, ok := config.GetSection("sec1")
	fmt.Printf("- get section\n ok: %v, val: %#v\n", ok, sec1)

	str, ok := config.GetString("sec1.key")
	fmt.Printf("- get sub-value by path 'section.key'\n ok: %v, val: %s\n", ok, str)

	// can parse env name(ParseEnv: true)
	fmt.Printf("get env 'envKey' val: %s\n", config.MustString("shell"))
	fmt.Printf("get env 'envKey1' val: %s\n", config.MustString("noEnv"))

	// set value
	config.Set("name", "new name")
	name, ok = config.GetString("name")
	fmt.Printf("- set string\n ok: %v, val: %v\n", ok, name)

	// export data to file
	// _, err = config.WriteToFile("testdata/export.ini")
	// if err != nil {
	// 	panic(err)
	// }
}
