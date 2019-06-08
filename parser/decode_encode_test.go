package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	out, err = Encode(fData, "defSec")
	st.Nil(err)
	st.NotEmpty(out)
	str = string(out)
	st.Contains(str, "[sec]")

	out, err = Encode(fData, "sec")
	st.Nil(err)
	st.NotEmpty(out)
	str = string(out)
	st.NotContains(str, "[sec]")
}
