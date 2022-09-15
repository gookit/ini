package ini_test

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/ini/v2"
)

func TestIni_Get(t *testing.T) {
	is := assert.New(t)

	err := ini.LoadStrings(iniStr)
	is.Nil(err)
	is.Nil(ini.Error())

	conf := ini.Default()

	// get value
	str, ok := conf.GetValue("age")
	is.True(ok)
	is.Eq("28", str)

	str, ok = ini.GetValue("not-exist")
	is.False(ok)
	is.Eq("", str)

	// get
	str = conf.Get("age")
	is.Eq("28", str)

	str = ini.Get("age")
	is.Eq("28", str)

	str = ini.Get("not-exist", "defval")
	is.Eq("defval", str)

	// get int
	iv := conf.Int("age")
	is.Eq(28, iv)

	iv = conf.Int("name")
	is.Eq(0, iv)

	iv = ini.Int("name", 23)
	is.True(ini.HasKey("name"))
	is.Eq(0, iv)
	is.Err(ini.Error())

	iv = conf.Int("age", 34)
	is.Eq(28, iv)
	iv = conf.Int("notExist", 34)
	is.Eq(34, iv)

	iv = conf.Int("age")
	is.Eq(28, iv)
	iv = conf.Int("notExist")
	is.Eq(0, iv)

	// get bool
	str = conf.Get("debug")
	is.True(conf.HasKey("debug"))
	is.Eq("true", str)

	bv := conf.Bool("debug")
	is.True(bv)

	bv = conf.Bool("name")
	is.False(bv)

	bv = ini.Bool("debug", false)
	is.Eq(true, bv)
	bv = conf.Bool("notExist")
	is.Eq(false, bv)

	bv = conf.Bool("notExist", true)
	is.True(bv)

	// get string
	val := conf.Get("name")
	is.Eq("inhere", val)

	str = conf.String("notExists")
	is.Eq("", str)

	str = ini.String("notExists", "defVal")
	is.Eq("defVal", str)

	str = conf.String("name")
	is.Eq("inhere", str)

	str = conf.String("notExists")
	is.Eq("", str)

	str = conf.String("hasQuota1")
	is.Eq("this is val", str)

	str = conf.String("hasquota1")
	is.Eq("", str)

	// get by path
	str = conf.Get("sec1.some")
	is.Eq("value", str)

	str = conf.Get("no-sec.some")
	is.Eq("", str)

	// get string map(section data)
	mp := conf.StringMap("sec1")
	is.Eq("val0", mp["key"])

	mp = ini.StringMap("sec1")
	is.Eq("val0", mp["key"])

	mp = conf.StringMap("notExist")
	is.Len(mp, 0)

	// def section
	mp = conf.StringMap("")
	is.Eq("inhere", mp["name"])
	is.NotContains(mp["notExist"], "${")

	str = conf.Get(" ")
	is.Eq("", str)

	ss := ini.Strings("themes")
	is.Eq([]string{"a", "b", "c"}, ss)

	ini.Reset()
}

func TestInt(t *testing.T) {
	ini.Reset()

	err := ini.LoadStrings(iniStr)
	assert.NoErr(t, err)

	// uint
	assert.Eq(t, uint(28), ini.Uint("age"))
	assert.Eq(t, uint(0), ini.Uint("not-exist"))
	assert.Eq(t, uint(10), ini.Uint("not-exist", 10))

	// int64
	assert.Eq(t, int64(28), ini.Int64("age"))
	assert.Eq(t, int64(0), ini.Int64("not-exist"))
	assert.Eq(t, int64(10), ini.Int64("not-exist", 10))

	ini.Reset()
}

func TestIni_Set(t *testing.T) {
	is := assert.New(t)

	err := ini.LoadStrings(iniStr)
	is.Nil(err)

	conf := ini.Default()

	err = conf.Set("float", 34.5)
	is.Nil(err)
	is.Eq("34.5", conf.String("float"))

	err = ini.Set(" ", "val")
	is.Err(err)
	is.False(conf.HasKey(" "))

	err = conf.Set("key", "val", "newSec")
	is.Nil(err)
	is.True(conf.HasSection("newSec"))

	val := conf.Get("newSec.key")
	is.Eq("val", val)

	mp := conf.StringMap("newSec")
	is.Eq("val", mp["key"])

	err = conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	is.Nil(err)
	is.True(conf.HasSection("newSec1"))

	mp = conf.Section("newSec1")
	is.Eq("v0", mp["k0"])

	err = conf.NewSection("NewSec2", map[string]string{"kEy0": "val"})
	is.Nil(err)

	err = conf.Set("int", 345, "newSec")
	is.Nil(err)
	iv := conf.Int("newSec.int")
	is.Eq(345, iv)

	err = conf.Set("bol", false, "newSec")
	is.Nil(err)
	bv := conf.Bool("newSec.bol")
	is.False(bv)

	err = conf.Set("bol", true, "newSec")
	is.Nil(err)
	bv = conf.Bool("newSec.bol")
	is.True(ini.HasKey("newSec.bol"))
	is.True(bv)

	err = conf.Set("name", "new name")
	is.Nil(err)
	str := conf.String("name")
	is.Eq("new name", str)

	err = conf.Set("can2arr", "va0,val1,val2")
	is.Nil(err)

	ss := conf.Strings("can2arr-no", ",")
	is.Empty(ss)

	ss = conf.Strings("can2arr", ",")
	is.Eq("[va0 val1 val2]", fmt.Sprint(ss))
}

func TestIni_Delete(t *testing.T) {
	st := assert.New(t)

	err := ini.LoadStrings(iniStr)
	st.Nil(err)

	conf := ini.Default()

	st.True(conf.HasKey("name"))
	ok := ini.Delete("name")
	st.True(ok)
	st.False(conf.HasKey("name"))

	st.False(conf.Delete(" "))
	st.False(conf.Delete("no-key"))

	ok = ini.Delete("sec1.notExist")
	st.False(ok)
	ok = conf.Delete("sec1.key")
	st.True(ok)

	ok = conf.Delete("no-sec.key")
	st.False(ok)
	st.True(conf.HasSection("sec1"))

	ok = conf.DelSection("sec1")
	st.True(ok)
	st.False(conf.HasSection("sec1"))

	ini.Reset()
}

func TestIni_MapStruct(t *testing.T) {
	is := assert.New(t)

	err := ini.LoadStrings(iniStr)
	is.Nil(err)

	type User struct {
		Age      int
		Some     string
		UserName string `ini:"user_name"`
		Subs     struct {
			Id  string
			Tag string
		}
	}

	u1 := &User{}
	is.NoErr(ini.MapStruct("sec1", u1))
	is.Eq(23, u1.Age)
	is.Eq("inhere", u1.UserName)
	ini.Reset()

	conf := ini.NewWithOptions(func(opt *ini.Options) {
		opt.DefSection = ""
	})
	err = conf.LoadStrings(`
age = 23
some = value
user_name = inhere
[subs]
id = 22
tag = golang
`)

	is.NoErr(err)

	u2 := &User{}
	is.NoErr(conf.Decode(u2))
	is.Eq(23, u2.Age)
	is.Eq("inhere", u2.UserName)
	is.Eq("golang", u2.Subs.Tag)

	is.Err(conf.MapStruct("not-exist", u2))
}
