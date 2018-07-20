package parser

import (
	"fmt"
	"testing"
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

func TestSimpleParser(t *testing.T) {
	st := assert.New(t)

	// simple mode will ignore all array values
	p := SimpleParser()
	st.Equal(ModeSimple, p.ParseMode())
	st.False(p.IgnoreCase)
	st.False(p.NoDefSection)

	err := p.ParseString("invalid string")
	st.Error(err)

	err = p.ParseString(iniStr)
	st.Nil(err)
}

func TestFullParser(t *testing.T) {
	st := assert.New(t)
	p := FullParser()
	st.Equal(ModeFull, p.ParseMode())
	st.False(p.IgnoreCase)
	st.False(p.NoDefSection)

	err := p.ParseString("invalid string")
	st.Error(err)

	err = p.ParseString(iniStr)
	st.Nil(err)

}
