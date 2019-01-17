package ini_test

import (
	"fmt"
	"github.com/gookit/ini"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIni_Get(t *testing.T) {
	st := assert.New(t)

	err := ini.LoadStrings(iniStr)
	st.Nil(err)

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
	st.Equal(23, iv)

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
	bv = conf.Bool("notExist", false)
	st.Equal(false, bv)

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
	st.True(ok)
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

	err = conf.Set(" ", "val")
	st.Error(err)
	st.False(conf.HasKey(" "))

	err = conf.Set("key", "val", "newSec")
	st.Nil(err)
	st.True(conf.HasSection("newSec"))

	val, ok := conf.Get("newSec.key")
	st.True(ok)
	st.Equal("val", val)

	mp, ok := conf.StringMap("newSec")
	st.True(ok)
	st.Equal("val", mp["key"])

	err = conf.SetSection("newSec1", map[string]string{"k0": "v0"})
	st.Nil(err)
	st.True(conf.HasSection("newSec1"))

	mp, ok = conf.Section("newSec1")
	st.True(ok)
	st.Equal("v0", mp["k0"])

	err = conf.NewSection("NewSec2", map[string]string{"kEy0": "val"})
	st.Nil(err)

	err = conf.SetInt("int", 345, "newSec")
	st.Nil(err)
	iv, ok := conf.Int("newSec.int")
	st.True(ok)
	st.Equal(345, iv)

	err = conf.SetBool("bol", false, "newSec")
	st.Nil(err)
	bv, ok := conf.Bool("newSec.bol")
	st.True(ok)
	st.False(bv)

	err = conf.SetBool("bol", true, "newSec")
	st.Nil(err)
	bv, ok = conf.Bool("newSec.bol")
	st.True(ok)
	st.True(bv)

	err = conf.SetString("name", "new name")
	st.Nil(err)
	str, ok := conf.String("name")
	st.True(ok)
	st.Equal("new name", str)

	err = conf.SetString("can2arr", "va0,val1,val2")
	st.Nil(err)
	_, ok = conf.Strings("can2arr-no", ",")
	st.False(ok)
	ss, ok := conf.Strings("can2arr", ",")
	st.True(ok)
	st.Equal("[va0 val1 val2]", fmt.Sprint(ss))
}

func TestIni_Delete(t *testing.T) {
	st := assert.New(t)

	err := ini.LoadStrings(iniStr)
	st.Nil(err)

	conf := ini.Default()

	st.True(conf.HasKey("name"))
	ok := conf.Delete("name")
	st.True(ok)
	st.False(conf.HasKey("name"))

	st.False(conf.Delete(" "))
	st.False(conf.Delete("no-key"))

	ok = conf.Delete("sec1.notExist")
	st.False(ok)
	ok = conf.Delete("sec1.key")
	st.True(ok)

	ok = conf.Delete("no-sec.key")
	st.False(ok)

	st.True(conf.HasSection("sec1"))

	ok = conf.DelSection("sec1")
	st.True(ok)

	st.False(conf.HasSection("sec1"))
}
