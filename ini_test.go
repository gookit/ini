package ini_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/ini/v2"
	"github.com/gookit/ini/v2/parser"
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

	// Out:
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
age = 23
key = val0
some = value
stuff = things
user_name = inhere
`

func TestLoad(t *testing.T) {
	is := assert.New(t)

	err := ini.LoadFiles("testdata/test.ini")
	is.Nil(err)
	is.False(ini.IsEmpty())
	is.NotEmpty(ini.Data())

	err = ini.LoadFiles("no-file.ini")
	is.Err(err)

	err = ini.LoadExists("testdata/test.ini", "no-file.ini")
	is.Nil(err)
	is.NotEmpty(ini.Data())

	err = ini.LoadStrings("name = inhere")
	is.Nil(err)
	is.NotEmpty(ini.Data())
	is.False(ini.IsEmpty())

	// reset
	ini.Reset()

	err = ini.LoadStrings(" ")
	is.Nil(err)
	is.Empty(ini.Data())

	// load data
	err = ini.LoadData(map[string]ini.Section{
		"sec0": {"k": "v"},
	})
	is.Nil(err)
	is.True(ini.HasKey("sec0.k"))

	// test auto init and load data
	conf := new(ini.Ini)
	err = conf.LoadData(map[string]ini.Section{
		"sec0": {"k": "v"},
	})
	is.Nil(err)
	err = conf.LoadData(map[string]ini.Section{
		"name": {"k": "v"},
	})
	is.Nil(err)

	// test error
	err = conf.LoadFiles("testdata/error.ini")
	is.Err(err)

	err = conf.LoadExists("testdata/error.ini")
	is.Err(err)

	err = conf.LoadStrings("invalid string")
	is.Err(err)

	// reset
	ini.Reset()
}

func TestBasic(t *testing.T) {
	is := assert.New(t)
	defer ini.ResetStd()

	conf := ini.Default()
	is.Eq(parser.DefSection, conf.DefSection())
	is.Eq(parser.DefSection, ini.DefSection())

	conf.WithOptions(func(opts *ini.Options) {
		opts.DefSection = "myDef"
	})
	is.Eq("myDef", conf.DefSection())

	err := conf.LoadStrings(iniStr)
	is.Nil(err)

	is.True(conf.HasKey("name"))
	is.False(conf.HasKey("notExist"))

	is.True(conf.HasSection("sec1"))
	is.False(conf.HasSection("notExist"))

	is.Panics(func() {
		ini.WithOptions(ini.IgnoreCase)
	})
}

func TestIgnoreCase(t *testing.T) {
	is := assert.New(t)
	conf := ini.NewWithOptions(ini.IgnoreCase)

	err := conf.LoadStrings(`
kEy = val
[sEc]
sK = val
`)
	is.Nil(err)

	opts := conf.Options()
	is.True(opts.IgnoreCase)

	str := conf.String("KEY")
	is.Eq("val", str)

	str = conf.String("key")
	is.Eq("val", str)

	is.True(conf.Delete("key"))
	is.False(conf.HasKey("kEy"))

	_ = conf.Set("NK", "val1")
	str = conf.String("nk")
	is.Eq("val1", str)

	str = conf.String("Nk")
	is.Eq("val1", str)

	sec := conf.StringMap("sec")
	is.Eq("val", sec["sk"])

	err = conf.NewSection("NewSec", map[string]string{"kEy0": "val"})
	is.Nil(err)

	sec = conf.StringMap("newSec")
	is.Eq("val", sec["key0"])

	_ = conf.SetSection("NewSec", map[string]string{"key1": "val0"})
	str = conf.String("newSec.key1")
	is.Eq("val0", str)

	_ = conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	is.True(conf.HasSection("newSec1"))
	is.True(conf.HasSection("newsec1"))
	is.True(conf.DelSection("newsec1"))
}

func TestReadonly(t *testing.T) {
	is := assert.New(t)
	conf := ini.NewWithOptions(ini.Readonly)

	err := conf.LoadStrings(`
