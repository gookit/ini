package parser

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
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

func Example_fullParse() {
	p, err := Parse(iniStr, ModeFull)
	// p, err := Parse(iniStr, ModeFull, NoDefSection)
	if err != nil {
		panic(err)
	}

	fmt.Printf("full parse:\n%#v\n", p.FullData())
}

func Example_simpleParse() {
	// simple mode will ignore all array values
	p, err := Parse(iniStr, ModeSimple)
	if err != nil {
		panic(err)
	}

	fmt.Printf("simple parse:\n%#v\n", p.SimpleData())
}

func TestParse(t *testing.T) {
	st := assert.New(t)

	p, err := Parse("invalid", ModeFull)
	st.Error(err)
	st.True(len(p.FullData()) == 0)

	p, err = Parse("invalid", ModeSimple)
	st.Error(err)
	st.True(len(p.SimpleData()) == 0)
}

func TestSimpleParser(t *testing.T) {
	st := assert.New(t)

	// simple mode will ignore all array values
	p := SimpleParser()
	st.Equal(ModeSimple, p.ParseMode())
	st.False(p.IgnoreCase)
	st.False(p.NoDefSection)

	err := p.ParseString("invalid string")
	st.Error(err)
	st.IsType(errSyntax{}, err)

	err = p.ParseString("")
	st.Nil(err)
	st.True(len(p.SimpleData()) == 0)

	p.Reset()
	err = p.ParseString(iniStr)
	st.Nil(err)

	data := p.SimpleData()
	str := fmt.Sprintf("%v", data)
	st.Contains(str, "hasQuota2:")
	st.NotContains(str, "hasquota1:")

	// ignore case
	p = SimpleParser(IgnoreCase)
	err = p.ParseString(iniStr)
	st.Nil(err)

	v := p.ParsedData()
	st.NotEmpty(v)

	data = p.SimpleData()
	str = fmt.Sprintf("%v", data)
	st.Contains(str, "hasquota2:")
	st.NotContains(str, "hasQuota1:")
}

func TestFullParser(t *testing.T) {
	st := assert.New(t)

	p := FullParser()
	st.Equal(ModeFull, p.ParseMode())
	st.False(p.IgnoreCase)
	st.False(p.NoDefSection)

	err := p.ParseString("invalid string")
	st.Error(err)

	err = p.ParseString(`
[__default]
newKey = new val
[sec1]
newKey = val5
[newSec]
key = val0
`)
	st.Nil(err)

	// fmt.Printf("%#v\n", p.ParsedData())

	p.Reset()
	err = p.ParseString(iniStr)
	st.Nil(err)

	v := p.ParsedData()
	st.NotEmpty(v)

	// options: ignore case
	p = FullParser(IgnoreCase)
	st.True(p.IgnoreCase)
	err = p.ParseString(iniStr)
	st.Nil(err)

	v = p.ParsedData()
	st.NotEmpty(v)

	data := p.FullData()
	str := fmt.Sprintf("%v", data)
	st.Contains(str, "hasquota2:")
	st.NotContains(str, "hasQuota1:")

	// options: NoDefSection
	p = FullParser(NoDefSection)
	st.Equal(ModeFull, p.ParseMode())
	st.False(p.IgnoreCase)
	st.True(p.NoDefSection)

	err = p.ParseString(iniStr)
	st.Nil(err)

	err = p.ParseString(`
[__default]
newKey = new val
[sec1]
newKey = val5
[newSec]
key = val0
`)
	st.Nil(err)
	// fmt.Printf("%#v\n", p.ParsedData())
}

func TestDecode(t *testing.T) {
	st := assert.New(t)

	bts := []byte(`
name = inhere
arr[] = a
arr[] = b
; comments
[sec]
key = val
`)
	data := make(map[string]interface{})

	err := Decode([]byte(""), data)
	st.Error(err)

	err = Decode(bts, nil)
	st.Error(err)

	err = Decode(bts, data)
	st.Error(err)

	err = Decode([]byte(`invalid`), &data)
	st.Error(err)

	err = Decode(bts, &data)
	st.Nil(err)
	st.True(len(data) > 0)
	st.Equal("inhere", data["name"])
	st.Equal("[a b]", fmt.Sprintf("%v", data["arr"]))
	st.Equal("map[key:val]", fmt.Sprintf("%v", data["sec"]))
}

func TestEncode(t *testing.T) {
	st := assert.New(t)

	out, err := Encode("invalid")
	st.Nil(out)
	st.Error(err)

	// empty
	out, err = Encode(map[string]interface{}{})
	st.Nil(out)
	st.Nil(err)

	// empty
	out, err = Encode(map[string]map[string]string{})
	st.Nil(out)
	st.Nil(err)

	// encode simple data
	sData := map[string]map[string]string{
		"_def": {"name": "inhere", "age": "100"},
		"sec":  {"key": "val", "key1": "34"},
	}
	out, err = Encode(sData)
	st.Nil(err)
	st.NotEmpty(out)

	str := string(out)
	st.Contains(str, "[_def]")
	st.Contains(str, "[sec]")
	st.Contains(str, "name = inhere")

	out, err = Encode(sData, "_def")
	st.Nil(err)
	st.NotEmpty(out)

	str = string(out)
	st.NotContains(str, "[_def]")
	st.Contains(str, "[sec]")
	st.Contains(str, "name = inhere")

	// encode full data
	fData := map[string]interface{}{
		"name":    "inhere",
		"age":     12,
		"debug":   false,
		"defArr":  []string{"a", "b"},
		"defArr1": []int{1, 2},
		// section
		"sec": map[string]interface{}{
			"key0":    "val",
			"key1":    45,
			"arr0":    []int{3, 4},
			"arr1":    []string{"c", "d"},
			"invalid": map[string]int{"k": 23},
		},
	}

	out, err = Encode(fData)
	st.Nil(err)
	st.NotEmpty(out)

	str = string(out)
	st.Contains(str, "age = 12")
	st.Contains(str, "debug = false")
	st.Contains(str, "name = inhere")
	st.Contains(str, "defArr[] = a")
	st.Contains(str, "[sec]")
	st.Contains(str, "arr1[] = c")
}
