package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// typeDefs list typeDef as map[name]type

// [types]
// operation enum GET POST PUT PATCH DELETE
// typeDef sequence <name identifier> <SPACE> <type identifier> (<SPACE> '#' <desc string>)?

var typeBaseRe = regexp.MustCompile("(?P<name>" + regexIdentifier + ")(\\s+(?P<baseType>" + regexIdentifier + "))?(?P<extra>.*)")
var listTypeArgsRe = regexp.MustCompile("(?P<itemType>[a-zA-Z0-9-_]+)(\\s+as\\s+(?P<mappedType>[a-zA-Z0-9-_]+)(\\[(?P<mappedTypeArg1>[a-zA-Z0-9-_]+)\\](?P<mappedTypeArg2>[a-zA-Z0-9-_]+)?)?)?")

//GetDadlSchema returns dadl schema
func GetDadlSchema() DadlSchema {
	return &dadlSchemaImpl{root: &genericSchemaNode{
		children: map[string]SchemaNode{
			"types":     &dadlSchemaTypeNode{},
			"structure": &dadlStructureSchemaNode{},
		},
	}}
}

type dadlSchemaTypeNode struct {
}

func (n *dadlSchemaTypeNode) valueType() valueType {
	return nil
}

func (n *dadlSchemaTypeNode) childNode(name string) (SchemaNode, error) {
	return n, nil
}

func (n *dadlSchemaTypeNode) childParser() (NodeParser, error) {
	return &dadlSchemaTypeParser{}, nil
}

func (n *dadlSchemaTypeNode) isSimple() bool {
	return false
}

type dadlSchemaTypeParser struct {
}

func (p *dadlSchemaTypeParser) parse(ctx *parseContext, value string) error {
	println("[dadlSchemaTypeParser.parse]", value)

	var err error
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
	name := res["name"]
	baseType := res["baseType"]
	if baseType == "" {
		baseType = "struct"
	}
	extra := res["extra"]
	result := Node{
		"baseType": baseType,
	}
	ctx.parent[name] = result
	if baseType == "struct" {
		structure := Node{}
		result["children"] = structure
		ctx.last = structure
	} else {
		ctx.last = result
	}
	ctx.lastSchema, err = ctx.parentSchema.childNode(name)
	if err != nil {
		return err
	}

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

type dadlStructureSchemaNode struct {
	children map[string]SchemaNode
}

func (n *dadlStructureSchemaNode) valueType() valueType {
	return nil
}

func (n *dadlStructureSchemaNode) childNode(name string) (SchemaNode, error) {
	return n, nil
}

func (n *dadlStructureSchemaNode) childParser() (NodeParser, error) {
	return &dadlSchemaTypeParser{}, nil
}

func (n *dadlStructureSchemaNode) isSimple() bool {
	return false
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

func parseSchema(schemaName string, resources ResourceProvider) (DadlSchema, error) {

	if strings.HasPrefix(schemaName, "dadl ") {
		return GetDadlSchema(), nil
	}

	p := NewParser()
	file, err := resources.GetResource(schemaName)
	if err != nil {
		panic(err)
		//return err
	}
	tree, err := p.Parse(file, resources)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Parsed schema tree: %v\n", tree)

	var typesMap map[string]valueType

	if tree["types"] != nil {
		typesMap, err = buildTypes(tree["types"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	}
	fmt.Printf("Types: %v\n", typesMap)

	root, err := buildGenericNode(tree["structure"].(map[string]interface{}), typesMap)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Schema: %v\n", root)
	return &dadlSchemaImpl{root: root}, nil
}

func buildTypes(typesDef map[string]interface{}) (map[string]valueType, error) {
	resolvedTypes := map[string]valueType{}
	for key := range typesDef {
		resolveType(key, typesDef, resolvedTypes)
	}
	return resolvedTypes, nil
}

func buildType(typeDef map[string]interface{}, typesDefs map[string]interface{}, resolvedTypes map[string]valueType) (valueType, error) {
	switch typeDef["baseType"] {
	case "string":
		return &stringValue{}, nil
	case "int":
		return &intValue{}, nil
	case "enum":
		enumValue := &enumValue{values: map[string]bool{}}
		for _, value := range typeDef["values"].([]string) {
			enumValue.values[value] = true
		}
		return enumValue, nil
	case "formula":
		return &formulaValue{}, nil
	case "sequence":
		return &sequenceValue{}, nil
	case "binary":
		return &binaryValue{}, nil
	}
	return nil, errors.New("Unknown type: " + typeDef["baseType"].(string))
}

func resolveType(typeName string, typesDefs map[string]interface{}, resolvedTypes map[string]valueType) (valueType, error) {
	if resolved, ok := resolvedTypes[typeName]; ok {
		return resolved, nil
	}
	var err error

	typeDef := typesDefs[typeName].(map[string]interface{})
	if typeDef == nil {
		return nil, errors.New("Unknown type: " + typeName)
	}
	resolvedTypes[typeName], err = buildType(typeDef, typesDefs, resolvedTypes)
	if err != nil {
		return nil, err
	}
	return resolvedTypes[typeName], nil
}

func buildGenericNode(children map[string]interface{}, typesMap map[string]valueType) (SchemaNode, error) {
	// baseTypes := []string{"string", "int", "enum", "struct"}
	node := &genericSchemaNode{children: map[string]SchemaNode{}}
	var err error

	for key, value := range children {
		value := value.(map[string]interface{})
		baseType := value["baseType"].(string)

		if baseType == "struct" {
			child, err := buildGenericNode(value["children"].(map[string]interface{}), typesMap)
			if err != nil {
				return nil, err
			}
			node.children[key] = child
		} else {

			valueType, ok := typesMap[baseType]
			if !ok {
				valueType, err = buildType(value, nil, typesMap)
				if err != nil {
					return nil, err
				}
			}
			_, isString := valueType.(*stringValue)
			if isString {
				node.children[key] = &stringValueNode{name: key}
			} else {
				node.children[key] = &simpleValueLeafNode{name: key, valueType: valueType}
			}
		}
		// else if baseType == "string" {
		// 	node.children[key] = &stringValueNode{name: key}
		// } else if baseType == "int" {
		// 	node.children[key] = &intValueNode{name: key}
		// }
		// else if baseType == "enum" {
		// 	node.children[key] = &enumValueNode{name: key}
		// }
	}
	return node, nil
}
