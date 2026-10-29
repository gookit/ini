package parser

import (
	"testing"

	"github.com/gookit/goutil/x/assert"
)

func TestEncodeIntegerArrays(t *testing.T) {
	data := map[string]any{
		"ports":   []int{80, 443, 0, -1, 80},
		"strings": []string{"alpha", "beta"},
		"network": map[string]any{
			"ports": []int{3000, 8080},
			"name":  "dev",
		},
	}
	want := map[string]any{
		"ports":   []string{"80", "443", "0", "-1", "80"},
		"strings": []string{"alpha", "beta"},
		"network": map[string]any{
			"ports": []string{"3000", "8080"},
			"name":  "dev",
		},
	}
	for name, encode := range map[string]func(any) ([]byte, error){
		"Encode":     Encode,
		"EncodeWith": func(v any) ([]byte, error) { return EncodeWith(v, &EncodeOptions{}) },
	} {
		t.Run(name, func(t *testing.T) {
			out, err := encode(data)
			assert.NoErr(t, err)
			p, err := Parse(string(out), ModeFull, NoDefSection)
			assert.NoErr(t, err)
			assert.Eq(t, want, p.FullData())
		})
	}
}

func TestEncodeFullIntegerArrays(t *testing.T) {
	data := map[string]any{
		"default": map[string]any{"ports": []int{80, 443}},
		"network": map[string]any{"ports": []int{3000, 8080}},
	}
	out, err := EncodeFull(data, "default")
	assert.NoErr(t, err)
	assert.NotContains(t, string(out), "[default]")
	p, err := Parse(string(out), ModeFull, NoDefSection)
	assert.NoErr(t, err)
	assert.Eq(t, map[string]any{
		"ports":   []string{"80", "443"},
		"network": map[string]any{"ports": []string{"3000", "8080"}},
	}, p.FullData())
}

func TestEncodeStructIntegerArray(t *testing.T) {
	data := struct {
		Ports []int
		Name  string
	}{Ports: []int{80, 443, 0, -1}, Name: "dev"}
	out, err := Encode(data)
	assert.NoErr(t, err)
	var decoded struct {
		Ports []int
		Name  string
	}
	assert.NoErr(t, Decode(out, &decoded))
	assert.Eq(t, data, decoded)
}
