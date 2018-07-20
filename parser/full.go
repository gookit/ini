package parser

import (
	"bufio"
	"strings"
)

/*************************************************************
 * full parse
 *************************************************************/

// fullParse will parse array item
// ref github.com/dombenson/go-ini
func (p *parser) fullParse(in *bufio.Scanner) (bytes int64, err error) {
	if p.parsed {
		return
	}

	section := p.DefSection
	if p.IgnoreCase && section != "" {
		section = strings.ToLower(section)
	}

	p.parsed = true
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
			} else {
				p.collectFullValue(section, key, val, false)
			}
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

func (p *parser) collectFullValue(section, key, val string, isArr bool) {
	if p.IgnoreCase {
		section = strings.ToLower(section)
		key = strings.ToLower(key)
	}

	// p.NoDefSection and current section is default section
	if p.NoDefSection && section == p.DefSection {
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
