package parser

import (
	"strconv"
	"strings"
)

type binaryAsTextFormat int

const (
	binaryFormatBase64 binaryAsTextFormat = iota
	binaryFormatHex
)

type valueType interface {
	toRegex() string
	parse(value string) (interface{}, error)
}

type stringValue struct{}

func (v *stringValue) parse(value string) (interface{}, error) {
	return strings.TrimSpace(value), nil
}

func (v *stringValue) toRegex() string {
	return ".*"
}

type intValue struct{}

func (v *intValue) parse(value string) (interface{}, error) {
	return strconv.Atoi(value)
}

func (v *intValue) toRegex() string {
	return "(-)?\\d+"
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

type formulaValue struct{}

func (v *formulaValue) parse(value string) (interface{}, error) {
	return nil, nil
}

func (v *formulaValue) toRegex() string {
	return ".*"
}

type sequenceValue struct{}

func (v *sequenceValue) parse(value string) (interface{}, error) {
	return nil, nil
}

func (v *sequenceValue) toRegex() string {
	return ".*"
}

type binaryValue struct {
	textFormat binaryAsTextFormat
}

func (v *binaryValue) parse(value string) (interface{}, error) {
	return nil, nil
}

func (v *binaryValue) toRegex() string {
	return ".*"
}

// httpVerb enum GET POST PUT PATCH DELETE
// hostname string \S+
// networkPort int 0..65535 #Network port number
// address formula <host hostname> ':' <port networkPort>
// addresses sequence address
// resource binary base64
