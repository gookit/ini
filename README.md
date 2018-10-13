# ini

[![GoDoc](https://godoc.org/github.com/gookit/ini?status.svg)](https://godoc.org/github.com/gookit/ini)
[![Build Status](https://travis-ci.org/gookit/ini.svg?branch=master)](https://travis-ci.org/gookit/ini)
[![Coverage Status](https://coveralls.io/repos/github/gookit/ini/badge.svg?branch=master)](https://coveralls.io/github/gookit/ini?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/ini)](https://goreportcard.com/report/github.com/gookit/ini)

ini parse by golang. ini config data manage

- easy to use(get: `Int` `Bool` `String` `StringMap`, set: `SetInt` `SetBool` `SetString` ...)
- support multi file, data load
- support data override merge
- support parse ENV variable
- complete unit test(coverage > 90%)
- support variable reference, default compatible with Python's configParser format `%(VAR)s`

> **[中文说明](README_cn.md)**

## More formats

If you want more support for file content formats, recommended use `gookit/config`

- [gookit/config](https://github/gookit/config) - Support multi formats: `JSON`(default), `INI`, `YAML`, `TOML`, `HCL`

## GoDoc

- [doc on gowalker](https://gowalker.org/github.com/gookit/ini)
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
	"github.com/gookit/ini"
)

// go run ./examples/demo.go
func main() {
	// config, err := ini.LoadFiles("testdata/tesdt.ini")
	// LoadExists will ignore not exists file
	config, err := ini.LoadExists("testdata/test.ini", "not-exist.ini")
	if err != nil {
		panic(err)
	}

	// load more, will override prev data by key
	config.LoadStrings(`
age = 100
[sec1]
newK = newVal
some = change val
`)
	// fmt.Printf("%v\n", config.Data())
}
```

### Read data

- get integer

```go
age, ok := config.Int("age")
fmt.Print(ok, age) // true 100
```

- get bool

```go
val, ok := config.Bool("debug")
fmt.Print(ok, age) // true true
```

- get string

```go
name, ok := config.String("name")
fmt.Print(ok, name) // true inhere
```

- get section data(string map)

```go
val, ok := config.StringMap("sec1")
fmt.Println(ok, val) 
// true map[string]string{"key":"val0", "some":"change val", "stuff":"things", "newK":"newVal"}
```

- value is ENV var

```go
value, ok := config.String("shell")
fmt.Printf("%v %q", ok, value)  // true "/bin/zsh"
```

- get value by key path

```go
value, ok := config.String("sec1.key")
fmt.Print(ok, value) // true val0
```

- use var refer

```go
value, ok := config.String("sec1.varRef")
fmt.Printf("%v %q", ok, value) // true "val in default section"
```

- setting new value

```go
// set value
config.Set("name", "new name")
name, ok = config.String("name")
fmt.Printf("%v %q", ok, value) // true "new name"
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

- setting options

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
