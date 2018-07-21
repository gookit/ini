package parser

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Decode
func Decode(blob []byte, v interface{}) (err error) {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("ini: Decode of non-pointer %s", reflect.TypeOf(v))
	}

	// if rv.IsNil() {
	// 	return fmt.Errorf("ini: Decode of nil %s", reflect.TypeOf(v))
	// }

	p, err := Parse(string(blob), ModeFull, NoDefSection)
	if err != nil {
		return
	}

	bs, err := json.Marshal(p.fullData)
	if err != nil {
		return
	}

	err = json.Unmarshal(bs, v)
	return
}
