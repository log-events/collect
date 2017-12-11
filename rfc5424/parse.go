package rfc5424

import (
	"fmt"
)

type StructuredDataParam map[string]string
type StructuredData map[string]StructuredDataParam

const (
	stateInitial    = iota
	stateID         = iota
	stateParamName  = iota
	stateParamValue = iota
)

func ParseStructuredData(data string) (StructuredData, error) {
	res := make(StructuredData)
	if data == "-" {
		return res, nil
	}

	state := stateInitial
	currentID := ""
	currentParamName := ""
	currentParamValue := ""
	valueQuoted := false
	escapeNextChar := false
	for _, r := range data {
		switch state {
		case stateInitial:
			if r != '[' {
				return nil, fmt.Errorf("Invalid token '%s', expecting '%s'", string(r), "[")
			}
			state = stateID
			currentID = ""
			continue
		case stateID:
			if r == '=' || r == ']' || r == '"' {
				return nil, fmt.Errorf("Invalid token '%s' in sd-id", string(r))
			}
			if r == ' ' && len(currentID) == 0 {
				return nil, fmt.Errorf("Invalid token '%s' in sd-id", string(r))
			}
			if r == ' ' {
				state = stateParamName
				currentParamName = ""
				res[currentID] = make(StructuredDataParam)
				continue
			}
			currentID += string(r)
		case stateParamName:
			if r == ']' || r == '"' {
				return nil, fmt.Errorf("Invalid token '%s'in param-name", string(r))
			}
			if r == '=' {
				state = stateParamValue
				currentParamValue = ""
				valueQuoted = false
				continue
			}
			currentParamName += string(r)
		case stateParamValue:
			if escapeNextChar {
				if r != '"' && r != '\\' && r != ']' {
					currentParamValue += "\\"
				}
				currentParamValue += string(r)
				escapeNextChar = false
				continue
			}
			if r == '\\' {
				escapeNextChar = true
				continue
			}
			if r == '"' {
				if !valueQuoted && len(currentParamValue) == 0 {
					valueQuoted = true
					continue
				} else if valueQuoted {
					valueQuoted = false
					continue
				}
			}
			if (r == ']' || r == ' ') && !valueQuoted {
				res[currentID][currentParamName] = currentParamValue
				state = stateParamName
				currentParamName = ""
				if r == ']' {
					state = stateInitial
				}
				continue
			}
			currentParamValue += string(r)
		}
	}

	return res, nil
}
