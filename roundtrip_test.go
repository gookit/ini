package ini_test

import (
	"bytes"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/ini/v2"
)

// WriteTo emitted values verbatim while LoadStrings trims whitespace and strips
// one matching quote pair, so passwords/paths and quoted or space-padded values
// were silently corrupted on reload. Check the public round-trip is exact.
func TestWriteToLoadStringsRoundTrip(t *testing.T) {
	is := assert.New(t)

	vals := map[string]string{
		"quoted_dq":   `"quoted"`,
		"quoted_sq":   "'quoted'",
		"lead_space":  "  leading",
		"trail_space": "trailing  ",
		"password":    "  s3cr3t ",
		"win_path":    `C:\dir\`,
		"json_val":    `{"a":1}`,
		"url":         "http://x/y?a=b#frag",
		"plain":       "hello",
	}

	c := ini.New()
	is.NoErr(c.SetSection("sec", vals))

	var buf bytes.Buffer
	_, err := c.WriteTo(&buf)
	is.NoErr(err)

	c2 := ini.New()
	is.NoErr(c2.LoadStrings(buf.String()))

	for k, orig := range vals {
		is.Eq(orig, c2.Get("sec."+k), "value %q corrupted on round-trip", k)
	}
}
