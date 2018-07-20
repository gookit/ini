/*
This a parser for parse INI format content to golang data

There are example data:

	# comments
	name = inhere
	age = 28
	debug = true
	hasQuota1 = 'this is val'
	hasQuota2 = "this is val1"
	shell = ${SHELL}
	noEnv = ${NotExist|defValue}

	; array in def section
	tags[] = a
	tags[] = b
	tags[] = c

	; comments
	[sec1]
	key = val0
	some = value
	stuff = things
	; array in section
	types[] = x
	types[] = y

how to use, please see examples:

*/
package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// ErrSyntax is returned when there is a syntax error in an INI file.
type errSyntax struct {
	Line   int
	Source string // The contents of the erroneous line, without leading or trailing whitespace
}

func (e errSyntax) Error() string {
	return fmt.Sprintf("invalid INI syntax on line %d: %s", e.Line, e.Source)
}

var (
	// match: [section]
	sectionRegex = regexp.MustCompile(`^\[(.*)\]$`)
	// match: foo[] = val
	assignArrRegex = regexp.MustCompile(`^([^=\[\]]+)\[\][^=]*=(.*)$`)
	// match: key = val
	assignRegex = regexp.MustCompile(`^([^=]+)=(.*)$`)
	quotesRegex = regexp.MustCompile(`^(['"])(.*)(['"])$`)
)

// parse mode
// ModeFull - will parse array
// ModeSimple - don't parse array value
const (
	ModeFull   parseMode = 1
	ModeSimple parseMode = 2
)

type parseMode uint8

// UserCollector custom data collector.
// notice: in simple mode, isArr always is false.
type UserCollector func(section, key, val string, isArr bool)

// parser
type parser struct {
	// for full parse(allow array, map section)
	fullData map[string]interface{}

	// for simple parse(section only allow map[string]string)
	simpleData map[string]map[string]string

	parsed    bool
	parseMode parseMode

	// options
	IgnoreCase bool
	DefSection string
	// only for full parse mode
	NoDefSection bool

	// you can custom data collector
	Collector UserCollector
}

// FullParser create a full mode parser
func FullParser(opts ...func(*parser)) *parser {
	p := &parser{
		fullData: make(map[string]interface{}),

		parseMode:  ModeFull,
		DefSection: "__default",
	}

	// apply options
	p.WithOptions(opts...)

	return p
}

// SimpleParser create a simple mode parser
func SimpleParser(opts ...func(*parser)) *parser {
	p := &parser{
		simpleData: make(map[string]map[string]string),

		parseMode:  ModeSimple,
		DefSection: "__default",
	}

	// apply options
	p.WithOptions(opts...)
	return p
}

// NoDefSection set don't return DefSection title
// usage:
// parser.NewWithOptions(ini.ParseEnv)
func NoDefSection(p *parser) {
	p.NoDefSection = true
}

// IgnoreCase set ignore-case
func IgnoreCase(p *parser) {
	p.IgnoreCase = true
}

// Parse a INI data string to golang
func Parse(data string, mode parseMode, opts ...func(*parser)) (p *parser, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(errSyntax); ok {
				return
			}
			panic(r)
		}
	}()

	if mode == ModeFull {
		p = FullParser(opts...)
	} else {
		p = SimpleParser(opts...)
	}

	err = p.ParseString(data)
	return
}

// WithOptions apply some options
func (p *parser) WithOptions(opts ...func(*parser)) {
	// apply options
	for _, opt := range opts {
		opt(p)
	}
}

// ParseFrom a data scanner
func (p *parser) ParseFrom(in *bufio.Scanner) (n int64, err error) {
	if p.parseMode == ModeFull {
		n, err = p.fullParse(in)
	} else {
		n, err = p.parse(in)
	}

	return
}

// ParseBytes parse from byte data
func (p *parser) ParseBytes(data []byte) error {
	var err error

	if len(data) == 0 {
		return nil
	}

	buf := &bytes.Buffer{}
	buf.Write(data)

	scanner := bufio.NewScanner(buf)

	if p.parseMode == ModeFull {
		_, err = p.fullParse(scanner)
	} else {
		_, err = p.parse(scanner)
	}

	return err
}

// ParseString parse from string data
func (p *parser) ParseString(data string) error {
	var err error

	if strings.TrimSpace(data) == "" {
		return nil
	}

	buf := &bytes.Buffer{}
	buf.WriteString(data)

	scanner := bufio.NewScanner(buf)

	if p.parseMode == ModeFull {
		_, err = p.fullParse(scanner)
	} else {
		_, err = p.parse(scanner)
	}

	return err
}

// ParsedData get parsed data
func (p *parser) ParsedData() interface{} {
	if p.parseMode == ModeFull {
		return p.fullData
	}

	return p.simpleData
}

// ParseMode get current mode
func (p *parser) ParseMode() parseMode {
	return p.parseMode
}

// FullData get parsed data by full parse
func (p *parser) FullData() map[string]interface{} {
	return p.fullData
}

// SimpleData get parsed data by simple parse
func (p *parser) SimpleData() map[string]map[string]string {
	return p.simpleData
}

// Reset parser, clear parsed data
func (p *parser) Reset() {
	p.parsed = false

	if p.parseMode == ModeFull {
		p.fullData = make(map[string]interface{})
	} else {
		p.simpleData = make(map[string]map[string]string)
	}
}

func trimWithQuotes(inputVal string) (ret string) {
	ret = strings.TrimSpace(inputVal)
	groups := quotesRegex.FindStringSubmatch(ret)

	if groups != nil {
		if groups[1] == groups[3] {
			ret = groups[2]
		}
	}
	return
}
