package internal

import "github.com/mitchellh/mapstructure"

// FullToStruct mapping full mode data to a struct ptr.
func FullToStruct(tagName, defSec string, data map[string]any, ptr any) error {
	// collect all default section data to top
	anyMap := make(map[string]any, len(data)+4)
	if defData, ok := data[defSec]; ok {
		for key, val := range defData.(map[string]any) {
			anyMap[key] = val
		}
	}

	for group, mp := range data {
		if group == defSec {
			continue
		}
		anyMap[group] = mp
	}
	return MapStruct(tagName, anyMap, ptr)
}

// LiteToStruct mapping lite mode data to a struct ptr.
func LiteToStruct(tagName, defSec string, data map[string]map[string]string, ptr any) error {
	defMap, ok := data[defSec]
	dataNew := make(map[string]any, len(defMap)+len(data)-1)

	// collect default section data to top
	if ok {
		for key, val := range defMap {
			dataNew[key] = val
		}
	}

	// collect other sections
	for secKey, secVals := range data {
		if secKey != defSec {
			dataNew[secKey] = secVals
		}
	}
	return MapStruct(tagName, dataNew, ptr)
}

// MapStruct mapping data to a struct ptr.
func MapStruct(tagName string, data any, ptr any) error {
	mapConf := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   ptr,
		TagName:  tagName,
		// will auto convert string to int/uint
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(mapConf)
	if err != nil {
		return err
	}
	return decoder.Decode(data)
}
