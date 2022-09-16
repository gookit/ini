package parser

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/testutil/assert"
)

var iniStr = `
# comments
name = inhere
age = 28
debug = true
hasQuota1 = 'this is val'
hasQuota2 = "this is val1"
shell = ${SHELL}
noEnv = ${NotExist|defValue}

; array in def section
tags[] = a
tags[] = b
tags[] = c

; comments
[sec1]
key = val0
some = value
stuff = things
; array in section
types[] = x
types[] = y
`

func ExampleNewFulled() {
	p, err := Parse(iniStr, ModeFull)
	// p, err := Parse(iniStr, ModeFull, NoDefSection)
	if err != nil {
		panic(err)
	}

	fmt.Printf("full parse:\n%#v\n", p.FullData())
}

func ExampleNewSimpled() {
	// simple mode will ignore all array values
	p, err := Parse(iniStr, ModeSimple)
	if err != nil {
		panic(err)
	}

	fmt.Printf("simple parse:\n%#v\n", p.SimpleData())
}

func TestParse(t *testing.T) {
	is := assert.New(t)

	p, err := Parse("invalid", ModeFull)
	is.Err(err)
	is.True(len(p.FullData()) == 0)

	p, err = Parse("invalid", ModeSimple)
	is.Err(err)
	is.True(len(p.LiteData()) == 0)
	is.True(len(p.SimpleData()) == 0)
}

func TestDecode(t *testing.T) {
	is := assert.New(t)
	bts := []byte(`
age = 23
name = inhere
arr[] = a
arr[] = b
; comments
[sec]
key = val
; comments
[sec1]
key = val
number = 2020
two_words = abc def
`)

	data := make(map[string]interface{})
	err := Decode([]byte(""), data)
	is.Err(err)

	err = Decode(bts, nil)
	is.Err(err)

	err = Decode(bts, data)
	is.Err(err)

	err = Decode([]byte(`invalid`), &data)
	is.Err(err)

	err = Decode(bts, &data)
	dump.P(data)

	is.Nil(err)
	is.True(len(data) > 0)
	is.Eq("inhere", data["name"])
	is.Eq("[a b]", fmt.Sprintf("%v", data["arr"]))
	is.Eq("map[key:val]", fmt.Sprintf("%v", data["sec"]))

	st := struct {
		Age  int
		Name string
		Sec1 struct {
			Key      string
			Number   int
			TwoWords string `ini:"two_words"`
		}
	}{}

	is.Nil(Decode(bts, &st))
	dump.P(st)
}

func TestNewSimpled(t *testing.T) {
	is := assert.New(t)

	// simple mode will ignore all array values
	p := NewSimpled()
	is.Eq(ModeLite, p.ParseMode)
	is.Eq(ModeLite.Unit8(), p.ParseMode.Unit8())
	is.False(p.IgnoreCase)
	is.False(p.NoDefSection)

	err := p.ParseString("invalid string")
	is.Err(err)
	is.IsType(errSyntax{}, err)
	is.Contains(err.Error(), "invalid INI syntax on line")

	err = p.ParseString("")
	is.NoErr(err)
	is.True(len(p.SimpleData()) == 0)

	p.Reset()
	err = p.ParseString(iniStr)
	is.Nil(err)

	data := p.SimpleData()
	str := fmt.Sprintf("%v", data)
	is.Contains(str, "hasQuota2:")
	is.NotContains(str, "hasquota1:")

	defSec := p.LiteSection(p.DefSection)
	is.NotEmpty(defSec)

	// ignore case
	p = NewSimpled(IgnoreCase)
	err = p.ParseString(iniStr)
	is.Nil(err)

	v := p.ParsedData()
	is.NotEmpty(v)

	data = p.LiteData()
	str = fmt.Sprintf("%v", data)
	is.Contains(str, "hasquota2:")
	is.NotContains(str, "hasQuota1:")
}

func TestNewFulled(t *testing.T) {
	is := assert.New(t)

	p := NewFulled()
	is.Eq(ModeFull, p.ParseMode)
	is.False(p.IgnoreCase)
	is.False(p.NoDefSection)

	err := p.ParseString("invalid string")
	is.Err(err)

	err = p.ParseString(`
[__default]
newKey = new val
[sec1]
newKey = val5
[newSec]
key = val0
`)
	is.Nil(err)
	dump.P(p.ParsedData())

	p.Reset()
	err = p.ParseString(iniStr)
	is.Nil(err)

	v := p.ParsedData()
	is.NotEmpty(v)

	// options: ignore case
	p = NewFulled(IgnoreCase)
	is.True(p.IgnoreCase)
	err = p.ParseString(iniStr)
	is.Nil(err)

	v = p.ParsedData()
	is.NotEmpty(v)

	data := p.FullData()
	str := fmt.Sprintf("%v", data)
	is.Contains(str, "hasquota2:")
	is.NotContains(str, "hasQuota1:")
}

func TestParser_ParseBytes(t *testing.T) {
	p := NewLite()

	is := assert.New(t)
	err := p.ParseBytes(nil)

	is.NoErr(err)
	is.Len(p.LiteData(), 0)
}

func TestParser_ParseString(t *testing.T) {
	p := New(WithParseMode(ModeFull))
	err := p.ParseString(`
key1 = val1
arr = val2
arr[] = val3
arr[] = val4
`)

	assert.NoErr(t, err)
	assert.NotEmpty(t, p.fullData)
	dump.P(p.ParsedData())
}
