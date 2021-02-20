package parser

import "errors"

type stringValueNode struct {
	name   string
	indent int
}

func (n *stringValueNode) valueType() valueType {
	return nil
}

func (n *stringValueNode) childNode(name string) (SchemaNode, error) {
	return n, nil
}

func (n *stringValueNode) childParser() (NodeParser, error) {
	return &stringValueParser{name: n.name, indent: &n.indent}, nil
}

func (n *stringValueNode) isSimple() bool {
	return true
}

type intValueNode struct {
	name string
}

func (n *intValueNode) valueType() valueType {
	return nil
}

func (n *intValueNode) childNode(name string) (SchemaNode, error) {
	return nil, nil
}

func (n *intValueNode) childParser() (NodeParser, error) {
	return &intValueParser{name: n.name}, nil
}

func (n *intValueNode) isSimple() bool {
	return true
}

type simpleValueLeafNode struct {
	name      string
	valueType valueType
}

func (n *simpleValueLeafNode) childNode(name string) (SchemaNode, error) {
	return nil, errors.New("Children not supported")
}

func (n *simpleValueLeafNode) childParser() (NodeParser, error) {
	return &simpleValueParser{name: n.name, valueType: n.valueType}, nil
}

func (n *simpleValueLeafNode) isSimple() bool {
	return true
}

type simpleValueParser struct {
	name      string
	valueType valueType
}

func (p *simpleValueParser) parse(ctx *parseContext, value string) error {
	// println("[simpleValueParser.parse]", value)
	parsed, err := p.valueType.parse(value)
	if err != nil {
		return err
	}
	ctx.parent[p.name] = parsed
	ctx.last = ctx.parent
	ctx.lastSchema = ctx.parentSchema
	return nil
}

func (p *simpleValueParser) childParser() (NodeParser, error) {
	return nil, nil
}
