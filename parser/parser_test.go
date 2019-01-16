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
	st.Equal(ModeSimple.Unit8(), p.ParseMode())
	st.False(p.IgnoreCase)
	st.False(p.NoDefSection)

	err := p.ParseString("invalid string")
	st.Error(err)
	st.IsType(errSyntax{}, err)
	st.Contains(err.Error(), "invalid INI syntax on line")

	err = p.ParseString("")
	st.Error(err)
	st.True(len(p.SimpleData()) == 0)

	p.Reset()
	err = p.ParseString(iniStr)
	st.Nil(err)

	data := p.SimpleData()
	str := fmt.Sprintf("%v", data)
	st.Contains(str, "hasQuota2:")
	st.NotContains(str, "hasquota1:")

	// ignore case
	p = NewSimpled(IgnoreCase)
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

	p := NewFulled()
	st.Equal(ModeFull.Unit8(), p.ParseMode())
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
	p = NewFulled(IgnoreCase)
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
	p = NewFulled(NoDefSection)
	st.Equal(ModeFull.Unit8(), p.ParseMode())
	st.False(p.IgnoreCase)
	st.True(p.NoDefSection)

	err = p.ParseString(iniStr)
	st.Nil(err)

	p.Reset()
	err = p.ParseString(`
[__default]
newKey = new val
[sec1]
newKey = val5
[newSec]
key = val0
arr[] = val0
arr[] = val1
[newSec]
key1 = val1
arr[] = val2
`)
	st.Nil(err)
	// fmt.Printf("%#v\n", p.ParsedData())
}
