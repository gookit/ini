package ini_test

import (
	"bytes"
	"fmt"
	"github.com/gookit/ini"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Example() {
	// config, err := LoadFiles("testdata/tesdt.ini")
	// LoadExists will ignore not exists file
	err := ini.LoadExists("testdata/test.ini", "not-exist.ini")
	if err != nil {
		panic(err)
	}

	config := ini.Default()

	// load more, will override prev data by key
	_ = config.LoadStrings(`
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
	_ = config.Set("name", "new name")
	name, ok = config.String("name")
	fmt.Printf("set string\n - ok: %v, val: %v\n", ok, name)

	// export data to file
	// _, err = config.WriteToFile("testdata/export.ini")
	// if err != nil {
	// 	panic(err)
	// }

	// Out:
	// get int
	// - ok: true, val: 100
	// get bool
	// - ok: true, val: true
	// get string
	// - ok: true, val: inhere
	// get section
	// - ok: true, val: map[string]string{"stuff":"things", "newK":"newVal", "key":"val0", "some":"change val"}
	// get sub-value by path 'section.key'
	// - ok: true, val: val0
	// get env 'envKey' val: /bin/zsh
	// get env 'envKey1' val: defValue
	// set string
	// - ok: true, val: new name
}

var iniStr = `# comments
name = inhere
age = 28
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
`

func TestLoad(t *testing.T) {
	st := assert.New(t)

	err := ini.LoadFiles("testdata/test.ini")
	st.Nil(err)
	st.False(ini.IsEmpty())
	st.NotEmpty(ini.Data())

	err = ini.LoadFiles("no-file.ini")
	st.Error(err)

	err = ini.LoadExists("testdata/test.ini", "no-file.ini")
	st.Nil(err)
	st.NotEmpty(ini.Data())

	err = ini.LoadStrings("name = inhere")
	st.Nil(err)
	st.NotEmpty(ini.Data())
	st.False(ini.IsEmpty())

	// reset
	ini.Reset()

	err = ini.LoadStrings(" ")
	st.Nil(err)
	st.Empty(ini.Data())

	// test auto init and load data
	conf := new(ini.Ini)
	err = conf.LoadData(map[string]ini.Section{
		"sec0": {"k": "v"},
	})
	st.Nil(err)
	err = conf.LoadData(map[string]ini.Section{
		"name": {"k": "v"},
	})
	st.Nil(err)

	// test error
	err = conf.LoadFiles("testdata/error.ini")
	st.Error(err)

	err = conf.LoadExists("testdata/error.ini")
	st.Error(err)

	err = conf.LoadStrings("invalid string")
	st.Error(err)
}

func TestBasic(t *testing.T) {
	st := assert.New(t)

	conf := ini.New()
	st.Equal(ini.DefSection, conf.DefSection())

	conf.WithOptions(func(opts *ini.Options) {
		opts.DefSection = "myDef"
	})
	st.Equal("myDef", conf.DefSection())

	err := conf.LoadStrings(iniStr)
	st.Nil(err)

	st.True(conf.HasKey("name"))
	st.False(conf.HasKey("notExist"))

	st.True(conf.HasSection("sec1"))
	st.False(conf.HasSection("notExist"))

	st.Panics(func() {
		conf.WithOptions(ini.IgnoreCase)
	})
}

func TestIgnoreCase(t *testing.T) {
	st := assert.New(t)
	conf := ini.NewWithOptions(ini.IgnoreCase)

	err := conf.LoadStrings(`
kEy = val
[sEc]
sK = val
`)
	st.Nil(err)

	opts := conf.Options()
	st.True(opts.IgnoreCase)

	str, ok := conf.String("KEY")
	st.True(ok)
	st.Equal("val", str)

	str, ok = conf.String("key")
	st.True(ok)
	st.Equal("val", str)

	st.True(conf.Delete("key"))
	st.False(conf.HasKey("kEy"))

	_ = conf.Set("NK", "val1")
	str, ok = conf.String("nk")
	st.True(ok)
	st.Equal("val1", str)

	str, ok = conf.String("Nk")
	st.True(ok)
	st.Equal("val1", str)

	sec, ok := conf.StringMap("sec")
	st.True(ok)
	st.Equal("val", sec["sk"])

	err = conf.NewSection("NewSec", map[string]string{"kEy0": "val"})
	st.Nil(err)

	sec, ok = conf.StringMap("newSec")
	st.True(ok)
	st.Equal("val", sec["key0"])

	_ = conf.SetSection("NewSec", map[string]string{"key1": "val0"})
	str, ok = conf.String("newSec.key1")
	st.True(ok)
	st.Equal("val0", str)

	_ = conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	st.True(conf.HasSection("newSec1"))
	st.True(conf.HasSection("newsec1"))
	st.True(conf.DelSection("newsec1"))
}

func TestReadonly(t *testing.T) {
	st := assert.New(t)
	conf := ini.NewWithOptions(ini.Readonly)

	err := conf.LoadStrings(`
key = val
[sec]
k = v
`)
	st.Nil(err)

	opts := conf.Options()
	st.True(opts.Readonly)

	err = conf.Set("newK", "newV")
	st.Error(err)

	err = conf.LoadData(map[string]ini.Section{
		"sec1": {"k": "v"},
	})
	st.Error(err)

	ok := conf.Delete("key")
	st.False(ok)

	ok = conf.DelSection("sec")
	st.False(ok)

	err = conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	st.Error(err)
	st.False(conf.HasSection("newSec1"))

	err = conf.NewSection("NewSec", map[string]string{"kEy0": "val"})
	st.Error(err)

	// Readonly and ParseVar
	conf = ini.NewWithOptions(ini.Readonly, ini.ParseVar)
	err = conf.LoadStrings(`
key = val
[sec]
k = v
k1 = %(key)s
`)
	st.Nil(err)

	opts = conf.Options()
	st.True(opts.ParseVar)

	str, ok := conf.Get("sec.k1")
	st.True(ok)
	st.Equal("val", str)
}

func TestParseEnv(t *testing.T) {
	st := assert.New(t)
	conf := ini.NewWithOptions(ini.ParseEnv)

	err := conf.LoadStrings(`
key = ${PATH}
invalid = ${invalid
notExist = ${NotExist}
hasDefault = ${HasDef|defValue}
`)
	st.Nil(err)

	opts := conf.Options()
	st.True(opts.ParseEnv)
	st.False(opts.ParseVar)

	str, ok := conf.Get("key")
	st.True(ok)
	st.NotContains(str, "${")

	str, ok = conf.Get("notExist")
	st.True(ok)
	st.Equal("${NotExist}", str)

	str, ok = conf.Get("invalid")
	st.True(ok)
	st.Contains(str, "${")
	st.Equal("${invalid", str)

	str, ok = conf.Get("hasDefault")
	st.True(ok)
	st.NotContains(str, "${")
	st.Equal("defValue", str)
}

func TestParseVar(t *testing.T) {
	st := assert.New(t)
	conf := ini.NewWithOptions(ini.ParseVar)
	err := conf.LoadStrings(`
key = val
ref = %(sec.host)s
invalid = %(secs
notExist = %(varNotExist)s
debug = true
[sec]
enable = %(debug)s
url = http://%(host)s/api
host = localhost
`)
	st.Nil(err)

	opts := conf.Options()
	st.False(opts.IgnoreCase)
	st.True(opts.ParseVar)
	// fmt.Println(conf.Data())

	str, ok := conf.Get("invalid")
	st.True(ok)
	st.Equal("%(secs", str)

	str, ok = conf.Get("notExist")
	st.True(ok)
	st.Equal("%(varNotExist)s", str)

	str, ok = conf.Get("sec.host")
	st.True(ok)
	st.Equal("localhost", str)

	str, ok = conf.Get("ref")
	st.True(ok)
	st.Equal("localhost", str)

	str, ok = conf.Get("sec.enable")
	st.True(ok)
	st.Equal("true", str)

	str, ok = conf.Get("sec.url")
	st.True(ok)
	st.Equal("http://localhost/api", str)

	mp, ok := conf.StringMap("sec")
	st.True(ok)
	st.Equal("true", mp["enable"])
}

func TestOther(t *testing.T) {
	st := assert.New(t)

	err := ini.LoadStrings(iniStr)
	st.Nil(err)

	conf := ini.Default()

	// export as INI string
	buf := &bytes.Buffer{}
	_, err = conf.WriteTo(buf)
	st.Nil(err)

	str := buf.String()
	st.Contains(str, "inhere")
	st.Contains(str, "[sec1]")

	// export as formatted JSON string
	str = conf.PrettyJSON()
	st.Contains(str, "inhere")
	st.Contains(str, "sec1")

	// export to file
	_, err = conf.WriteToFile("not/exist/export.ini")
	st.Error(err)
	n, err := conf.WriteToFile("testdata/export.ini")
	st.True(n > 0)
	st.Nil(err)

	conf.Reset()
	st.Empty(conf.Data())

	conf = ini.New()
	str = conf.PrettyJSON()
	st.Equal("", str)
}
