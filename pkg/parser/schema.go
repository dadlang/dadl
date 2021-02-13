package parser

import (
	"errors"
	"fmt"
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

	keyWithDelegatedValueRe = regexp.MustCompile("(?P<key>" + regexIdentifierOrQuoted + ")(\\s+(?P<rest>.*))?")
)

//DadlSchema defines dad file schema.
type DadlSchema interface {
	getRoot() SchemaNode
	getNode(path string) (SchemaNode, error)
	//parser(nodePath string) (NodeParser, error)
}

//SchemaNode defines schema node
type SchemaNode interface {
	childParser() (NodeParser, error)
	childNode(name string) (SchemaNode, error)
}

type dadlSchemaImpl struct {
	root SchemaNode
}

func (s *dadlSchemaImpl) getRoot() SchemaNode {
	return s.root
}

func (s *dadlSchemaImpl) getNode(nodePath string) (SchemaNode, error) {
	node := s.root
	var err error
	pathElements := strings.Split(nodePath, ".")
	for _, pathElement := range pathElements {
		node, err = node.childNode(pathElement)
		if err != nil {
			return nil, err
		}
	}
	return node, nil
}

// func (s *dadlSchemaImpl) parser(nodePath string) (NodeParser, error) {
// 	node := s.root
// 	var err error
// 	pathElements := strings.Split(nodePath, ".")
// 	for _, pathElement := range pathElements {
// 		node, err = node.childNode(pathElement)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return node.childParser(), nil
// }

type genericSchemaNode struct {
	children map[string]SchemaNode
}

func (n *genericSchemaNode) childNode(name string) (SchemaNode, error) {
	val, ok := n.children[name]
	if !ok {
		return nil, ErrUnexpectedNode
	}
	return val, nil
}

func (n *genericSchemaNode) childParser() (NodeParser, error) {
	return &keyWithDelegatedValueParser{}, nil
}

type stringValueNode struct {
	name string
}

func (n *stringValueNode) childNode(name string) (SchemaNode, error) {
	return nil, nil
}

func (n *stringValueNode) childParser() (NodeParser, error) {
	return &stringValueParser{name: n.name}, nil
}

type keyWithDelegatedValueParser struct{}

func (p *keyWithDelegatedValueParser) parse(ctx *parseContext, value string) error {
	//println("[keyWithDelegatedValueParser.parse]", value)

	res := map[string]string{}
	match := keyWithDelegatedValueRe.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		return errors.New("invalid format")
	}
	if match != nil {
		for i, name := range keyWithDelegatedValueRe.SubexpNames() {
			if i != 0 && name != "" {
				res[name] = match[i]
			}
		}
	}

	fmt.Printf("MATCH %v\n", res)
	key := removeQuotes(res["key"])
	child, err := ctx.parentSchema.childNode(key)
	if err != nil {
		return err
	}
	ctx.lastSchema = child
	if res["rest"] != "" {
		childParser, err := child.childParser()
		if err != nil {
			return nil
		}
		childParser.parse(ctx, res["rest"])
	} else {
		//fmt.Printf("   - parent: %x\n", ctx.parent)
		//	println("OK")
		ctx.parent[key] = Node{}
		ctx.last = ctx.parent[key].(Node)
	}
	return nil
}

func (p *keyWithDelegatedValueParser) childParser() (NodeParser, error) {
	return &keyWithDelegatedValueParser{}, nil
}

type stringValueParser struct {
	name string
}

func (p *stringValueParser) parse(ctx *parseContext, value string) error {
	//println("[stringValueParser.parse]", value)
	ctx.parent[p.name] = value
	return nil
}

func (p *stringValueParser) childParser() (NodeParser, error) {
	return nil, nil
}

type keyValueListNode struct {
	name string
}

func (n *keyValueListNode) childNode(name string) (SchemaNode, error) {
	//println("ERRR:", name)
	return &keyValueListNode{name: name}, nil
}

