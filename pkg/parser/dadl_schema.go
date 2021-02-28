package parser

import (
	"errors"
	"regexp"
	"strings"
)

var typeBaseRe = regexp.MustCompile("(?P<name>" + regexIdentifier + ")(\\s+(?P<baseType>" + regexIdentifier + "))?(?P<extra>.*)")
var listTypeArgsRe = regexp.MustCompile("(?P<itemType>[a-zA-Z0-9-_]+)(\\s+as\\s+(?P<mappedType>[a-zA-Z0-9-_]+)(\\[(?P<mappedTypeArg1>[a-zA-Z0-9-_]+)\\](?P<mappedTypeArg2>[a-zA-Z0-9-_]+)?)?)?")
var formulaTypeArgsRe = regexp.MustCompile("(?:\\<(?P<name>" + regexIdentifier + ")\\s+(?P<baseType>" + regexIdentifier + ")\\>)|(?:\\'(?P<literal>.*?)\\')")
var mapTypeArgsRe = regexp.MustCompile("\\[(?P<keyType>" + regexIdentifier + ")\\](?P<valueType>" + regexIdentifier + ")?")

func GetDadlSchema() DadlSchema {
	return &dadlSchemaImpl{root: &structValue{
		children: map[string]valueType{
			"types":     &dadlCustomTypeValue{},
			"structure": &dadlCustomTypeValue{},
		},
	}}
}

type dadlCustomTypeValue struct {
}

func (v *dadlCustomTypeValue) parse(builder valueBuilder, value string) error {
	// println("dadlCustomTypeValue [parse]:", value)
	builder.setValue(Node{})
	return nil
}

func (v *dadlCustomTypeValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {

	// println("dadlCustomTypeValue [parseChild]:", value)
	// var err error
	res := map[string]string{}
	match := typeBaseRe.FindStringSubmatch(strings.TrimSpace(value))
	// var keyValue string
	if match != nil {
		for i, name := range typeBaseRe.SubexpNames() {
			if i != 0 && name != "" {
				res[name] = match[i]
			}
		}
	} else {
		return nil, errors.New("Ivalid type definition")
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
	builder.getValue().(Node)[name] = result

	var def Node
	var val Node
	if baseType == "struct" {
		structure := Node{}
		result["children"] = structure
		def = structure
		val = structure
	} else if baseType == "map" {
		structure := Node{}
		valueType := Node{}
		valueType["children"] = structure
		result["valueType"] = valueType
		def = result
		val = structure
	} else {
		def = result
		val = result
	}

	if baseType == "string" {
		parseStringArgs(def, strings.TrimSpace(extra), "regex")
	} else if baseType == "enum" {
		parseEnumArgs(def, strings.TrimSpace(extra), "values")
	} else if baseType == "oneof" {
		parseOneofArgs(def, strings.TrimSpace(extra), "options")
	} else if baseType == "list" {
		parseListTypeArgs(def, strings.TrimSpace(extra), "values")
	} else if baseType == "formula" {
		parseFormulaArgs(def, strings.TrimSpace(extra), "formula")
	} else if baseType == "sequence" {
		parseSequenceArgs(def, strings.TrimSpace(extra), "sequence")
	} else if baseType == "map" {
		parseMapTypeArgs(def, strings.TrimSpace(extra), "map")
	}

	return &nodeInfo{
		valueType: v,
		builder: &delegatedValueBuilder{
			get: func() interface{} {
				return val
			},
			set: func(interface{}) {
				panic("Not supported")
			},
		},
	}, nil
}

func (v *dadlCustomTypeValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("not supported")
}

func (v *dadlCustomTypeValue) toRegex() string {
	return ".*"
}

func parseStringArgs(def Node, value string, key string) error {
	def[key] = strings.TrimSpace(value)
	return nil
}

func parseEnumArgs(def Node, value string, key string) error {
	def[key] = strings.Split(value, " ")
	return nil
}

func parseOneofArgs(def Node, value string, key string) error {
	def[key] = strings.Split(value, " ")
	return nil
}

func parseSequenceArgs(def Node, value string, key string) error {
	def[key] = map[string]string{"itemType": strings.TrimSpace(value)}
	return nil
}

func parseFormulaArgs(def Node, value string, key string) error {
	matches := formulaTypeArgsRe.FindAllStringSubmatch(strings.TrimSpace(value), -1)
	tokens := []map[string]interface{}{}
	for _, match := range matches {
		if match[1] != "" && match[2] != "" {
			tokens = append(tokens, map[string]interface{}{
				"type":     "token",
				"name":     match[1],
				"baseType": match[2],
			})
		} else if match[3] != "" {
			tokens = append(tokens, map[string]interface{}{
				"type":  "constant",
				"value": match[3],
			})
		}
	}
	def[key] = tokens
	return nil
}

func parseListTypeArgs(def Node, value string, key string) error {
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
	def["itemType"] = Node{"baseType": res["itemType"]}
	if res["mappedType"] != "" {
		mapped := map[string]string{}
		mapped["mappedType"] = res["mappedType"]
		mapped["mappedTypeArg1"] = res["mappedTypeArg1"]
		mapped["mappedTypeArg2"] = res["mappedTypeArg2"]
		def["mapped"] = mapped
	}
	return nil
}

func parseMapTypeArgs(def Node, value string, key string) error {
	match := mapTypeArgsRe.FindStringSubmatch(strings.TrimSpace(value))
	// var keyValue string
	res := map[string]string{}
	if match != nil {
		for i, name := range mapTypeArgsRe.SubexpNames() {
			if i != 0 && name != "" {
				res[name] = match[i]
			}
		}
	}
	def["keyType"] = Node{"baseType": res["keyType"]}
	valueType := res["valueType"]
	if valueType == "" {
		valueType = "struct"
	}
	def["valueType"].(Node)["baseType"] = valueType
	return nil
}

func parseSchema2(schemaName string, resources ResourceProvider) (DadlSchema, error) {

	if schemaName == "dadl" {
		return GetDadlSchema(), nil
	}

	p := NewParser()
	file, err := resources.GetResource(schemaName)
	if err != nil {
		return nil, err
	}
	tree, err := p.Parse(file, resources)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("Parsed schema tree: %v\n", tree)

	var typesDefs map[string]interface{}
	if tree["types"] != nil {
		typesDefs = tree["types"].(map[string]interface{})
	} else {
		typesDefs = map[string]interface{}{}
	}
	resolver := newResolver(typesDefs)

	root, err := resolver.buildType(map[string]interface{}{
		"baseType": "struct",
		"children": tree["structure"],
	})

	if err != nil {
		return nil, err
	}
	// fmt.Printf("Schema: %+v\n", root)
	return &dadlSchemaImpl{root: root}, nil
}
