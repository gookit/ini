package ini

import (
	"fmt"
	"bufio"
	"bytes"
	"strings"
	"regexp"
	"reflect"
	"encoding/json"
)

// Encode
func Encode(v interface{}) (out []byte, err error) {

	return
}

// Decode
func Decode(blob []byte, v interface{}) (err error) {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("ini: Decode of non-pointer %s", reflect.TypeOf(v))
	}

	if rv.IsNil() {
		return fmt.Errorf("ini: Decode of nil %s", reflect.TypeOf(v))
	}

	p, err := parse(string(blob))
	if err != nil {
		return
	}

	bs, err := json.Marshal(p.simpleData)
	if err != nil {
		return
	}

	err = json.Unmarshal(bs, v)
	return
}

// ErrSyntax is returned when there is a syntax error in an INI file.
type errSyntax struct {
	Line   int
	Source string // The contents of the erroneous line, without leading or trailing whitespace
}

func (e errSyntax) Error() string {
	return fmt.Sprintf("invalid INI syntax on line %d: %s", e.Line, e.Source)
}

var (
	// [section]
	sectionRegex = regexp.MustCompile(`^\[(.*)\]$`)
	// foo[] = val
	assignArrRegex = regexp.MustCompile(`^([^=\[\]]+)\[\][^=]*=(.*)$`)
	// key = val
	assignRegex = regexp.MustCompile(`^([^=]+)=(.*)$`)
	quotesRegex = regexp.MustCompile(`^(['"])(.*)(['"])$`)
)

type Sec struct {
	isArray  bool
	mapValue map[string]string
	arrValue map[string][]string
}

// parse mode
const FullMode parseMode = 1
const SimpleMode parseMode = 2

type parseMode uint8

// section data in ini
type MapValue map[string]string
type ArrValue map[string][]string

// parser
type parser struct {
	// for full parse(allow array, map section)
	fullData map[string]interface{}

	// for simple parse(section only allow map[string]string)
	simpleData map[string]map[string]string

	ParseMode  parseMode
	IgnoreCase bool
	DefSection string
	// only for full parse mode
	NoDefSection bool
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
func FullParser() *parser {
	return &parser{
		fullData: make(map[string]interface{}),

		ParseMode: FullMode,
		DefSection: "__default",
	}
}

// SimpleParser
func SimpleParser() *parser {
	return &parser{
		simpleData: make(map[string]map[string]string),

		ParseMode: SimpleMode,
		DefSection: "__default",
	}
}

// Parse
func Parse(data string, mode parseMode) (p *parser, err error) {
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
		p = FullParser()
	} else {
		p = SimpleParser()
	}

	err = p.parseString(data)

	return
}

func (p *parser) parseBytes(data []byte) error {
	buf := &bytes.Buffer{}
	buf.Write(data)

	scanner := bufio.NewScanner(buf)
	_, err := p.parse(scanner)

	return err
}

func (p *parser) parseString(data string) error {
	buf := &bytes.Buffer{}
	buf.WriteString(data)

	scanner := bufio.NewScanner(buf)
	_, err := p.parse(scanner)

	return err
}

/*************************************************************
 * full parse
 *************************************************************/

// from github.com/dombenson/go-ini
func (p *parser) fullParse(in *bufio.Scanner) (bytes int64, err error) {
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

		inDef := section == p.DefSection

		// skip array parse
		if groups := assignArrRegex.FindStringSubmatch(line); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), trimWithQuotes(val)
			curVal, ok := file.section(section).arrayValues[key]
			if ok {
				file.section(section).arrayValues[key] = append(curVal, val)
			} else {
				file.section(section).arrayValues[key] = make([]string, 1, 4)
				file.section(section).arrayValues[key][0] = val
			}
		} else if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), trimWithQuotes(val)
			// file.section(section).stringValues[key] = val

			p.simpleData[section] = p.addToSection(section, key, val)
		} else if groups := sectionRegex.FindStringSubmatch(line); groups != nil {
			name := strings.TrimSpace(groups[1])
			section = name
			// Create the section if it does not exist
			// file.section(section)
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

func (p *parser) appendToMapping(section string, isArr bool) {

}

/*************************************************************
 * simple parse
 *************************************************************/

// from github.com/dombenson/go-ini
func (p *parser) simpleParse(in *bufio.Scanner) (bytes int64, err error) {
	return
}

// from github.com/dombenson/go-ini
func (p *parser) parse(in *bufio.Scanner) (bytes int64, err error) {
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

		// skip array parse
		if groups := assignArrRegex.FindStringSubmatch(line); groups != nil {
			// key, val := groups[1], groups[2]
			// key, val = strings.TrimSpace(key), trimWithQuotes(val)
			// curVal, ok := file.section(section).arrayValues[key]
			// if ok {
			// 	file.section(section).arrayValues[key] = append(curVal, val)
			// } else {
			// 	file.section(section).arrayValues[key] = make([]string, 1, 4)
			// 	file.section(section).arrayValues[key][0] = val
			// }
		} else if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), trimWithQuotes(val)
			// file.section(section).stringValues[key] = val

			p.simpleData[section] = p.addToSection(section, key, val)
		} else if groups := sectionRegex.FindStringSubmatch(line); groups != nil {
			name := strings.TrimSpace(groups[1])
			section = name
			// Create the section if it does not exist
			// file.section(section)
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

func (p *parser) addToSection(name string, key, val string) Section {
	if p.IgnoreCase {
		name = strings.ToLower(name)
		key = strings.ToLower(key)
	}

	if sec, ok := p.simpleData[name]; ok {
		sec[key] = val
		return sec
	}

	// create the section if it does not exist
	return Section{key: val}
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
