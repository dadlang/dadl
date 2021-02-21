package parser

import (
	"errors"
	"regexp"
	"strconv"
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
}

//SchemaNode defines schema node
type SchemaNode interface {
	childParser() (NodeParser, error)
	childNode(name string) (SchemaNode, error)
	isSimple() bool
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

func (n *genericSchemaNode) valueType() valueType {
	return nil
}

func (n *genericSchemaNode) childNode(name string) (SchemaNode, error) {
	val, ok := n.children[name]
	if !ok {
		println("Not found node: ", name)
		return nil, ErrUnexpectedNode
	}
	return val, nil
}

func (n *genericSchemaNode) childParser() (NodeParser, error) {
	return &keyWithDelegatedValueParser{}, nil
}

func (n *genericSchemaNode) isSimple() bool {
	return false
}

type genericMapNode struct {
	value SchemaNode
}

func (n *genericMapNode) childNode(name string) (SchemaNode, error) {
	return n.value, nil
}

func (n *genericMapNode) childParser() (NodeParser, error) {
	return &keyWithDelegatedValueParser{}, nil
}

func (n *genericMapNode) isSimple() bool {
	return false
}

type keyWithDelegatedValueParser struct{}

func (p *keyWithDelegatedValueParser) parse(ctx *parseContext, value string) error {
	// println("[keyWithDelegatedValueParser.parse]", value)

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

	key := removeQuotes(res["key"])
	child, err := ctx.parentSchema.childNode(key)
	if err != nil {
		return err
	}
	ctx.lastSchema = child
	ctx.last = nil
	if !child.isSimple() {
		node := Node{}
		ctx.parent[key] = node
		ctx.last = node
	}
	if res["rest"] != "" {
		childParser, err := child.childParser()
		if err != nil {
			return nil
		}
		childParser.parse(ctx, res["rest"])
	}
	// else {
	// 	ctx.last = ctx.parent
	// }
	return nil
}

func (p *keyWithDelegatedValueParser) childParser() (NodeParser, error) {
	return &keyWithDelegatedValueParser{}, nil
}

type stringValueParser struct {
	name   string
	indent *int
}

func (p *stringValueParser) parse(ctx *parseContext, value string) error {
	//println("[stringValueParser.parse]", value)
	if *p.indent == 0 {
		*p.indent = calcIndentWeight(value)
	}
	value = value[*p.indent:]
	if ctx.parent[p.name] == nil {
		//println("SET[", p.name, "]", value)
		ctx.parent[p.name] = value
	} else {
		// println("APPEND[", p.name, "]", value)
		ctx.parent[p.name] = ctx.parent[p.name].(string) + "\n" + value
	}
	ctx.last = ctx.parent
	ctx.lastSchema = ctx.parentSchema
	return nil
}

func (p *stringValueParser) childParser() (NodeParser, error) {
	return nil, nil
}

type intValueParser struct {
	name string
}

func (p *intValueParser) parse(ctx *parseContext, value string) error {
	//println("[intValueParser.parse]", value)
	var err error
	ctx.parent[p.name], err = strconv.Atoi(value)
	if err != nil {
		return err
	}
	return nil
}

func (p *intValueParser) childParser() (NodeParser, error) {
	return nil, nil
}

// type keyValueListNode struct {
// 	name string
// }

// func (n *keyValueListNode) childNode(name string) (SchemaNode, error) {
// 	return &keyValueListNode{name: name}, nil
// }

// func (n *keyValueListNode) childParser() (NodeParser, error) {
// 	return &genericKeyValueParser{node: n}, nil
// }

// func (n *keyValueListNode) isSimple() bool {
// 	return true
// }

type genericKeyValueParser struct {
	node SchemaNode
}

func (p *genericKeyValueParser) parse(ctx *parseContext, value string) error {
	// println("[genericKeyValueParser.parse]", value)
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

func (p *genericKeyValueParser) isSimple() bool {
	return false
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

func (n *childListOnlyNode) valueType() valueType {
	return nil
}

func (n *childListOnlyNode) childNode(name string) (SchemaNode, error) {
	return n.childType, nil
}

func (n *childListOnlyNode) childParser() (NodeParser, error) {
	return n.childType.childParser()
}

func (n *childListOnlyNode) isSimple() bool {
	return false
}

type identifierListNode struct {
	childType SchemaNode
}

func (n *identifierListNode) valueType() valueType {
	return nil
}

func (n *identifierListNode) childNode(name string) (SchemaNode, error) {
	return n.childType, nil
}

func (n *identifierListNode) childParser() (NodeParser, error) {
	return &genericKeyParser{childType: n.childType}, nil
}

func (n *identifierListNode) isSimple() bool {
	return false
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

func (n *customTokensNode) valueType() valueType {
	return nil
}

func (n *customTokensNode) childNode(name string) (SchemaNode, error) {
	panic("NOT IMPLEMENTED")
}

func (n *customTokensNode) childParser() (NodeParser, error) {
	return newCustomTokensNodeParser(n.tokens, n.keyTokenName), nil
}

func (n *customTokensNode) isSimple() bool {
	return false
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

func removeQuotes(val string) string {
	if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
		return val[1 : len(val)-1]
	}
	return val
}
