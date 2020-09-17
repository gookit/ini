package ini_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gookit/ini/v2"
	"github.com/stretchr/testify/assert"
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

	iv := config.Int("age")
	fmt.Printf("get int\n - val: %v\n", iv)

	bv := config.Bool("debug")
	fmt.Printf("get bool\n - val: %v\n", bv)

	name := config.String("name")
	fmt.Printf("get string\n - val: %v\n", name)

	sec1 := config.StringMap("sec1")
	fmt.Printf("get section\n - val: %#v\n", sec1)

	str := config.String("sec1.key")
	fmt.Printf("get sub-value by path 'section.key'\n - val: %s\n", str)

	// can parse env name(ParseEnv: true)
	fmt.Printf("get env 'envKey' val: %s\n", config.String("shell"))
	fmt.Printf("get env 'envKey1' val: %s\n", config.String("noEnv"))

	// set value
	_ = config.Set("name", "new name")
	name = config.String("name")
	fmt.Printf("set string\n - val: %v\n", name)

	// export data to file
	// _, err = config.WriteToFile("testdata/export.ini")
	// if err != nil {
	// 	panic(err)
	// }

	// Output:
	// get int
	// - val: 100
	// get bool
	// - val: true
	// get string
	// - val: inhere
	// get section
	// - val: map[string]string{"stuff":"things", "newK":"newVal", "key":"val0", "some":"change val"}
	// get sub-value by path 'section.key'
	// - val: val0
	// get env 'envKey' val: /bin/zsh
	// get env 'envKey1' val: defValue
	// set string
	// - val: new name
}

var iniStr = `# comments
name = inhere
age = 28
debug = true
themes = a,b,c
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

	// load data
	err = ini.LoadData(map[string]ini.Section{
		"sec0": {"k": "v"},
	})
	st.Nil(err)
	st.True(ini.HasKey("sec0.k"))

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

	// reset
	ini.Reset()
}

func TestBasic(t *testing.T) {
	st := assert.New(t)

	conf := ini.Default()
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
		ini.WithOptions(ini.IgnoreCase)
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

	str := conf.String("KEY")
	st.Equal("val", str)

	str = conf.String("key")
	st.Equal("val", str)

	st.True(conf.Delete("key"))
	st.False(conf.HasKey("kEy"))

	_ = conf.Set("NK", "val1")
	str = conf.String("nk")
	st.Equal("val1", str)

	str = conf.String("Nk")
	st.Equal("val1", str)

	sec := conf.StringMap("sec")
	st.Equal("val", sec["sk"])

	err = conf.NewSection("NewSec", map[string]string{"kEy0": "val"})
	st.Nil(err)

	sec = conf.StringMap("newSec")
	st.Equal("val", sec["key0"])

	_ = conf.SetSection("NewSec", map[string]string{"key1": "val0"})
	str = conf.String("newSec.key1")
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

	str := conf.Get("sec.k1")
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

	str := conf.Get("key")
	st.NotContains(str, "${")

	str = conf.Get("notExist")
	st.Equal("${NotExist}", str)

	str = conf.Get("invalid")
	st.Contains(str, "${")
	st.Equal("${invalid", str)

	str = conf.Get("hasDefault")
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

	str := conf.Get("invalid")
	st.Equal("%(secs", str)

	str = conf.Get("notExist")
	st.Equal("%(varNotExist)s", str)

	str = conf.Get("sec.host")
	st.Equal("localhost", str)

	str = conf.Get("ref")
	st.Equal("localhost", str)

	str = conf.Get("sec.enable")
	st.Equal("true", str)

	str = conf.Get("sec.url")
	st.Equal("http://localhost/api", str)

	mp := conf.StringMap("sec")
	st.Equal("true", mp["enable"])
}

func TestOther(t *testing.T) {
	st := assert.New(t)

	err := ini.LoadStrings(iniStr)
	st.Nil(err)

	ns := ini.SectionKeys(false)
	st.Contains(ns, "sec1")
	st.NotContains(ns, ini.GetOptions().DefSection)

	ns = ini.SectionKeys(true)
	st.Contains(ns, "sec1")
	st.Contains(ns, ini.GetOptions().DefSection)

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
