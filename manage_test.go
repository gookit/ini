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

	str, ok = conf.Get("no-sec.some")
	st.False(ok)
	st.Equal("", str)

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
	st.NotContains(mp["notExist"], "${")

	str, ok = conf.Get(" ")
	st.False(ok)
	st.Equal("", str)
}

func TestIni_Set(t *testing.T) {
	st := assert.New(t)

	err := ini.LoadStrings(iniStr)
	st.Nil(err)

	conf := ini.Default()

	err = conf.Set("float", 34.5)
	st.Nil(err)
	st.Equal("34.5", conf.MustString("float"))

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
