package ini_test

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/ini/v2"
)

func TestOptions_ReplaceNl(t *testing.T) {
	text := `
name = inhere
desc = i'm a developer, use\n go,php,java
`

	m := ini.NewWithOptions(ini.ReplaceNl)
	assert.NoErr(t, m.LoadStrings(text))

	assert.Eq(t, "i'm a developer, use\n go,php,java", m.String("desc"))
}
