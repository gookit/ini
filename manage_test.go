package ini_test

import (
	"fmt"
	"testing"

	"github.com/gookit/ini/v2"
	"github.com/stretchr/testify/assert"
)

func TestIni_Get(t *testing.T) {
	st := assert.New(t)

	err := ini.LoadStrings(iniStr)
	st.Nil(err)
	st.Nil(ini.Error())

	conf := ini.Default()

	// get value
	str, ok := conf.GetValue("age")
	st.True(ok)
	st.Equal("28", str)

	str, ok = ini.GetValue("not-exist")
	st.False(ok)
	st.Equal("", str)

	// get
	str = conf.Get("age")
	st.Equal("28", str)

	str = ini.Get("age")
	st.Equal("28", str)

	str = ini.Get("not-exist", "defval")
	st.Equal("defval", str)

	// get int
	iv := conf.Int("age")
	st.Equal(28, iv)

	iv = conf.Int("name")
	st.Equal(0, iv)

	iv = ini.Int("name", 23)
	st.True(ini.HasKey("name"))
	st.Equal(0, iv)
	st.Error(ini.Error())

	iv = conf.Int("age", 34)
	st.Equal(28, iv)
	iv = conf.Int("notExist", 34)
	st.Equal(34, iv)

	iv = conf.Int("age")
	st.Equal(28, iv)
	iv = conf.Int("notExist")
	st.Equal(0, iv)

	// get bool
	str = conf.Get("debug")
	st.True(conf.HasKey("debug"))
	st.Equal("true", str)

	bv := conf.Bool("debug")
	st.True(bv)

	bv = conf.Bool("name")
	st.False(bv)

	bv = ini.Bool("debug", false)
	st.Equal(true, bv)
	bv = conf.Bool("notExist")
	st.Equal(false, bv)

	bv = conf.Bool("notExist", true)
	st.True(bv)

	// get string
	val := conf.Get("name")
	st.Equal("inhere", val)

	str = conf.String("notExists")
	st.Equal("", str)

	str = ini.String("notExists", "defVal")
	st.Equal("defVal", str)

	str = conf.String("name")
	st.Equal("inhere", str)

	str = conf.String("notExists")
	st.Equal("", str)

	str = conf.String("hasQuota1")
	st.Equal("this is val", str)

	str = conf.String("hasquota1")
	st.Equal("", str)

	// get by path
	str = conf.Get("sec1.some")
	st.Equal("value", str)

	str = conf.Get("no-sec.some")
	st.Equal("", str)

	// get string map(section data)
	mp := conf.StringMap("sec1")
	st.Equal("val0", mp["key"])

	mp = ini.StringMap("sec1")
	st.Equal("val0", mp["key"])

	mp = conf.StringMap("notExist")
	st.Len(mp, 0)

	// def section
	mp = conf.StringMap("")
	st.Equal("inhere", mp["name"])
	st.NotContains(mp["notExist"], "${")

	str = conf.Get(" ")
	st.Equal("", str)

	ss := ini.Strings("themes")
	st.Equal([]string{"a", "b", "c"}, ss)

	ini.Reset()
}

func TestInt(t *testing.T) {
	ini.Reset()

	err := ini.LoadStrings(iniStr)
	assert.NoError(t, err)

	// uint
	assert.Equal(t, uint(28), ini.Uint("age"))
	assert.Equal(t, uint(0), ini.Uint("not-exist"))
	assert.Equal(t, uint(10), ini.Uint("not-exist", 10))

	// int64
	assert.Equal(t, int64(28), ini.Int64("age"))
	assert.Equal(t, int64(0), ini.Int64("not-exist"))
	assert.Equal(t, int64(10), ini.Int64("not-exist", 10))

	ini.Reset()
}

func TestIni_Set(t *testing.T) {
	st := assert.New(t)

	err := ini.LoadStrings(iniStr)
	st.Nil(err)

	conf := ini.Default()

	err = conf.Set("float", 34.5)
	st.Nil(err)
	st.Equal("34.5", conf.String("float"))

	err = ini.Set(" ", "val")
	st.Error(err)
	st.False(conf.HasKey(" "))

	err = conf.Set("key", "val", "newSec")
	st.Nil(err)
	st.True(conf.HasSection("newSec"))

	val := conf.Get("newSec.key")
	st.Equal("val", val)

	mp := conf.StringMap("newSec")
	st.Equal("val", mp["key"])

	err = conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	st.Nil(err)
	st.True(conf.HasSection("newSec1"))

	mp = conf.Section("newSec1")
	st.Equal("v0", mp["k0"])

	err = conf.NewSection("NewSec2", map[string]string{"kEy0": "val"})
	st.Nil(err)

	err = conf.Set("int", 345, "newSec")
	st.Nil(err)
	iv := conf.Int("newSec.int")
	st.Equal(345, iv)

	err = conf.Set("bol", false, "newSec")
	st.Nil(err)
	bv := conf.Bool("newSec.bol")
	st.False(bv)

	err = conf.Set("bol", true, "newSec")
	st.Nil(err)
	bv = conf.Bool("newSec.bol")
	st.True(ini.HasKey("newSec.bol"))
	st.True(bv)

	err = conf.Set("name", "new name")
	st.Nil(err)
	str := conf.String("name")
	st.Equal("new name", str)

	err = conf.Set("can2arr", "va0,val1,val2")
	st.Nil(err)

	ss := conf.Strings("can2arr-no", ",")
	st.Empty(ss)

	ss = conf.Strings("can2arr", ",")
	st.Equal("[va0 val1 val2]", fmt.Sprint(ss))
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
	is.NoError(ini.MapStruct("sec1", u1))
	is.Equal(23, u1.Age)
	is.Equal("inhere", u1.UserName)
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

	is.NoError(err)

	u2 := &User{}
	is.NoError(conf.MapStruct("", u2))
	is.Equal(23, u2.Age)
	is.Equal("inhere", u2.UserName)
	is.Equal("golang", u2.Subs.Tag)

	is.Error(conf.MapStruct("not-exist", u2))
}
