package parser

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestEncode(t *testing.T) {
	is := assert.New(t)

	out, err := Encode("invalid")
	is.Nil(out)
	is.Err(err)

	// empty
	out, err = Encode(map[string]interface{}{})
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

	out, err = EncodeWithDefName(sData, "_def")
	is.Nil(err)
	is.NotEmpty(out)

	str = string(out)
	fmt.Println("---- lite mode: ----")
	fmt.Println(str)
	is.NotContains(str, "[_def]")
	is.Contains(str, "[sec]")
	is.Contains(str, "name = inhere")

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

	out, err = EncodeWithDefName(fData, "defSec")
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
