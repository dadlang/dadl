package parser

import (
	"errors"
	"regexp"
	"strings"
)

//Errors
var (
	ErrUnexpectedNode = errors.New("Node is not defined in schema")
)

var (
	regexIdentifierOrQuoted = "(?:[a-zA-Z0-9-_]+)|(['].*['])"
	regexIdentifier         = "[a-zA-Z0-9-_]+"

	keyWithDelegatedValueRe = regexp.MustCompile("^(?P<key>" + regexIdentifierOrQuoted + ")(\\s+(?P<rest>.*))?$")
)

type DadlSchema interface {
	getRoot() valueType
	getNode(path string, builder valueBuilder, meta parseMetadata) (valueType, valueBuilder, error)
}

type dadlSchemaImpl struct {
	root valueType
}

func (s *dadlSchemaImpl) getRoot() valueType {
	return s.root
}

func (s *dadlSchemaImpl) getNode(nodePath string, builder valueBuilder, meta parseMetadata) (valueType, valueBuilder, error) {
	node := s.root
	var err error
	pathElements := strings.Split(nodePath, ".")
	for _, pathElement := range pathElements {
		node, builder, err = node.getChild(pathElement, builder, meta)
		if err != nil {
			return nil, nil, err
		}
	}
	return node, builder, nil
}

type TokenSpec interface {
	getName() string
	isOptional() bool
	toRegex() string
	transform(value string) interface{}
}

type regexToken struct {
	name        string
	regex       string
	optional    bool
	transformer func(string) interface{}
}

func (t regexToken) getName() string {
	return t.name
}

func (t regexToken) toRegex() string {
	return t.regex
}

func (t regexToken) isOptional() bool {
	return t.optional
}

func (t regexToken) transform(value string) interface{} {
	if t.transformer != nil {
		return t.transformer(value)
	}
	return value
}

func removeQuotes(val string) string {
	if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
		return val[1 : len(val)-1]
	}
	return val
}
