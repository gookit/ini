package parser

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/strutil/textscan"
	"github.com/gookit/goutil/testutil/assert"
)

var iniStr = `
# comments 1
name = inhere
age = 28
debug = true
hasQuota1 = 'this is val'
hasQuota2 = "this is val1"
shell = ${SHELL}
noEnv = ${NotExist|defValue}

; array in default section
tags[] = a
tags[] = b
tags[] = c

; comments 2
[sec1]
key = val0
some = value
stuff = things

; array in section sec1
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

	is.True(IsCommentChar(';'))
	is.True(IsCommentChar('#'))
	is.False(IsCommentChar('a'))
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

	data := make(map[string]any)
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

	type myConf struct {
		Age  int
		Name string
		Sec1 struct {
			Key      string
			Number   int
			TwoWords string `ini:"two_words"`
		}
	}

	st := &myConf{}
	is.NoErr(Decode(bts, st))
	is.Eq(23, st.Age)
	is.Eq("inhere", st.Name)
	is.Eq(2020, st.Sec1.Number)
	is.Eq("abc def", st.Sec1.TwoWords)
	dump.P(st)

	// Unmarshal
	p := NewLite(func(opt *Options) {
		opt.NoDefSection = true
	})

	st = &myConf{}
	is.NoErr(p.Unmarshal(bts, st))
	is.Eq(23, st.Age)
	is.Eq("inhere", st.Name)
	is.Eq(2020, st.Sec1.Number)
	is.Eq("abc def", st.Sec1.TwoWords)
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
	is.IsType(textscan.ErrScan{}, err)
	// is.Contains(err.Error(), "invalid syntax, no matcher available")
	is.Contains(err.Error(), `line 1: "invalid string"`)

	err = p.ParseString("")
	is.NoErr(err)
	is.True(len(p.SimpleData()) == 0)

	p.Reset()
	err = p.ParseString(iniStr)
	is.Nil(err)
	is.NotEmpty(p.Comments())

	data := p.SimpleData()
	dump.P(data, p.Comments())
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
	dump.P(v, p.Comments())
	is.NotEmpty(v)
	is.ContainsKey(v, "sec1")

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

func TestParser_ParseFrom(t *testing.T) {
	p := New()
	n, err := p.ParseFrom(bufio.NewScanner(strings.NewReader("")))
	assert.Eq(t, int64(0), n)
	assert.NoErr(t, err)
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
	assert.NotEmpty(t, p.FullData())
	dump.P(p.ParsedData())

	p.Reset()
	assert.NoErr(t, p.ParseString(`
# no values
`))
}

func TestParser_multiLineValue(t *testing.T) {
	p := New(WithParseMode(ModeFull))
	err := p.ParseString(`
; comments 1
key1 = """multi line
value for key1
"""

arr[] = val3
; comments 2
arr[] = '''multi line
value at array
'''
`)

	assert.NoErr(t, err)
	data := p.FullData()
	assert.NotEmpty(t, data)
	defMp := data[DefSection].(map[string]any)
	dump.P(defMp)
	assert.Eq(t, "multi line\nvalue for key1\n", defMp["key1"])
	assert.Eq(t, "multi line\nvalue at array\n", maputil.DeepGet(defMp, "arr.1"))
}

func TestParser_valueUrl(t *testing.T) {
	p := NewLite()
	err := p.ParseString(`
url_ip=http://127.0.0.1
url_ip_port=http://127.0.0.1:9090
url_value=https://github.com
url_value1=https://github.com/inhere
`)
	assert.NoErr(t, err)
	data := p.LiteData()
	assert.NotEmpty(t, data)
	defMap := data[DefSection]
	assert.NotEmpty(t, defMap)
	dump.P(defMap)

	sMap := maputil.SMap(defMap)
	assert.Eq(t, "http://127.0.0.1", sMap.Str("url_ip"))
	assert.Eq(t, "http://127.0.0.1:9090", sMap.Str("url_ip_port"))
	assert.Eq(t, "https://github.com/inhere", sMap.Str("url_value1"))
}
