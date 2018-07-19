package ini

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
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

	iv, ok := config.GetInt("age")
	fmt.Printf("get int\n - ok: %v, val: %v\n", ok, iv)

	bv, ok := config.GetBool("debug")
	fmt.Printf("get bool\n - ok: %v, val: %v\n", ok, bv)

	name, ok := config.GetString("name")
	fmt.Printf("get string\n - ok: %v, val: %v\n", ok, name)

	sec1, ok := config.GetSection("sec1")
	fmt.Printf("get section\n - ok: %v, val: %#v\n", ok, sec1)

	str, ok := config.GetString("sec1.key")
	fmt.Printf("get sub-value by path 'section.key'\n - ok: %v, val: %s\n", ok, str)

	// can parse env name(ParseEnv: true)
	fmt.Printf("get env 'envKey' val: %s\n", config.MustString("shell"))
	fmt.Printf("get env 'envKey1' val: %s\n", config.MustString("noEnv"))

	// set value
	config.Set("name", "new name")
	name, ok = config.GetString("name")
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

func TestIni_Get(t *testing.T) {
	st := assert.New(t)

	conf, err := LoadStrings(iniStr)
	st.Nil(err)

	// get int
	str, ok := conf.Get("age")
	st.True(ok)
	st.Equal("28", str)

	iv, ok := conf.GetInt("age")
	st.True(ok)
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

	bv, ok := conf.GetBool("debug")
	st.True(ok)
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

	str, ok = conf.GetString("notExists")
	st.False(ok)
	st.Equal("", str)

	str = conf.DefString("notExists", "defVal")
	st.Equal("defVal", str)

	str = conf.MustString("name")
	st.Equal("inhere", str)

	str = conf.MustString("notExists")
	st.Equal("", str)

	str, ok = conf.GetString("hasQuota1")
	st.True(ok)
	st.Equal("this is val", str)

	str, ok = conf.GetString("hasquota1")
	st.False(ok)
	st.Equal("", str)

	// get by path
	str, ok = conf.Get("sec1.some")
	st.True(ok)
	st.Equal("value", str)

	mp, ok := conf.GetStringMap("sec1")
	st.True(ok)
	st.Equal("val0", mp["key"])
}

func TestIni_Set(t *testing.T) {
	st := assert.New(t)

	conf, err := LoadStrings(iniStr)
	st.Nil(err)

	conf.Set("key", "val", "newSec")
	st.True(conf.HasSection("newSec"))

	val, ok := conf.Get("newSec.key")
	st.True(ok)
	st.Equal("val", val)

	mp, ok := conf.GetStringMap("newSec")
	st.True(ok)
	st.Equal("val", mp["key"])

	conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	st.True(conf.HasSection("newSec1"))

	mp, ok = conf.GetStringMap("newSec1")
	st.True(ok)
	st.Equal("v0", mp["k0"])

	conf.SetInt("int", 345, "newSec")
	iv, ok := conf.GetInt("newSec.int")
	st.True(ok)
	st.Equal(345, iv)

	conf.SetBool("bol", false, "newSec")
	bv, ok := conf.GetBool("newSec.bol")
	st.True(ok)
	st.False(bv)
}

func TestIgnoreCase(t *testing.T) {
	st := assert.New(t)
	conf := NewWithOptions(IgnoreCase)

	err := conf.LoadStrings(`kEy = val`)
	st.Nil(err)

	opts := conf.Options()
	st.True(opts.IgnoreCase)

	str, ok := conf.GetString("KEY")
	st.True(ok)
	st.Equal("val", str)

	str, ok = conf.GetString("key")
	st.True(ok)
	st.Equal("val", str)

	conf.Set("NK", "val1")

	str, ok = conf.GetString("nk")
	st.True(ok)
	st.Equal("val1", str)

	str, ok = conf.GetString("Nk")
	st.True(ok)
	st.Equal("val1", str)
}
