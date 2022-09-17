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
	"io"
	"reflect"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// errSyntax is returned when there is a syntax error in an INI file.
type errSyntax struct {
	Line int
	// Source The contents of the erroneous line, without leading or trailing whitespace
	Source string
}

// Error message return
func (e errSyntax) Error() string {
	return fmt.Sprintf("invalid INI syntax on line %d: %s", e.Line, e.Source)
}

var (
	// match: [section]
	sectionRegex = regexp.MustCompile(`^\[(.*)]$`)
	// match: foo[] = val
	assignArrRegex = regexp.MustCompile(`^([^=\[\]]+)\[][^=]*=(.*)$`)
	// match: key = val
	assignRegex = regexp.MustCompile(`^([^=]+)=(.*)$`)
	// quote ' "
	quotesRegex = regexp.MustCompile(`^(['"])(.*)(['"])$`)
)

// special chars consts
const (
	MultiLineValMarkS = "'''"
	MultiLineValMarkD = `"""`
)

// token consts
const (
	TokMLValMarkS = 'm' // multi line value by single quotes: '''
	TokMLValMarkD = 'M' // multi line value by double quotes: """
)

// Parser definition
type Parser struct {
	*Options
	// parsed bool

	// for full parse(allow array, map section)
	fullData map[string]any
	// for simple parse(section only allow map[string]string)
	liteData map[string]map[string]string
}

// New a lite mode Parser with some options
func New(fns ...OptFunc) *Parser {
	return &Parser{Options: NewOptions(fns...)}
}

// NewLite create a lite mode Parser. alias of New()
func NewLite(fns ...OptFunc) *Parser { return New(fns...) }

// NewSimpled create a lite mode Parser
func NewSimpled(fns ...func(*Parser)) *Parser {
	return New().WithOptions(fns...)
}

// NewFulled create a full mode Parser with some options
func NewFulled(fns ...func(*Parser)) *Parser {
	return New(WithParseMode(ModeFull)).WithOptions(fns...)
}

// Parse a INI data string to golang
func Parse(data string, mode parseMode, opts ...func(*Parser)) (p *Parser, err error) {
	p = New(WithParseMode(mode)).WithOptions(opts...)
	err = p.ParseString(data)
	return
}

// Decode INI content to golang data
func Decode(blob []byte, ptr any) error {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("ini: Decode of non-pointer %s", reflect.TypeOf(ptr))
	}

	p, err := Parse(string(blob), ModeFull, NoDefSection)
	if err != nil {
		return err
	}

	return p.MapStruct(ptr)
}

// NoDefSection set don't return DefSection title
//
// Usage:
//
//	Parser.NoDefSection()
func NoDefSection(p *Parser) { p.NoDefSection = true }

// IgnoreCase set ignore-case
func IgnoreCase(p *Parser) { p.IgnoreCase = true }

