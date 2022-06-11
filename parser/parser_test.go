package parser

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/stretchr/testify/assert"
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

func TestNewSimpled(t *testing.T) {
	st := assert.New(t)

	// simple mode will ignore all array values
	p := NewSimpled()
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

func TestNewFulled(t *testing.T) {
	is := assert.New(t)

	p := NewFulled()
	is.Equal(ModeFull.Unit8(), p.ParseMode())
	is.False(p.IgnoreCase)
	is.False(p.NoDefSection)

	err := p.ParseString("invalid string")
	is.Error(err)

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

func TestNewFulled_NoDefSection(t *testing.T) {
	is := assert.New(t)

	// options: NoDefSection
	p := NewFulled(NoDefSection)
	is.Equal(ModeFull.Unit8(), p.ParseMode())
	is.False(p.IgnoreCase)
	is.True(p.NoDefSection)

	err := p.ParseString(iniStr)
	is.Nil(err)

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
	is.Nil(err)
	dump.P(p.ParsedData())
}
