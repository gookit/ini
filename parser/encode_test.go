package parser

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/ini/v2/internal"
)

func TestEncode(t *testing.T) {
	is := assert.New(t)

	out, err := Encode("invalid")
	is.Nil(out)
	is.Err(err)

	// empty
	out, err = Encode(map[string]any{})
	is.Nil(out)
	is.Nil(err)

	// empty
	out, err = Encode(map[string]map[string]string{})
	is.Nil(out)
	is.Nil(err)

	// encode simple data
	sData := map[string]map[string]string{
		"_def": {"name": "inhere", "age": "100"},
		"sec":  {"key": "val", "key1": "34"},
	}
	out, err = Encode(sData)
	is.Nil(err)
	is.NotEmpty(out)

	str := string(out)
	is.Contains(str, "[_def]")
	is.Contains(str, "[sec]")
	is.Contains(str, "name = inhere")

	out, err = EncodeSimple(sData, "_def")
	is.Nil(err)
	is.NotEmpty(out)

	str = string(out)
	fmt.Println("---- lite mode: ----")
	fmt.Println(str)
	is.NotContains(str, "[_def]")
	is.Contains(str, "[sec]")
	is.Contains(str, "name = inhere")

	// encode full data
	fData := map[string]any{
		"name":    "inhere",
		"age":     12,
		"debug":   false,
		"defArr":  []string{"a", "b"},
		"defArr1": []int{1, 2},
		// section
		"sec": map[string]any{
			"key0":    "val",
			"key1":    45,
			"arr0":    []int{3, 4},
			"arr1":    []string{"c", "d"},
			"invalid": map[string]int{"k": 23},
		},
	}

	out, err = Encode(fData)
	is.Nil(err)
	is.NotEmpty(out)

	str = string(out)
	fmt.Println("---- full mode: ----")
	fmt.Println(str)
	is.Contains(str, "age = 12")
	is.Contains(str, "debug = false")
	is.Contains(str, "name = inhere")
	is.Contains(str, "defArr[] = a")
	is.Contains(str, "[sec]")
	is.Contains(str, "arr1[] = c")

	out, err = EncodeFull(fData, "defSec")
	is.Nil(err)
	is.NotEmpty(out)
	str = string(out)
	is.Contains(str, "[sec]")

	out, err = EncodeWithDefName(fData, "sec")
	is.Nil(err)
	is.NotEmpty(out)
	str = string(out)
	is.NotContains(str, "[sec]")
}

func TestEncode_struct(t *testing.T) {
	is := assert.New(t)

	// encode a struct
	type Sample struct {
		Debug  bool
		Name   string `json:"name"`
		DefArr []string
	}
	sp := &Sample{
		Debug:  true,
		Name:   "inhere",
		DefArr: []string{"a", "b"},
	}
	out, err := Encode(sp)
	is.Nil(err)
	is.NotEmpty(out)
	str := string(out)
	fmt.Println(str)
	is.Contains(str, "Debug = true")
	is.Contains(str, "name = inhere")
}

var liteData = map[string]map[string]string{
	"z_def": {"name": "inhere", "age": "100"},
	"sec":   {"key": "val", "key1": "34"},
}

func TestEncodeLite(t *testing.T) {
	is := assert.New(t)

	out, err := EncodeLite(liteData, "z_def")
	is.Nil(err)
	is.NotEmpty(out)

	str := string(out)
	fmt.Println("---- lite mode: ----")
	fmt.Println(str)
	is.NotContains(str, "[z_def]")
	is.Contains(str, "[sec]")
	is.Contains(str, "name = inhere")
}

