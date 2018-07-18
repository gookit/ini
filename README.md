# ini

[![GoDoc](https://godoc.org/github.com/gookit/ini?status.svg)](https://godoc.org/github.com/gookit/ini)

ini parse by golang. ini config data manage

> parse content is ref the project: https://github.com/dombenson/go-ini, Thank you very much

## Godoc

- [godoc for gopkg](https://godoc.org/gopkg.in/gookit/ini.v1)
- [godoc for github](https://godoc.org/github.com/gookit/ini)

## Usage

- example data(`testdata/test.ini`):

```ini
# comments
name = inhere
age = 50
debug = true
hasQuota1 = 'this is val'
hasQuota2 = "this is val1"
shell = ${SHELL}
noEnv = ${NotExist|defValue}

; comments
[sec1]
key = val0
some = value
stuff = things
```

- usage

```go
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
```

- output

```text
- get int
 ok: true, val: 100
- get bool
 ok: true, val: true
- get string
 ok: true, val: inhere
- get section
 ok: true, val: map[string]string{"key":"val0", "some":"change val", "stuff":"things", "newK":"newVal"}
- get sub-value by path 'section.key'
 ok: true, val: val0
get env 'envKey' val: /bin/zsh
get env 'envKey1' val: defValue
- set string
 ok: true, val: new name
```

## License

**MIT**
