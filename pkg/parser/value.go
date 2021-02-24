package parser

import (
	"errors"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type valueBuilder interface {
	getValue() interface{}
	setValue(value interface{})
}

type valueType interface {
	toRegex() string
	parse(value string) (interface{}, error)
	isSimple() bool
}

type binaryAsTextFormat int

const (
	binaryFormatBase64 binaryAsTextFormat = iota
	binaryFormatHex
)

type stringValue struct {
	regex string
}

func (v *stringValue) parse(value string) (interface{}, error) {
	return strings.TrimSpace(value), nil
}

func (v *stringValue) toRegex() string {
	if v.regex != "" {
		return v.regex
	}
	return ".*"
}

func (v *stringValue) isSimple() bool {
	return true
}

type boolValue struct {
}

func (v *boolValue) parse(value string) (interface{}, error) {
	return strconv.ParseBool(strings.TrimSpace(value))
}

func (v *boolValue) toRegex() string {
	return "(?:true)|(?:false)"
}

func (v *boolValue) isSimple() bool {
	return true
}

type constantValue struct {
	value string
}

func (v *constantValue) parse(value string) (interface{}, error) {
	return strings.TrimSpace(value), nil
}

func (v *constantValue) toRegex() string {
	return regexp.QuoteMeta(v.value)
}

func (v *constantValue) isSimple() bool {
	return true
}

type intValue struct {
	min big.Int
	max big.Int
}

func (v *intValue) parse(value string) (interface{}, error) {
	return strconv.Atoi(value)
}

func (v *intValue) toRegex() string {
	return "(-)?\\d+"
}

func (v *intValue) isSimple() bool {
	return true
}

type numberValue struct {
	min big.Float
	max big.Float
}

func (v *numberValue) parse(value string) (interface{}, error) {
	return value, nil
}

func (v *numberValue) toRegex() string {
	return "-?(?:\\d+)|(?:\\d*\\.\\d+)"
}

func (v *numberValue) isSimple() bool {
	return true
}

type enumValue struct {
	values map[string]bool
}

func (v *enumValue) parse(value string) (interface{}, error) {
	return strings.TrimSpace(value), nil
}

func (v *enumValue) toRegex() string {
	return "\\S+"
}

func (v *enumValue) isSimple() bool {
	return true
}

type formulaItem struct {
	name      string
	valueType valueType
}

type formulaValue struct {
	formula []formulaItem
	re      *regexp.Regexp
}

func (v *formulaValue) parse(value string) (interface{}, error) {

	if v.re == nil {
		var sb strings.Builder
		for _, token := range v.formula {
			if token.name != "" {
				sb.WriteString("(?P<" + token.name + ">" + token.valueType.toRegex() + ")")
			} else {
				sb.WriteString("(?:" + token.valueType.toRegex() + ")")
			}
		}
		v.re = regexp.MustCompile(sb.String())
	}
	var err error
	parsed := map[string]string{}
	match := v.re.FindStringSubmatch(strings.TrimSpace(value))
	// var keyValue string
	if match != nil {
		for i, name := range v.re.SubexpNames() {
			if i != 0 && name != "" {
				parsed[name] = match[i]
			}
		}
	}

	result := map[string]interface{}{}
	for _, token := range v.formula {
		if token.name != "" {
			result[token.name], err = token.valueType.parse(parsed[token.name])
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func (v *formulaValue) toRegex() string {
	var sb strings.Builder
	for _, token := range v.formula {
		sb.WriteString("(?:" + token.valueType.toRegex() + ")")
	}
	return sb.String()
}

func (v *formulaValue) isSimple() bool {
	return true
}

type sequenceValue struct {
	itemType  valueType
	separator string
	re        *regexp.Regexp
}

func (v *sequenceValue) parse(value string) (interface{}, error) {
	if v.re == nil {
		sep := v.separator
		if sep == "" {
			sep = "\\s"
		}
		v.re = regexp.MustCompile("(?:" + sep + ")?(" + v.itemType.toRegex() + ")")
	}

	var err error
	matches := v.re.FindAllStringSubmatch(strings.TrimSpace(value), -1)
	result := make([]interface{}, len(matches))
	for i, match := range matches {
		result[i], err = v.itemType.parse(match[1])
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (v *sequenceValue) toRegex() string {
	return "(?:" + v.itemType.toRegex() + ")(?:" + v.separator + "(?:" + v.itemType.toRegex() + "))*"
}

func (v *sequenceValue) isSimple() bool {
	return true
}

type binaryValue struct {
	textFormat binaryAsTextFormat
}

func (v *binaryValue) parse(value string) (interface{}, error) {
	return strings.TrimSpace(value), nil
}

func (v *binaryValue) toRegex() string {
	return ".*"
}

func (v *binaryValue) isSimple() bool {
	return true
}

type structValue struct {
	children map[string]interface{}
}

func (v *structValue) parse(value string) (interface{}, error) {
	return nil, errors.New("Not supported")
}

func (v *structValue) toRegex() string {
	return ".*"
}

func (v *structValue) isSimple() bool {
	return false
}
