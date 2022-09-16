package parser_test

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/ini/v2/parser"
)

func TestOptions_ReplaceNl(t *testing.T) {
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