// WithOptions apply some options
func (p *Parser) WithOptions(opts ...func(p *Parser)) *Parser {
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Unmarshal parse ini text and decode to struct
func (p *Parser) Unmarshal(v []byte, ptr any) error {
	if err := p.ParseBytes(v); err != nil {
		return err
	}
	return p.MapStruct(ptr)
}

/*************************************************************
 * do parsing
 *************************************************************/

// ParseString parse from string data
func (p *Parser) ParseString(str string) error {
	if str = strings.TrimSpace(str); str == "" {
		return nil
	}
	return p.ParseReader(strings.NewReader(str))
}

// ParseBytes parse from bytes data
func (p *Parser) ParseBytes(bts []byte) (err error) {
	if len(bts) == 0 {
		return nil
	}
	return p.ParseReader(bytes.NewBuffer(bts))
}

// ParseReader parse from io reader
func (p *Parser) ParseReader(r io.Reader) (err error) {
	_, err = p.ParseFrom(bufio.NewScanner(r))
	return
}

// init parser
func (p *Parser) init() {
	// if p.IgnoreCase {
	// 	p.DefSection = strings.ToLower(p.DefSection)
	// }

	if p.ParseMode == ModeFull {
		p.fullData = make(map[string]any)

		if p.Collector == nil {
			p.Collector = p.collectFullValue
		}
	} else {
		p.liteData = make(map[string]map[string]string)

		if p.Collector == nil {
			p.Collector = p.collectLiteValue
		}
	}
}

// ParseFrom a data scanner
func (p *Parser) ParseFrom(in *bufio.Scanner) (bytes int64, err error) {
	p.init()

	bytes = -1
	lineNum := 0
	section := p.DefSection

	var readOk bool
	for readOk = in.Scan(); readOk; readOk = in.Scan() {
		line := in.Text()

		bytes++ // newline
		bytes += int64(len(line))

		lineNum++
		line = strings.TrimSpace(line)
		if len(line) == 0 { // Skip blank lines
			continue
		}

		if line[0] == ';' || line[0] == '#' { // Skip comments
			continue
		}

		// array/slice data
		if matched := assignArrRegex.FindStringSubmatch(line); matched != nil {
			// skip array parse on lite mode
			if p.ParseMode == ModeLite {
				continue
			}

			key, val := strings.TrimSpace(matched[1]), trimWithQuotes(matched[2])

			p.collectValue(section, key, val, true)
		} else if matched := assignRegex.FindStringSubmatch(line); matched != nil {
			key, val := strings.TrimSpace(matched[1]), trimWithQuotes(matched[2])

			p.collectValue(section, key, val, false)
		} else if matched := sectionRegex.FindStringSubmatch(line); matched != nil {
			section = strings.TrimSpace(matched[1])
		} else {
			err = errSyntax{lineNum, line}
			return
		}
	}

	err = in.Err()
	if bytes < 0 {
		bytes = 0
	}
	return
}

func (p *Parser) collectValue(section, key, val string, isSlice bool) {
	if p.IgnoreCase {
		key = strings.ToLower(key)
		section = strings.ToLower(section)
	}

	if p.InlineComment {
		val, _ = splitInlineComment(val)
	}

	if p.ReplaceNl {
		val = strings.ReplaceAll(val, `\n`, "\n")
	}

	p.Collector(section, key, val, false)
}

func (p *Parser) collectFullValue(section, key, val string, isSlice bool) {
	defSec := p.DefSection
	// p.NoDefSection and current section is default section
	if p.NoDefSection && section == defSec {
		if isSlice {
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

	secData, exists := p.fullData[section]
	// first create
	if !exists {
		if isSlice {
			p.fullData[section] = map[string]any{key: []string{val}}
		} else {
			p.fullData[section] = map[string]any{key: val}
		}
		return
	}

	switch sd := secData.(type) {
	case map[string]any: // existed section
		curVal, ok := sd[key]
		if ok {
			switch cv := curVal.(type) {
			case string:
				if isSlice {
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
			if isSlice {
				sd[key] = []string{val}
			} else {
				sd[key] = val
			}
		}
		p.fullData[section] = sd
	case string: // found default section value
		if isSlice {
			p.fullData[section] = map[string]any{key: []string{val}}
		} else {
			p.fullData[section] = map[string]any{key: val}
		}
	}
}

func (p *Parser) collectLiteValue(sec, key, val string, _ bool) {
	if p.IgnoreCase {
		key = strings.ToLower(key)
		sec = strings.ToLower(sec)
	}

	if strMap, ok := p.liteData[sec]; ok {
		strMap[key] = val
		p.liteData[sec] = strMap
	} else {
		// create the section if it does not exist
		p.liteData[sec] = map[string]string{key: val}
	}
}

func splitInlineComment(val string) (string, string) {
	if pos := strings.IndexRune(val, '#'); pos > -1 {
		return strings.TrimRight(val[0:pos], " "), val[pos:]
	}

	if pos := strings.Index(val, "//"); pos > -1 {
		return strings.TrimRight(val[0:pos], " "), val[pos:]
	}
	return val, ""
}

/*************************************************************
 * export data
 *************************************************************/

// Decode mapping the parsed data to struct ptr
func (p *Parser) Decode(ptr any) error {
	return p.MapStruct(ptr)
}

// MapStruct mapping the parsed data to struct ptr
func (p *Parser) MapStruct(ptr any) (err error) {
	if p.ParseMode == ModeFull {
		if p.NoDefSection {
			return mapStruct(p.TagName, p.fullData, ptr)
		}

		// collect all default section data to top
		anyMap := make(map[string]any, len(p.fullData)+4)
		if defData, ok := p.fullData[p.DefSection]; ok {
			for key, val := range defData.(map[string]any) {
				anyMap[key] = val
			}
		}

		for group, mp := range p.fullData {
			if group == p.DefSection {
				continue
			}
			anyMap[group] = mp
		}
		return mapStruct(p.TagName, anyMap, ptr)
	}

	defData := p.liteData[p.DefSection]
	defLen := len(defData)
	anyMap := make(map[string]any, len(p.liteData)+defLen)

	// collect all default section data to top
	if defLen > 0 {
		for key, val := range defData {
			anyMap[key] = val
		}
	}

	for group, smp := range p.liteData {
		if group == p.DefSection {
			continue
		}
		anyMap[group] = smp
	}

	return mapStruct(p.TagName, anyMap, ptr)
}

func mapStruct(tagName string, data any, ptr any) error {
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

/*************************************************************
 * helper methods
 *************************************************************/

// ParsedData get parsed data
func (p *Parser) ParsedData() interface{} {
	if p.ParseMode == ModeFull {
		return p.fullData
	}
	return p.liteData
}

// FullData get parsed data by full parse
func (p *Parser) FullData() map[string]interface{} {
	return p.fullData
}

// LiteData get parsed data by simple parse
func (p *Parser) LiteData() map[string]map[string]string {
	return p.liteData
}

// SimpleData get parsed data by simple parse
func (p *Parser) SimpleData() map[string]map[string]string {
	return p.liteData
}

// LiteSection get parsed data by simple parse
func (p *Parser) LiteSection(name string) map[string]string {
	return p.liteData[name]
}

// Reset parser, clear parsed data
func (p *Parser) Reset() {
	// p.parsed = false
	if p.ParseMode == ModeFull {
		p.fullData = make(map[string]any)
	} else {
		p.liteData = make(map[string]map[string]string)
	}
}

func trimWithQuotes(inputVal string) (filtered string) {
	filtered = strings.TrimSpace(inputVal)
	groups := quotesRegex.FindStringSubmatch(filtered)

	if len(groups) > 2 && groups[1] == groups[3] {
		filtered = groups[2]
	}
	return
}
