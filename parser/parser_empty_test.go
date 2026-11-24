package parser

import (
	"strings"
	"testing"
)

func TestParserEmptyInputClearsState(t *testing.T) {
	for _, mode := range []struct {
		name string
		mode parseMode
	}{{"lite", ModeLite}, {"full", ModeFull}} {
		t.Run(mode.name, func(t *testing.T) {
			for _, tc := range []struct {
				name  string
				parse func(*Parser) error
			}{
				{"empty string", func(p *Parser) error { return p.ParseString("") }},
				{"whitespace string", func(p *Parser) error { return p.ParseString(" \t\r\n ") }},
				{"nil bytes", func(p *Parser) error { return p.ParseBytes(nil) }},
				{"empty bytes", func(p *Parser) error { return p.ParseBytes([]byte{}) }},
				{"whitespace bytes", func(p *Parser) error { return p.ParseBytes([]byte(" \t\r\n ")) }},
				{"empty reader", func(p *Parser) error { return p.ParseReader(strings.NewReader("")) }},
			} {
				t.Run(tc.name, func(t *testing.T) {
					p := New(WithParseMode(mode.mode))
					if err := p.ParseString("# old comment\nold = value\n[section]\nkey = value\n"); err != nil {
						t.Fatal(err)
					}
					if len(p.Comments()) == 0 || len(p.FullData())+len(p.LiteData()) == 0 {
						t.Fatal("initial parse did not collect data and comments")
					}
					if err := tc.parse(p); err != nil {
						t.Fatal(err)
					}
					if len(p.FullData())+len(p.LiteData()) != 0 {
						t.Errorf("empty input retained old data: %#v", p.ParsedData())
					}
					if len(p.Comments()) != 0 {
						t.Errorf("empty input retained old comments: %#v", p.Comments())
					}
				})
			}
		})
	}
}

func TestParserUnmarshalEmptyInput(t *testing.T) {
	p := NewLite()
	var first struct{ Name string }
	if err := p.Unmarshal([]byte("name = previous\n"), &first); err != nil {
		t.Fatal(err)
	}
	if first.Name != "previous" {
		t.Fatalf("initial Name = %q, want previous", first.Name)
	}
	var next struct{ Name string }
	if err := p.Unmarshal(nil, &next); err != nil {
		t.Fatal(err)
	}
	if next.Name != "" {
		t.Errorf("empty input decoded Name = %q, want empty", next.Name)
	}
}
