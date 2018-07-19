package parser

import "errors"

// Encode
func Encode(v interface{}) (out []byte, err error) {
	switch vd := v.(type) {
	case map[string]interface{}: // from full mode
		return EncodeFullData(vd)
	case map[string]map[string]string: // from simple mode
		return EncodeSimpleData(vd)
	default:
		err = errors.New("ini: invalid data to encode as ini")
	}

	return
}

// EncodeFullData
func EncodeFullData(data map[string]interface{}) (out []byte, err error) {

	return
}

// EncodeSimpleData
func EncodeSimpleData(data map[string]map[string]string) (out []byte, err error) {

	return
}