func TestEncodeWith(t *testing.T) {
	is := assert.New(t)

	// with comments and raw value
	out, err := EncodeWith(liteData, &EncodeOptions{
		Comments: map[string]string{
			"z_def":     "; this is a comment",
			"z_def_age": "# comment for age",
		},
		RawValueMap: map[string]string{
			"sec_key": "${ENV_VAR1}",
		},
	})
	str := string(out)
	fmt.Println(str)
	is.StrContains(str, "key = ${ENV_VAR1}")
	is.StrContains(str, "# comment for age")
	is.StrContains(str, "; this is a comment")

	// with nil options
	out, err = EncodeWith(liteData, nil)
	is.Nil(err)
	is.NotEmpty(out)
	str = string(out)
	is.StrContains(str, "[z_def]")

	// invalid params
	_, err = EncodeWith("invalid", nil)
	is.Err(err)
	_, err = EncodeWith(nil, nil)
	is.Err(err)
}

func TestMapStruct_err(t *testing.T) {
	assert.Err(t, internal.MapStruct("json", "invalid", nil))
}

// values the parser mutates on read (trims spaces, strips one matching quote
// pair, treats a trailing backslash or triple-quote prefix as a multi-line
// marker). The encoder must protect them so parse(encode(x)) == x.
var roundTripValues = []string{
	`"quoted"`, "'quoted'", "  leading", "trailing  ", "  s3cr3t ",
	"   ", `C:\path\`, `end\`, `"""`, "'''", `""x""`, `"'mixed'"`,
}

// values that already survive a round-trip verbatim; the encoder must leave the
// serialized bytes unchanged for these.
var verbatimValues = []string{
	"hello", "007", "a=b=c", "a # b", "a ; b", `{"a":1}`,
	"http://x/y?a=b#frag", "", `"`, `'notclosed`, `foo"`, `he said "hi"`,
}

func TestQuoteValue(t *testing.T) {
	is := assert.New(t)
	for _, v := range verbatimValues {
		is.Eq(v, quoteValue(v))
	}
	is.Eq(`""quoted""`, quoteValue(`"quoted"`))
	is.Eq(`"'quoted'"`, quoteValue("'quoted'"))
	is.Eq(`"  leading"`, quoteValue("  leading"))
	is.Eq(`"C:\path\"`, quoteValue(`C:\path\`))
	is.Eq(`'"""'`, quoteValue(`"""`)) // single-quote fallback: `"`-wrap would start a `"""` block
}

func TestEncodeLiteRoundTrip(t *testing.T) {
	is := assert.New(t)
	for _, v := range append(append([]string{}, roundTripValues...), verbatimValues...) {
		out, err := EncodeLite(map[string]map[string]string{"sec": {"k": v}})
		is.NoErr(err)

		p, err := Parse(string(out), ModeLite)
		is.NoErr(err)
		is.Eq(v, p.LiteData()["sec"]["k"], "value %q corrupted on round-trip", v)
	}
}

func TestEncodeFullRoundTrip(t *testing.T) {
	is := assert.New(t)
	// hits every encoder site: section k-v, section array, def-section k-v, def-section array
	data := map[string]any{
		"__default": map[string]any{
			"dq":  `"quoted"`,
			"bs":  `C:\path\`,
			"arr": []string{`"a"`, "  b", `c\`},
		},
		"sec":    map[string]any{"pw": "  s3cr3t "},
		"defkv":  "  spaced  ",
		"defarr": []string{`"z"`, "trail "},
	}
	out, err := EncodeFull(data, "__default")
	is.NoErr(err)

	p, err := Parse(string(out), ModeFull, NoDefSection)
	is.NoErr(err)

	fd := p.FullData()
	is.Eq(`"quoted"`, fd["dq"])
	is.Eq(`C:\path\`, fd["bs"])
	is.Eq([]string{`"a"`, "  b", `c\`}, fd["arr"])
	is.Eq("  spaced  ", fd["defkv"])
	is.Eq([]string{`"z"`, "trail "}, fd["defarr"])
	is.Eq("  s3cr3t ", fd["sec"].(map[string]any)["pw"])
}

func TestEncodeVerbatimBytesUnchanged(t *testing.T) {
	is := assert.New(t)
	for _, v := range verbatimValues {
		out, err := EncodeLite(map[string]map[string]string{"sec": {"k": v}})
		is.NoErr(err)
		is.Contains(string(out), "k = "+v+"\n")
	}
}
