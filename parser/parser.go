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
// FullMode - will parse array
// SimpleMode - don't parse array value
const (
	FullMode   parseMode = 1
	SimpleMode parseMode = 2
)

type parseMode uint8
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

// FullData
func (p *parser) FullData() map[string]interface{} {
	return p.fullData
}

// SimpleData
func (p *parser) SimpleData() map[string]map[string]string {
	return p.simpleData
}

// FullParser
func FullParser(opts ...func(*parser)) *parser {
	p := &parser{
		fullData: make(map[string]interface{}),

		parseMode:  FullMode,
		DefSection: "__default",
	}

	// apply options
	p.WithOptions(opts...)

	return p
}

// SimpleParser
func SimpleParser(opts ...func(*parser)) *parser {
	p := &parser{
		simpleData: make(map[string]map[string]string),

		parseMode:  SimpleMode,
		DefSection: "__default",
	}

	// apply options
	p.WithOptions(opts...)

	return p
}

// ParseEnv
// usage:
// parser.NewWithOptions(ini.ParseEnv)
func NoDefSection(p *parser) {
	p.NoDefSection = true
}

// IgnoreCase
func IgnoreCase(p *parser) {
	p.IgnoreCase = true
}

// Parse
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

	if mode == FullMode {
		p = FullParser(opts...)
	} else {
		p = SimpleParser(opts...)
	}

	err = p.ParseString(data)
	return
}

// WithOptions
func (p *parser) WithOptions(opts ...func(*parser)) {
	// apply options
	for _, opt := range opts {
		opt(p)
	}
}

// ParseFrom
func (p *parser) ParseFrom(in *bufio.Scanner) (n int64, err error) {
	if p.parseMode == FullMode {
		n, err = p.fullParse(in)
	} else {
		n, err = p.parse(in)
	}

	return
}

// ParseBytes
func (p *parser) ParseBytes(data []byte) error {
	var err error

	if len(data) == 0 {
		return nil
	}

	buf := &bytes.Buffer{}
	buf.Write(data)

	scanner := bufio.NewScanner(buf)

	if p.parseMode == FullMode {
		_, err = p.fullParse(scanner)
	} else {
		_, err = p.parse(scanner)
	}

	return err
}

// ParseString
func (p *parser) ParseString(data string) error {
	var err error

	if strings.TrimSpace(data) == "" {
		return nil
	}

	buf := &bytes.Buffer{}
	buf.WriteString(data)

	scanner := bufio.NewScanner(buf)

	if p.parseMode == FullMode {
		_, err = p.fullParse(scanner)
	} else {
		_, err = p.parse(scanner)
	}

	return err
}

// ParsedData
func (p *parser) ParsedData() interface{} {
	if p.parseMode == FullMode {
		return p.fullData
	}

	return p.simpleData
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
