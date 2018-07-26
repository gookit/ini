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
- 支持变量参考，默认兼容Python的configParser格式 `%(VAR)s`

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

## 变量参考解析

```ini
[portal] 
url = http://%(host)s:%(port)s/Portal
host = localhost 
port = 8080
```

启用变量解析后，将会解析这里的 `%(host)s` 并替换为相应的变量值 `localhost`：

```go
cfg := ini.New()
// 启用变量解析
cfg.WithOptions(ini.ParseVar)

fmt.Print(cfg.MustString("portal.url"))
// OUT: 
// http://localhost:8080/Portal 
```

## 可用选项

```go
type Options struct {
	// 设置为只读模式. default False
	Readonly bool
	// 解析 ENV 变量名称. default True
	ParseEnv bool
	// 解析变量引用 "%(varName)s". default False
	ParseVar bool

	// 变量左侧字符. default "%("
	VarOpen string
	// 变量右侧字符. default ")s"
	VarClose string

	// 忽略键名称大小写. default False
	IgnoreCase bool
	// 默认的section名称. default "__default"
	DefSection string
	// 路径分隔符，当通过key获取子级值时. default ".", 例如 "section.subKey"
	SectionSep string
}
```

- 应用选项

```go
cfg := ini.New()
cfg.WithOptions(ini.ParseEnv,ini.ParseVar, func (opts *Options) {
	opts.SectionSep = ":"
	opts.DefSection = "default"
})
```

## 测试

- 测试并输出覆盖率

```bash
go test ./... -cover
```

- 运行 GoLint 检查

```bash
golint ./... 
```

- 查看代码覆盖率 https://gocover.io/github.com/gookit/ini

## 参考 

- [go-ini/ini](https://github.com/go-ini/ini) ini parser and config manage
- [dombenson/go-ini](https://github.com/dombenson/go-ini) ini parser and config manage

## License

**MIT**
