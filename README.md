# INI

[![GoDoc](https://godoc.org/github.com/gookit/ini?status.svg)](https://godoc.org/github.com/gookit/ini)
[![Build Status](https://travis-ci.org/gookit/ini.svg?branch=master)](https://travis-ci.org/gookit/ini)
[![Coverage Status](https://coveralls.io/repos/github/gookit/ini/badge.svg?branch=master)](https://coveralls.io/github/gookit/ini?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/ini)](https://goreportcard.com/report/github.com/gookit/ini)

INI data parse by golang. INI config data management tool library.

> **[中文说明](README.zh-CN.md)**

## Features

- Easy to use(get: `Int` `Int64` `Bool` `String` `StringMap` ..., set: `Set`)
- Support multi file, data load
- Support data override merge
- Support parse ENV variable
- Complete unit test(coverage > 90%)
- Support variable reference, default compatible with Python's configParser format `%(VAR)s`

## More formats

If you want more support for file content formats, recommended use `gookit/config`

- [gookit/config](https://github/gookit/config) - Support multi formats: `JSON`(default), `INI`, `YAML`, `TOML`, `HCL`

## GoDoc

- [doc on gowalker](https://gowalker.org/github.com/gookit/ini)
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
can2arr = val0,val1,val2
shell = ${SHELL}
noEnv = ${NotExist|defValue}
nkey = val in default section

; comments
[sec1]
key = val0
some = value
stuff = things
varRef = %(nkey)s
```

### Load data

```go
package main

import (
	"github.com/gookit/ini/v2"
)

// go run ./examples/demo.go
func main() {
	// config, err := ini.LoadFiles("testdata/tesdt.ini")
	// LoadExists will ignore not exists file
	err := ini.LoadExists("testdata/test.ini", "not-exist.ini")
	if err != nil {
		panic(err)
	}

	// load more, will override prev data by key
	err = ini.LoadStrings(`
age = 100
[sec1]
newK = newVal
some = change val
`)
	// fmt.Printf("%v\n", config.Data())
}
```

### Read data

- Get integer

```go
age := ini.Int("age")
fmt.Print(age) // 100
```

- Get bool

```go
val := ini.Bool("debug")
fmt.Print(val) // true
```

- Get string

```go
name := ini.String("name")
fmt.Print(name) // inhere
```

- Get section data(string map)

```go
val := ini.StringMap("sec1")
fmt.Println(val) 
// map[string]string{"key":"val0", "some":"change val", "stuff":"things", "newK":"newVal"}
```

- Value is ENV var

```go
value := ini.String("shell")
fmt.Printf("%q", value)  // "/bin/zsh"
```

- **Get value by key path**

```go
value := ini.String("sec1.key")
fmt.Print(value) // val0
```

- Use var refer

```go
value := ini.String("sec1.varRef")
fmt.Printf("%q", value) // "val in default section"
```

- Setting new value

```go
// set value
ini.Set("name", "new name")
name = ini.String("name")
fmt.Printf("%q", value) // "new name"
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

## Available options

```go
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
	// default section name. default "__default"
	DefSection string
	// sep char for split key path. default ".", use like "section.subKey"
	SectionSep string
}
```

- setting options for default instance

```go
ini.WithOptions(ini.ParseEnv,ini.ParseVar)
```

- setting options with new instance

```go
cfg := ini.New()
cfg.WithOptions(ini.ParseEnv,ini.ParseVar, func (opts *Options) {
	opts.SectionSep = ":"
	opts.DefSection = "default"
})
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

## Refer 

- [go-ini/ini](https://github.com/go-ini/ini) ini parser and config manage
- [dombenson/go-ini](https://github.com/dombenson/go-ini) ini parser and config manage

## License

**MIT**
