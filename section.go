package ini

// section data in ini
type MapValue map[string]string
type ArrValue map[string][]string

type Sec struct {
	isArray  bool
	mapValue map[string]string
	arrValue map[string][]string
}