func (n *keyValueListNode) childParser() (NodeParser, error) {
	return &genericKeyValueParser{node: n}, nil
}

type genericKeyValueParser struct {
	node SchemaNode
}

func (p *genericKeyValueParser) parse(ctx *parseContext, value string) error {
	//	println("[genericKeyValueParser.parse]", value)
	value = strings.TrimSpace(value)
	vals := strings.SplitN(value, " ", 2)
	if len(vals) > 1 {
		ctx.parent[vals[0]] = vals[1]
	} else {
		//	println("ERRRRRRRRRRRRR")
	}
	return nil
}

func (p *genericKeyValueParser) childParser() (NodeParser, error) {
	return &genericKeyValueParser{node: p.node}, nil
}

// type customItemListNode struct {
// 	name string
// }

// func (n *customItemListNode) childNode(name string) (SchemaNode, error) {
// 	return nil, nil
// }

// func (n *customItemListNode) parser() NodeParser {
// 	return &customValueParser{}
// }

// type customValueParser struct{}

// func (p *customValueParser) parse(ctx *parseContext, value string) error {
// 	println("[customValueParser.parse]", value)

// 	re := regexp.MustCompile("(?P<key>[a-zA-Z0-9-_]+)(?P<optional>[?])?\\s+(?P<type>[a-zA-Z0-9-_]+)\\s+(?:#(?P<desc>.*))?")

// 	parsed := matchRegexp(re, strings.TrimSpace(value))

// 	ctx.parent[parsed["key"]] = Node{
// 		"type":     parsed["type"],
// 		"desc":     parsed["desc"],
// 		"optional": flagToBool(parsed["optional"]),
// 	}
// 	return nil
// }

// func flagToBool(val string) bool {
// 	return val != ""
// }

// func matchRegexp(re *regexp.Regexp, value string) map[string]string {
// 	println("MATCH:", value)
// 	match := re.FindStringSubmatch(value)
// 	result := make(map[string]string)
// 	if match != nil {
// 		for i, name := range re.SubexpNames() {
// 			if i != 0 && name != "" {
// 				result[name] = match[i]
// 			}
// 		}
// 	}
// 	return result
// }

// func (p *customValueParser) childParser() (NodeParser, error) {
// 	return nil, nil
// }

type childListOnlyNode struct {
	childType SchemaNode
}

func (n *childListOnlyNode) childNode(name string) (SchemaNode, error) {
	return n.childType, nil
}

func (n *childListOnlyNode) childParser() (NodeParser, error) {
	return n.childType.childParser()
}

type identifierListNode struct {
	childType SchemaNode
}

func (n *identifierListNode) childNode(name string) (SchemaNode, error) {
	return n.childType, nil
}

func (n *identifierListNode) childParser() (NodeParser, error) {
	return &genericKeyParser{childType: n.childType}, nil
}

type genericKeyParser struct {
	childType SchemaNode
}

func (p *genericKeyParser) parse(ctx *parseContext, value string) error {
	//	println("[genericKeyParser.parse]", value)
	key := strings.TrimSpace(value)
	v := Node{}
	ctx.parent[key] = v
	ctx.last = v
	ctx.lastSchema = p.childType
	return nil
}

func (p *genericKeyParser) childParser() (NodeParser, error) {
	return p.childType.childParser()
}

