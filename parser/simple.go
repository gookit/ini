package parser

import (
	"bufio"
	"strings"
)

/*************************************************************
 * simple parse
 *************************************************************/

func (p *parser) simpleParse(in *bufio.Scanner) (bytes int64, err error) {
	return p.parse(in)
}

// from github.com/dombenson/go-ini
func (p *parser) parse(in *bufio.Scanner) (bytes int64, err error) {
	bytes = -1
	lineNum := 0
	readLine := true
	section := p.DefSection

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

		// skip array parse in simple mode
		if groups := assignArrRegex.FindStringSubmatch(line); groups != nil {
			continue
		} else if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), trimWithQuotes(val)

			if p.Collector != nil {
				p.Collector(section, key, val, false)
			} else  {
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

func (p *parser) collectMapValue(name string, key, val string) {
	if p.IgnoreCase {
		name = strings.ToLower(name)
		key = strings.ToLower(key)
	}

	if sec, ok := p.simpleData[name]; ok {
		sec[key] = val
		p.simpleData[name] =  sec
	} else {
		// create the section if it does not exist
		p.simpleData[name] = map[string]string{key: val}
	}
}

