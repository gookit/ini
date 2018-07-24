# ini

[![GoDoc](https://godoc.org/github.com/gookit/ini?status.svg)](https://godoc.org/github.com/gookit/ini)
[![Build Status](https://travis-ci.org/gookit/ini.svg?branch=master)](https://travis-ci.org/gookit/ini)
[![Coverage Status](https://coveralls.io/repos/github/gookit/ini/badge.svg?branch=master)](https://coveralls.io/github/gookit/ini?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/ini)](https://goreportcard.com/report/github.com/gookit/ini)

使用INI格式作为配置，配置数据的加载，管理，使用

- 使用简单(获取: `Int` `Bool` `String` `StringMap`, 设置: `SetInt` `SetBool` `SetString` ...)
- 支持多文件，数据加载
- 支持数据覆盖合并
- 支持解析 ENV 变量名
- 完善的单元测试(coverage > 90%)

> **[EN README](README.md)**

## Godoc

- [godoc for gopkg](https://godoc.org/gopkg.in/gookit/ini.v1)
- [godoc for github](https://godoc.org/github.com/gookit/ini)

## 快速使用

- 示例数据(`testdata/test.ini`):

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

- 开始使用

```go
package main

import (
	"github.com/gookit/ini"
	"fmt"
)

// go run ./examples/demo.go
func main() {
	// config, err := ini.LoadFiles("testdata/tesdt.ini")
	// LoadExists 将忽略不存在的文件
	config, err := ini.LoadExists("testdata/test.ini", "not-exist.ini")
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%v\n", config.Data())

	// 加载更多，将按键覆盖之前数据
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

- 输出(by `go run ./examples/demo.go`)

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

## 测试

- go tests with cover

```bash
go test ./... -cover
```

- run lint by GoLint

```bash
golint ./... 
```

## 参考 

- [go-ini/ini](https://github.com/go-ini/ini) ini parser and config manage
- [dombenson/go-ini](https://github.com/dombenson/go-ini) ini parser and config manage

## License

**MIT**