//ParseSchema parses schema
func ParseSchema() (DadlSchema, error) {
	return &dadlSchemaImpl{root: &genericSchemaNode{
		children: map[string]SchemaNode{
			"name":     &stringValueNode{name: "name"},
			"codename": &stringValueNode{name: "codename"},
			"global": &genericSchemaNode{
				children: map[string]SchemaNode{
					"types": &childListOnlyNode{
						childType: &customTokensNode{
							keyTokenName: "name",
							tokens: []TokenSpec{
								regexToken{name: "name", regex: "[a-zA-Z0-9-_]+"},
								regexToken{regex: "\\s+", optional: false},
								regexToken{name: "baseType", regex: "[a-zA-Z0-9-_]+"},
							},
						},
					},
				},
			},
			"modules": &childListOnlyNode{
				childType: &genericSchemaNode{
					children: map[string]SchemaNode{
						"name":      &stringValueNode{name: "name"},
						"namespace": &stringValueNode{name: "namespace"},
						"types": &childListOnlyNode{
							childType: &customTokensNode{
								keyTokenName: "name",
								tokens: []TokenSpec{
									regexToken{name: "name", regex: "[a-zA-Z0-9-_]+"},
									regexToken{regex: "\\s+", optional: false},
									regexToken{name: "baseType", regex: "[a-zA-Z0-9-_]+"},
								},
							},
						},
					},
				},
			},
			"contexts": &identifierListNode{
				childType: &customTokensNode{
					keyTokenName: "name",
					tokens: []TokenSpec{
						regexToken{name: "name", regex: "[a-zA-Z0-9-_]+"},
						regexToken{name: "optional", regex: "[?]", optional: true, transformer: func(val string) interface{} { return val != "" }},
						regexToken{regex: "\\s+"},
						regexToken{name: "type", regex: "[a-zA-Z0-9-_]+"},
						regexToken{regex: "\\s+"},
						regexToken{name: "desc", regex: "#.+", transformer: func(val string) interface{} { return val[1:] }},
					},
				},
			},
			//,
		},
	}}, nil
}

// name Online Boutique
// codename boutique

// [modules.cart << ./modules/cart.dad]

// [global.types]
// UserId String
// ProductId String

// [contexts]
// user
//     userID? UserID #Identifier of the user
// user2
//     userID2_1 UserID #Identifier of the user
//     userID2_2? UserID #Identifier of the user
// user3
//     userID3_1? UserID #Identifier of the user

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

type customTokensNode struct {
	tokens       []TokenSpec
	keyTokenName string
}

func (n *customTokensNode) childNode(name string) (SchemaNode, error) {
	panic("NOT IMPLEMENTED")
}

func (n *customTokensNode) childParser() (NodeParser, error) {
	return newCustomTokensNodeParser(n.tokens, n.keyTokenName), nil
}

func newCustomTokensNodeParser(tokens []TokenSpec, keyTokenName string) NodeParser {
	var sb strings.Builder
	transformers := make(map[string]func(string) interface{})
	for _, token := range tokens {
		if token.getName() != "" {
			sb.WriteString("(?P<")
			sb.WriteString(token.getName())
			sb.WriteString(">")
			transformers[token.getName()] = token.transform
		} else {
			sb.WriteString("(?:")
		}
		sb.WriteString(token.toRegex())
		sb.WriteString(")")
		if token.isOptional() {
			sb.WriteString("?")
		}
	}
	return &customTokensNodeParser{re: regexp.MustCompile(sb.String()), transformers: transformers, keyTokenName: keyTokenName}
}

type customTokensNodeParser struct {
	re           *regexp.Regexp
	keyTokenName string
	transformers map[string]func(string) interface{}
}

func (p *customTokensNodeParser) parse(ctx *parseContext, value string) error {
	//println("[customTokensNodeParser.parse]", value)

	res := Node{}
	match := p.re.FindStringSubmatch(strings.TrimSpace(value))
	var keyValue string
	if match != nil {
		for i, name := range p.re.SubexpNames() {
			if i != 0 && name != "" {
				value := p.transformers[name](match[i])
				if name == p.keyTokenName {
					keyValue = value.(string)
				} else {
					res[name] = value
				}
			}
		}
	}
	ctx.parent[keyValue] = res
	return nil
}

func (p *customTokensNodeParser) childParser() (NodeParser, error) {
	return nil, nil
}