key = val
[sec]
k = v
`)
	is.Nil(err)

	opts := conf.Options()
	is.True(opts.Readonly)

	err = conf.Set("newK", "newV")
	is.Err(err)

	err = conf.LoadData(map[string]ini.Section{
		"sec1": {"k": "v"},
	})
	is.Err(err)

	ok := conf.Delete("key")
	is.False(ok)

	ok = conf.DelSection("sec")
	is.False(ok)

	err = conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	is.Err(err)
	is.False(conf.HasSection("newSec1"))

	err = conf.NewSection("NewSec", map[string]string{"kEy0": "val"})
	is.Err(err)

	// Readonly and ParseVar
	conf = ini.NewWithOptions(ini.Readonly, ini.ParseVar)
	err = conf.LoadStrings(`
key = val
[sec]
k = v
k1 = %(key)s
`)
	is.Nil(err)

	opts = conf.Options()
	is.True(opts.ParseVar)

	str := conf.Get("sec.k1")
	is.Eq("val", str)
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
	st.Eq("", str)

	str = conf.Get("invalid")
	st.Contains(str, "${")
	st.Eq("${invalid", str)

	str = conf.Get("hasDefault")
	st.NotContains(str, "${")
	st.Eq("defValue", str)
}

func TestParseVar(t *testing.T) {
	is := assert.New(t)
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
	is.Nil(err)

	opts := conf.Options()
	is.False(opts.IgnoreCase)
	is.True(opts.ParseVar)
	// fmt.Println(conf.Data())

	str := conf.Get("invalid")
	is.Eq("%(secs", str)

	str = conf.Get("notExist")
	is.Eq("%(varNotExist)s", str)

	str = conf.Get("sec.host")
	is.Eq("localhost", str)

	str = conf.Get("ref")
	is.Eq("localhost", str)

	str = conf.Get("sec.enable")
	is.Eq("true", str)

	str = conf.Get("sec.url")
	is.Eq("http://localhost/api", str)

	mp := conf.StringMap("sec")
	is.Eq("true", mp["enable"])
}

func TestParseCustomRef(t *testing.T) {
	is := assert.New(t)
	conf := ini.NewWithOptions(func(opts *ini.Options) {
		*opts = ini.Options{
			IgnoreCase: false,
			ParseEnv:   false,
			ParseVar:   true,
			SectionSep: "|",
			VarOpen:    "${",
			VarClose:   "}",
			Readonly:   true,
		}
	})
	err := conf.LoadStrings(`
key = val
ref = ${sec|host}
invalid = ${secs
notExist = ${varNotExist}
debug = true
[sec]
enable = ${debug}
url = http://${host}/api
host = localhost
`)
	is.Nil(err)

	opts := conf.Options()
	is.False(opts.IgnoreCase)
	is.False(opts.ParseEnv)
	is.True(opts.ParseVar)
	is.Eq("|", opts.SectionSep)
	is.Eq("${", opts.VarOpen)
	is.Eq("}", opts.VarClose)

	str := conf.Get("invalid")
	is.Eq("${secs", str)

	str = conf.Get("notExist")
	is.Eq("${varNotExist}", str)

	str = conf.Get("sec|host")
	is.Eq("localhost", str)

	str = conf.Get("ref")
	is.Eq("localhost", str)

	str = conf.Get("sec|enable")
	is.Eq("true", str)

	str = conf.Get("sec|url")
	is.Eq("http://localhost/api", str)

	mp := conf.StringMap("sec")
	is.Eq("true", mp["enable"])

}

func TestIni_WriteTo(t *testing.T) {
	is := assert.New(t)

	err := ini.LoadStrings(iniStr)
	is.Nil(err)

	ns := ini.SectionKeys(false)
	is.Contains(ns, "sec1")
	is.NotContains(ns, ini.GetOptions().DefSection)

	ns = ini.SectionKeys(true)
	is.Contains(ns, "sec1")
	is.Contains(ns, ini.GetOptions().DefSection)

	conf := ini.Default()

	// export as INI string
	buf := &bytes.Buffer{}
	_, err = conf.WriteTo(buf)
	is.Nil(err)

	str := buf.String()
	is.Contains(str, "inhere")
	is.Contains(str, "[sec1]")

	// export as formatted JSON string
	str = conf.PrettyJSON()
	is.Contains(str, "inhere")
	is.Contains(str, "sec1")

	// export to file
	_, err = conf.WriteToFile("not/exist/export.ini")
	is.Err(err)

	n, err := conf.WriteToFile("testdata/export.ini")
	is.True(n > 0)
	is.Nil(err)

	conf.Reset()
	is.Empty(conf.Data())

	conf = ini.New()
	str = conf.PrettyJSON()
	is.Eq("", str)
}
