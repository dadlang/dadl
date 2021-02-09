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

//DadlSchema defines dad file schema.
type DadlSchema interface {
	rootParser() NodeParser
	parser(nodePath string) (NodeParser, error)
}

//SchemaNode defines schema node
type SchemaNode interface {
	parser() NodeParser
	childNode(name string) (SchemaNode, error)
}

type dadlSchemaImpl struct {
	root SchemaNode
}

func (s *dadlSchemaImpl) rootParser() NodeParser {
	return s.root.parser()
}

func (s *dadlSchemaImpl) parser(nodePath string) (NodeParser, error) {
	node := s.root
	var err error
	pathElements := strings.Split(nodePath, ".")
	for _, pathElement := range pathElements {
		node, err = node.childNode(pathElement)
		if err != nil {
			return nil, err
		}
	}
	return node.parser(), nil
}

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

func (n *genericSchemaNode) parser() NodeParser {
	return &keyWithDelegatedValueParser{n}
}

type stringValueNode struct {
	name string
}

func (n *stringValueNode) childNode(name string) (SchemaNode, error) {
	return nil, nil
}

func (n *stringValueNode) parser() NodeParser {
	return &stringValueParser{name: n.name}
}

type keyWithDelegatedValueParser struct {
	node SchemaNode
}

func (p *keyWithDelegatedValueParser) parse(ctx *parseContext, value string) error {
	println("K(WD)V:", value)
	value = strings.TrimSpace(value)
	vals := strings.SplitN(value, " ", 2)
	if len(vals) > 1 {
		key := vals[0]
		child, err := p.node.childNode(key)
		if err != nil {
			return err
		}
		child.parser().parse(ctx, vals[1])

		// ctx.parent[vals[0]] =
	}
	return nil
}

func (p *keyWithDelegatedValueParser) childParser() (NodeParser, error) {
	return nil, nil
}

type stringValueParser struct {
	name string
}

func (p *stringValueParser) parse(ctx *parseContext, value string) error {
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
	return nil, nil
}

func (n *keyValueListNode) parser() NodeParser {
	return &genericKeyValueParser{}
}

type genericKeyValueParser struct{}

func (p *genericKeyValueParser) parse(ctx *parseContext, value string) error {
	println("KV: ", value)
	value = strings.TrimSpace(value)
	vals := strings.SplitN(value, " ", 2)
	if len(vals) > 1 {
		ctx.parent[vals[0]] = vals[1]
	}
	return nil
}

func (p *genericKeyValueParser) childParser() (NodeParser, error) {
	return nil, nil
}

type customItemListNode struct {
	name string
}

func (n *customItemListNode) childNode(name string) (SchemaNode, error) {
	return nil, nil
}

func (n *customItemListNode) parser() NodeParser {
	return &customValueParser{}
}

type customValueParser struct{}

func (p *customValueParser) parse(ctx *parseContext, value string) error {

	re := regexp.MustCompile("(?P<key>[a-zA-Z0-9-_]+)(?P<optional>[?])?\\s+(?P<type>[a-zA-Z0-9-_]+)\\s+(?:#(?P<desc>.*))?")

	parsed := matchRegexp(re, strings.TrimSpace(value))

	ctx.parent[parsed["key"]] = Node{
		"type":     parsed["type"],
		"desc":     parsed["desc"],
		"optional": flagToBool(parsed["optional"]),
	}
	return nil
}

func flagToBool(val string) bool {
	return val != ""
}

func matchRegexp(re *regexp.Regexp, value string) map[string]string {
	println("MATCH:", value)
	match := re.FindStringSubmatch(value)
	result := make(map[string]string)
	if match != nil {
		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
	}
	return result
}

func (p *customValueParser) childParser() (NodeParser, error) {
	return nil, nil
}

type identifierListNode struct {
	childType SchemaNode
}

func (n *identifierListNode) childNode(name string) (SchemaNode, error) {
	return n.childType, nil
}

func (n *identifierListNode) parser() NodeParser {
	return &genericKeyParser{childType: n.childType}
}

type genericKeyParser struct {
	childType SchemaNode
}

func (p *genericKeyParser) parse(ctx *parseContext, value string) error {
	println("K: ", value)
	key := strings.TrimSpace(value)
	v := Node{}
	ctx.parent[key] = v
	ctx.current = v
	return nil
}

func (p *genericKeyParser) childParser() (NodeParser, error) {
	return p.childType.parser(), nil
}

//ParseSchema parses schema
func ParseSchema() (DadlSchema, error) {
	return &dadlSchemaImpl{root: &genericSchemaNode{
		children: map[string]SchemaNode{
			"name":     &stringValueNode{name: "name"},
			"codename": &stringValueNode{name: "codename"},
			"global": &genericSchemaNode{
				children: map[string]SchemaNode{
					"types": &keyValueListNode{},
				},
			},
			"contexts": &identifierListNode{
				childType: &customItemListNode{},
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
