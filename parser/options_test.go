package parser_test

import (
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/ini/v2/parser"
)

// User struct
type User struct {
	Age  int      `ini:"age"`
	Name string   `ini:"name"`
	Tags []string `ini:"tags"`
}

func TestNoDefSection(t *testing.T) {
	is := assert.New(t)

	// options: NoDefSection
	p := parser.NewFulled(parser.NoDefSection)
	is.Eq(parser.ModeFull, p.ParseMode)
	is.False(p.IgnoreCase)
	is.True(p.NoDefSection)

	err := p.ParseString(`
name = inhere
desc = i'm a developer`)
	is.Nil(err)

	p.Reset()
	is.Empty(p.ParsedData())

	err = p.ParseString(`
age = 345
name = inhere
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
	is.NotEmpty(p.ParsedData())

	mp := p.FullData()
	is.ContainsKey(mp, "age")
	is.ContainsKey(mp, "name")

	u := &User{}
	err = p.Decode(u)
	assert.NoErr(t, err)
	assert.Eq(t, 345, u.Age)
	assert.Eq(t, "inhere", u.Name)
}

func TestReplaceNl(t *testing.T) {
	text := `
name = inhere
desc = i'm a developer, use\n go,php,java
`

	p := parser.New(parser.WithDefSection(""))
	err := p.ParseString(text)
	assert.NoErr(t, err)
	assert.NotEmpty(t, p.LiteData())
	assert.Eq(t, `i'm a developer, use\n go,php,java`, p.LiteSection("")["desc"])

	p = parser.New(parser.WithReplaceNl)
	err = p.ParseString(text)
	assert.NoErr(t, err)
	assert.NotEmpty(t, p.LiteData())
	assert.Eq(t, "i'm a developer, use\n go,php,java", p.LiteSection(p.DefSection)["desc"])
}

func TestWithParseMode_full(t *testing.T) {
	text := `
age = 345
name = inhere
tags[] = go
tags[] = php
tags[] = java
[site]
github = github.com/inhere
`

	// lite mode
	p := parser.New()
	err := p.ParseBytes([]byte(text))
	assert.NoErr(t, err)
	assert.NotEmpty(t, p.LiteData())
	dump.P(p.ParsedData())

	u := &User{}
	err = p.Decode(u)
	assert.NoErr(t, err)
	assert.Eq(t, 345, u.Age)
	assert.Eq(t, "inhere", u.Name)
	assert.Empty(t, u.Tags)

	// full mode
	p = parser.New(parser.WithParseMode(parser.ModeFull))
	err = p.ParseString(text)
	assert.NoErr(t, err)
	assert.NotEmpty(t, p.FullData())
	dump.P(p.ParsedData())

	u1 := &User{}
	err = p.Decode(u1)
	assert.NoErr(t, err)
	assert.Eq(t, 345, u1.Age)
	assert.Eq(t, "inhere", u1.Name)
	assert.NotEmpty(t, u1.Tags)
}

func TestWithTagName(t *testing.T) {
	text := `
age = 345
name = inhere
desc = i'm a developer, use\n go,php,java
[site]
github = github.com/inhere
`

	p := parser.NewLite(parser.WithTagName("json"))
	err := p.ParseString(text)
	assert.NoErr(t, err)
	assert.NotEmpty(t, p.LiteData())

	// User struct
	type User struct {
		Age  int    `json:"age"`
		Name string `json:"name"`
	}

	u := &User{}
	err = p.Decode(u)
	assert.NoErr(t, err)
	assert.Eq(t, 345, u.Age)
	assert.Eq(t, "inhere", u.Name)

	// UserErr struct
	type UserErr struct {
		Age map[int]string `json:"age"`
	}

	ue := &UserErr{}
	err = p.Decode(ue)
	// dump.P(ue)
	assert.Err(t, err)
}