//GetDadlSchema returns dadl schema
func GetDadlSchema() DadlSchema {
	return &dadlSchemaImpl{root: &genericSchemaNode{
		children: map[string]SchemaNode{
			"types": &dadlSchemaTypeNode{},
			"structure": &genericSchemaNode{
				children: map[string]SchemaNode{
					"long name": &stringValueNode{"long name"},
				},
			},
		},
	}}
}

/////////////

type dadlSchemaTypeNode struct {
}

func (n *dadlSchemaTypeNode) childNode(name string) (SchemaNode, error) {
	panic("NOT IMPLEMENTED")
}

func (n *dadlSchemaTypeNode) childParser() (NodeParser, error) {
	return &dadlSchemaTypeParser{}, nil
}

var typeBaseRe = regexp.MustCompile("(?P<name>" + regexIdentifier + ")\\s+(?P<baseType>" + regexIdentifier + ")(?P<extra>.*)")
var listTypeArgsRe = regexp.MustCompile("(?P<itemType>[a-zA-Z0-9-_]+)(\\s+as\\s+(?P<mappedType>[a-zA-Z0-9-_]+)(\\[(?P<mappedTypeArg1>[a-zA-Z0-9-_]+)\\](?P<mappedTypeArg2>[a-zA-Z0-9-_]+)?)?)?")

// typeDefs list typeDef as map[name]type

type dadlSchemaTypeParser struct {
}

func (p *dadlSchemaTypeParser) parse(ctx *parseContext, value string) error {
	println("[dadlSchemaTypeParser.parse]", value)

	res := map[string]string{}
	match := typeBaseRe.FindStringSubmatch(strings.TrimSpace(value))
	// var keyValue string
	if match != nil {
		for i, name := range typeBaseRe.SubexpNames() {
			if i != 0 && name != "" {
				res[name] = match[i]
			}
		}
	}
	baseType := res["baseType"]
	extra := res["extra"]
	result := Node{
		"baseType": baseType,
	}
	ctx.parent[res["name"]] = result

	if baseType == "enum" {
		parseEnumArgs(&parseContext{
			parentSchema: ctx.parentSchema,
			lastSchema:   ctx.lastSchema,
			parent:       result,
			last:         result,
		}, strings.TrimSpace(extra), "values")
	} else if baseType == "list" {
		parseListTypeArgs(&parseContext{
			parentSchema: ctx.parentSchema,
			lastSchema:   ctx.lastSchema,
			parent:       result,
			last:         result,
		}, strings.TrimSpace(extra), "values")
	}

	return nil
}

func (p *dadlSchemaTypeParser) childParser() (NodeParser, error) {
	panic("NOT IMPLEMENTED")
}

func parseEnumArgs(ctx *parseContext, value string, key string) error {
	ctx.parent[key] = strings.Split(value, " ")
	return nil
}

func parseListTypeArgs(ctx *parseContext, value string, key string) error {
	match := listTypeArgsRe.FindStringSubmatch(strings.TrimSpace(value))
	// var keyValue string
	res := map[string]string{}
	if match != nil {
		for i, name := range listTypeArgsRe.SubexpNames() {
			if i != 0 && name != "" {
				res[name] = match[i]
			}
		}
	}
	ctx.parent["itemType"] = res["itemType"]
	if res["mappedType"] != "" {
		mapped := map[string]string{}
		mapped["mappedType"] = res["mappedType"]
		mapped["mappedTypeArg1"] = res["mappedTypeArg1"]
		mapped["mappedTypeArg2"] = res["mappedTypeArg2"]
		ctx.parent["mapped"] = mapped
	}
	return nil
}

// [types]
// operation enum GET POST PUT PATCH DELETE
// typeDef sequence <name identifier> <SPACE> <type identifier> (<SPACE> '#' <desc string>)?

func removeQuotes(val string) string {
	if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
		return val[1 : len(val)-1]
	}
	return val
}