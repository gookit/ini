# ini

[![GoDoc](https://godoc.org/github.com/gookit/ini?status.svg)](https://godoc.org/github.com/gookit/ini)
[![Build Status](https://travis-ci.org/gookit/ini.svg?branch=master)](https://travis-ci.org/gookit/ini)
[![Coverage Status](https://coveralls.io/repos/github/gookit/ini/badge.svg?branch=master)](https://coveralls.io/github/gookit/ini?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/ini)](https://goreportcard.com/report/github.com/gookit/ini)

ini parse by golang. ini config data manage

- easy to use(get: `Int` `Bool` `String` `StringMap`, set: `SetInt` `SetBool` `SetString` ...)
- support multi file, data load
- support data override merge
- support parse ENV key
- complete unit test(coverage > 90%)
- support variable reference, default compatible with Python's configParser format `%(VAR)s`

> **[中文说明](README_cn.md)**

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
	
	iv, ok := config.Int("age")
	fmt.Printf("get int\n - ok: %v, val: %v\n", ok, iv)

	bv, ok := config.Bool("debug")
	fmt.Printf("get bool\n - ok: %v, val: %v\n", ok, bv)

	name, ok := config.String("name")
	fmt.Printf("get string\n - ok: %v, val: %v\n", ok, name)

	sec1, ok := config.StringMap("sec1")
	fmt.Printf("get section\n - ok: %v, val: %#v\n", ok, sec1)

	str, ok := config.String("sec1.key")
	fmt.Printf("get sub-value by path 'section.key'\n - ok: %v, val: %s\n", ok, str)

	// can parse env name(ParseEnv: true)
	fmt.Printf("get env 'envKey' val: %s\n", config.MustString("shell"))
	fmt.Printf("get env 'envKey1' val: %s\n", config.MustString("noEnv"))

	// set value
	config.Set("name", "new name")
	name, ok = config.String("name")
	fmt.Printf("set string\n - ok: %v, val: %v\n", ok, name)

	// export data to file
	// _, err = config.WriteToFile("testdata/export.ini")
	// if err != nil {
	// 	panic(err)
	// }
}
```

- output(by `go run ./examples/demo.go`)

```text
get int
 - ok: true, val: 100
get bool
 - ok: true, val: true
get string
 - ok: true, val: inhere
get section
 - ok: true, val: map[string]string{"key":"val0", "some":"change val", "stuff":"things", "newK":"newVal"}
get sub-value by path 'section.key'
 - ok: true, val: val0
get env 'envKey' val: /bin/zsh
get env 'envKey1' val: defValue
set string
 - ok: true, val: new name
```

## Variable reference resolution

```ini
[portal] 
url = http://%(host)s:%(port)s/api
host = localhost 
port = 8080
```

If variable resolution is enabled，will parse `%(host)s` and replace it：

```go
cfg := ini.New()
// enable ParseVar
cfg.WithOptions(ini.ParseVar)

fmt.Print(cfg.MustString("portal.url"))
// OUT: 
// http://localhost:8080/api 
```

## Tests

- go tests with cover

```bash
go test ./... -cover
```

- run lint by GoLint

```bash
golint ./... 
```

## Ref 

- [go-ini/ini](https://github.com/go-ini/ini) ini parser and config manage
- [dombenson/go-ini](https://github.com/dombenson/go-ini) ini parser and config manage

## License

**MIT**
