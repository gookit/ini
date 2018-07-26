/*
Package parser is a Parser for parse INI format content to golang data

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
	// quote ' "
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

// Unit8 to uint8
func (m parseMode) Unit8() uint8 {
	return uint8(m)
}

// UserCollector custom data collector.
// notice: in simple mode, isArr always is false.
type UserCollector func(section, key, val string, isArr bool)

// Parser definition
type Parser struct {
	// for full parse(allow array, map section)
	fullData map[string]interface{}

	// for simple parse(section only allow map[string]string)
	simpleData map[string]map[string]string

	// parsed    bool
	parseMode parseMode

	// options
	IgnoreCase bool
	DefSection string
	// only for full parse mode
	NoDefSection bool

	// you can custom data collector
	Collector UserCollector
}

// FullParser create a full mode Parser with some options
func FullParser(opts ...func(*Parser)) *Parser {
	p := &Parser{
		fullData: make(map[string]interface{}),

		parseMode:  ModeFull,
		DefSection: "__default",
	}

	// apply options
	p.WithOptions(opts...)

	return p
}

// SimpleParser create a simple mode Parser
func SimpleParser(opts ...func(*Parser)) *Parser {
	p := &Parser{
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
// Parser.NewWithOptions(ini.ParseEnv)
func NoDefSection(p *Parser) {
	p.NoDefSection = true
}

// IgnoreCase set ignore-case
func IgnoreCase(p *Parser) {
	p.IgnoreCase = true
}

// Parse a INI data string to golang
func Parse(data string, mode parseMode, opts ...func(*Parser)) (p *Parser, err error) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		var ok bool
	// 		if err, ok = r.(errSyntax); ok {
	// 			return
	// 		}
	// 		panic(r)
	// 	}
	// }()

	if mode == ModeFull {
		p = FullParser(opts...)
	} else {
		p = SimpleParser(opts...)
	}

	err = p.ParseString(data)
	return
}

// WithOptions apply some options
func (p *Parser) WithOptions(opts ...func(*Parser)) {
	// apply options
	for _, opt := range opts {
		opt(p)
	}
}

// ParseFrom a data scanner
func (p *Parser) ParseFrom(in *bufio.Scanner) (n int64, err error) {
	n, err = p.parse(in)

	return
}

// ParseString parse from string data
func (p *Parser) ParseString(data string) error {
	var err error

	if strings.TrimSpace(data) == "" {
		return nil
	}

	buf := &bytes.Buffer{}
	buf.WriteString(data)

	scanner := bufio.NewScanner(buf)
	_, err = p.ParseFrom(scanner)

	return err
}

// ParsedData get parsed data
func (p *Parser) ParsedData() interface{} {
	if p.parseMode == ModeFull {
		return p.fullData
	}

	return p.simpleData
}

// ParseMode get current mode
func (p *Parser) ParseMode() uint8 {
	return uint8(p.parseMode)
}

// FullData get parsed data by full parse
func (p *Parser) FullData() map[string]interface{} {
	return p.fullData
}

// SimpleData get parsed data by simple parse
func (p *Parser) SimpleData() map[string]map[string]string {
	return p.simpleData
}

// Reset Parser, clear parsed data
func (p *Parser) Reset() {
	// p.parsed = false

	if p.parseMode == ModeFull {
		p.fullData = make(map[string]interface{})
	} else {
		p.simpleData = make(map[string]map[string]string)
	}
}

// fullParse will parse array item
// ref github.com/dombenson/go-ini
func (p *Parser) parse(in *bufio.Scanner) (bytes int64, err error) {
	section := p.DefSection
	lineNum := 0
	bytes = -1
	readLine := true

	for readLine = in.Scan(); readLine; readLine = in.Scan() {
		line := in.Text()

		bytes++
		bytes += int64(len(line))

		lineNum++
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			// Skip blank lines
			continue
		}
		if line[0] == ';' || line[0] == '#' {
			// Skip comments
			continue
		}

		if groups := assignArrRegex.FindStringSubmatch(line); groups != nil {
			// skip array parse on simple mode
			if p.parseMode == ModeSimple {
				continue
			}

			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), trimWithQuotes(val)

			if p.Collector != nil {
				p.Collector(section, key, val, true)
			} else {
				p.collectFullValue(section, key, val, true)
			}
		} else if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), trimWithQuotes(val)

			if p.Collector != nil {
				p.Collector(section, key, val, false)
			} else if p.parseMode == ModeFull {
				p.collectFullValue(section, key, val, false)
			} else {
				p.collectMapValue(section, key, val)
			}
		} else if groups := sectionRegex.FindStringSubmatch(line); groups != nil {
			name := strings.TrimSpace(groups[1])
			section = name
		} else {
			err = errSyntax{lineNum, line}
			return
		}
	}

	if bytes < 0 {
		bytes = 0
	}

	err = in.Err()

	return
}

func (p *Parser) collectFullValue(section, key, val string, isArr bool) {
	defSection := p.DefSection

	if p.IgnoreCase {
		key = strings.ToLower(key)
		section = strings.ToLower(section)
		defSection = strings.ToLower(defSection)
	}

	// p.NoDefSection and current section is default section
	if p.NoDefSection && section == defSection {
		if isArr {
			curVal, ok := p.fullData[key]
			if ok {
				switch cd := curVal.(type) {
				case []string:
					p.fullData[key] = append(cd, val)
				}
			} else {
				p.fullData[key] = []string{val}
			}
		} else {
			p.fullData[key] = val
		}
		return
	}

	secData, ok := p.fullData[section]
	// first create
	if !ok {
		if isArr {
			p.fullData[section] = map[string]interface{}{key: []string{val}}
		} else {
			p.fullData[section] = map[string]interface{}{key: val}
		}
		return
	}

	switch sd := secData.(type) {
	case map[string]interface{}: // existed section
		curVal, ok := sd[key]
		if ok {
			switch cv := curVal.(type) {
			case string:
				if isArr {
					sd[key] = []string{cv, val}
				} else {
					sd[key] = val
				}
			case []string:
				sd[key] = append(cv, val)
			default:
				return
			}
		} else {
			if isArr {
				sd[key] = []string{val}
			} else {
				sd[key] = val
			}
		}
		p.fullData[section] = sd
	case string: // found default section value
		if isArr {
			p.fullData[section] = map[string]interface{}{key: []string{val}}
		} else {
			p.fullData[section] = map[string]interface{}{key: val}
		}
	}
}

func (p *Parser) collectMapValue(name string, key, val string) {
	if p.IgnoreCase {
		name = strings.ToLower(name)
		key = strings.ToLower(key)
	}

	if sec, ok := p.simpleData[name]; ok {
		sec[key] = val
		p.simpleData[name] = sec
	} else {
		// create the section if it does not exist
		p.simpleData[name] = map[string]string{key: val}
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
