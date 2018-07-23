package ini

// test cover details: https://gocover.io/github.com/gookit/ini
import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Example() {
	// config, err := LoadFiles("testdata/tesdt.ini")
	// LoadExists will ignore not exists file
	config, err := LoadExists("testdata/test.ini", "not-exist.ini")
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

	conf, err := LoadFiles("testdata/test.ini")
	st.Nil(err)
	st.NotEmpty(conf.Data())

	conf, err = LoadFiles("no-file.ini")
	st.Error(err)
	st.Empty(conf.Data())

	conf, err = LoadExists("testdata/test.ini", "no-file.ini")
	st.Nil(err)
	st.NotEmpty(conf.Data())

	conf, err = LoadStrings("name = inhere")
	st.Nil(err)
	st.NotEmpty(conf.Data())

	conf, err = LoadStrings(" ")
	st.Nil(err)
	st.Empty(conf.Data())

	// test auto init and load data
	conf = new(Ini)
	err = conf.LoadData(map[string]Section{
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

	conf, err := LoadStrings(iniStr)
	st.Nil(err)

	st.True(conf.HasKey("name"))
	st.False(conf.HasKey("notExist"))

	st.True(conf.HasSection("sec1"))
	st.False(conf.HasSection("notExist"))

	st.Panics(func() {
		conf.WithOptions(IgnoreCase)
	})

	st.True(conf.DefSection() == DefSection)
}

func TestIni_Get(t *testing.T) {
	st := assert.New(t)

	conf, err := LoadStrings(iniStr)
	st.Nil(err)

	// get int
	str, ok := conf.Get("age")
	st.True(ok)
	st.Equal("28", str)

	iv, ok := conf.Int("age")
	st.True(ok)
	st.Equal(28, iv)

	// invalid
	iv, ok = conf.Int("name")
	st.False(ok)
	st.Equal(0, iv)

	iv = conf.DefInt("age", 34)
	st.Equal(28, iv)
	iv = conf.DefInt("notExist", 34)
	st.Equal(34, iv)

	iv = conf.MustInt("age")
	st.Equal(28, iv)
	iv = conf.MustInt("notExist")
	st.Equal(0, iv)

	// get bool
	str, ok = conf.Get("debug")
	st.True(ok)
	st.Equal("true", str)

	bv, ok := conf.Bool("debug")
	st.True(ok)
	st.Equal(true, bv)

	// invalid
	bv, ok = conf.Bool("name")
	st.False(ok)
	st.False(bv)

	bv = conf.DefBool("debug", false)
	st.Equal(true, bv)
	bv = conf.DefBool("notExist", false)
	st.Equal(false, bv)

	bv = conf.MustBool("debug")
	st.Equal(true, bv)
	bv = conf.MustBool("notExist")
	st.Equal(false, bv)

	// get string
	val, ok := conf.Get("name")
	st.True(ok)
	st.Equal("inhere", val)

	str, ok = conf.String("notExists")
	st.False(ok)
	st.Equal("", str)

	str = conf.DefString("notExists", "defVal")
	st.Equal("defVal", str)

	str = conf.MustString("name")
	st.Equal("inhere", str)

	str = conf.MustString("notExists")
	st.Equal("", str)

	str, ok = conf.String("hasQuota1")
	st.True(ok)
	st.Equal("this is val", str)

	str, ok = conf.String("hasquota1")
	st.False(ok)
	st.Equal("", str)

	// get by path
	str, ok = conf.Get("sec1.some")
	st.True(ok)
	st.Equal("value", str)

	// get string map(section data)
	mp, ok := conf.StringMap("sec1")
	st.True(ok)
	st.Equal("val0", mp["key"])

	mp = conf.MustMap("sec1")
	st.Equal("val0", mp["key"])

	mp = conf.MustMap("notExist")
	st.Len(mp, 0)

	// def section
	mp, ok = conf.StringMap("")
	st.True(ok)
	st.Equal("inhere", mp["name"])

	str, ok = conf.Get(" ")
	st.False(ok)
	st.Equal("", str)
}

func TestIni_Set(t *testing.T) {
	st := assert.New(t)

	conf, err := LoadStrings(iniStr)
	st.Nil(err)

	conf.Set(" ", "val")
	st.False(conf.HasKey(" "))

	conf.Set("key", "val", "newSec")
	st.True(conf.HasSection("newSec"))

	val, ok := conf.Get("newSec.key")
	st.True(ok)
	st.Equal("val", val)

	mp, ok := conf.StringMap("newSec")
	st.True(ok)
	st.Equal("val", mp["key"])

	conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	st.True(conf.HasSection("newSec1"))

	mp, ok = conf.Section("newSec1")
	st.True(ok)
	st.Equal("v0", mp["k0"])

	err = conf.NewSection("NewSec2", map[string]string{"kEy0": "val"})
	st.Nil(err)

	conf.SetInt("int", 345, "newSec")
	iv, ok := conf.Int("newSec.int")
	st.True(ok)
	st.Equal(345, iv)

	conf.SetBool("bol", false, "newSec")
	bv, ok := conf.Bool("newSec.bol")
	st.True(ok)
	st.False(bv)

	conf.SetBool("bol", true, "newSec")
	bv, ok = conf.Bool("newSec.bol")
	st.True(ok)
	st.True(bv)

	conf.SetString("name", "new name")
	str, ok := conf.String("name")
	st.True(ok)
	st.Equal("new name", str)
}

func TestIgnoreCase(t *testing.T) {
	st := assert.New(t)
	conf := NewWithOptions(IgnoreCase)

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

	st.True(conf.Del("key"))
	st.False(conf.HasKey("kEy"))

	conf.Set("NK", "val1")

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

	conf.SetSection("NewSec", map[string]string{"key1": "val0"})
	str, ok = conf.String("newSec.key1")
	st.True(ok)
	st.Equal("val0", str)

	conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	st.True(conf.HasSection("newSec1"))
	st.True(conf.HasSection("newsec1"))
	st.True(conf.DelSection("newsec1"))
}

func TestReadonly(t *testing.T) {
	st := assert.New(t)
	conf := NewWithOptions(Readonly)

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

	err = conf.LoadData(map[string]Section{
		"sec1": Section{"k": "v"},
	})
	st.Error(err)

	ok := conf.Del("key")
	st.False(ok)

	ok = conf.DelSection("sec")
	st.False(ok)

	err = conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	st.Error(err)
	st.False(conf.HasSection("newSec1"))

	err = conf.NewSection("NewSec", map[string]string{"kEy0": "val"})
	st.Error(err)
}

func TestParseEnv(t *testing.T) {
	st := assert.New(t)
	conf := NewWithOptions(ParseEnv)

	err := conf.LoadStrings(`
key = ${PATH}
notExist = ${NotExist|defValue}
`)
	st.Nil(err)

	opts := conf.Options()
	st.True(opts.ParseEnv)

	str, ok := conf.String("key")
	st.True(ok)
	st.NotContains(str, "${")

	str, ok = conf.String("notExist")
	st.True(ok)
	st.NotContains(str, "${")
	st.Equal("defValue", str)
}

func TestIni_Del(t *testing.T) {
	st := assert.New(t)

	conf, err := LoadStrings(iniStr)
	st.Nil(err)

	st.True(conf.HasKey("name"))
	ok := conf.Del("name")
	st.True(ok)
	st.False(conf.HasKey("name"))

	st.False(conf.Del(" "))
	st.False(conf.Del("no-key"))

	ok = conf.Del("sec1.notExist")
	st.False(ok)
	ok = conf.Del("sec1.key")
	st.True(ok)

	st.True(conf.HasSection("sec1"))

	ok = conf.DelSection("sec1")
	st.True(ok)

	st.False(conf.HasSection("sec1"))
}

func TestOther(t *testing.T) {
	st := assert.New(t)

	conf, err := LoadStrings(iniStr)
	st.Nil(err)

	// export as INI string
	str := conf.Export()
	st.Contains(str, "inhere")
	st.Contains(str, "[sec1]")

	// export as formatted JSON string
	str = conf.PrettyJson()
	st.Contains(str, "inhere")
	st.Contains(str, "sec1")

	// export to file
	n, err := conf.WriteToFile("not/exist/export.ini")
	st.Error(err)
	n, err = conf.WriteToFile("testdata/export.ini")
	st.True(n > 0)
	st.Nil(err)

	conf.Reset()
	st.Empty(conf.Data())

	conf = New()
	conf.data = nil
	str = conf.PrettyJson()

	conf = New()
	str = conf.PrettyJson()
	str = conf.Export()
	st.Equal("", str)
}
