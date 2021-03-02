package parser

import (
	"errors"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
)

var typeBaseRe = regexp.MustCompile("(?P<name>" + regexIdentifier + ")(\\s+(?P<baseType>" + regexIdentifier + "))?(?P<extra>.*)")
var listTypeArgsRe = regexp.MustCompile("(?P<itemType>[a-zA-Z0-9-_]+)(\\s+as\\s+(?P<mappedType>[a-zA-Z0-9-_]+)(\\[(?P<mappedTypeArg1>[a-zA-Z0-9-_]+)\\](?P<mappedTypeArg2>[a-zA-Z0-9-_]+)?)?)?")
var formulaTypeArgsRe = regexp.MustCompile("(?:\\<(?P<name>" + regexIdentifier + ")\\s+(?P<baseType>" + regexIdentifier + ")\\>)|(?:\\'(?P<literal>.*?)\\')")
var mapTypeArgsRe = regexp.MustCompile("\\[(?P<keyType>" + regexIdentifier + ")\\](?P<valueType>" + regexIdentifier + ")?")

func GetDadlSchema() DadlSchema {
	keyType := &stringValue{regex: "[A-Za-z0-9-_]+"}
	whitespace := &stringValue{regex: "\\s+"}
	typeDef := &oneofValue{}
	// typeDef := &formulaValue{
	// 	formula: []formulaItem{
	// 		{
	// 			name:      "typeDef",
	// 			valueType: anyType,
	// 		},
	// 		{
	// 			optional:  true,
	// 			composite: true,
	// 			children: []formulaItem{
	// 				{
	// 					valueType: &stringValue{regex: "\\s*#"},
	// 				},
	// 				{
	// 					name:      "comment",
	// 					valueType: &stringValue{},
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	structType := &complexValue{
		textValue: &formulaValue{
			formula: []formulaItem{
				{
					valueType: &stringValue{regex: "(?:struct)|"},
				},
			},
		},
		structValue: &mapValue{
			keyType:   keyType,
			valueType: typeDef,
		},
	}
	stringDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "string"},
			},
			{
				optional:  true,
				composite: true,
				children: []formulaItem{
					{
						valueType: whitespace,
					},
					{
						name:      "regex",
						valueType: &stringValue{regex: "\\S+"},
					},
				},
			},
		},
	}
	identifierDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "identifier"},
			},
		},
	}
	intDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "int"},
			},
			{
				optional:  true,
				composite: true,
				children: []formulaItem{
					{
						valueType: whitespace,
					},
					{
						valueType: &stringValue{regex: "-?[0-9]+\\.\\.-?[0-9]+"},
					},
				},
			},
			//TODO move comment somwhere else
			{
				optional:  true,
				composite: true,
				children: []formulaItem{
					{
						valueType: whitespace,
					},
					{
						valueType: &constantValue{value: "#"},
					},
					{
						valueType: &stringValue{},
					},
				},
			},
		},
	}
	boolDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "bool"},
			},
		},
	}
	numberDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "number"},
			},
		},
	}
	enumDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "enum"},
			},
			{
				valueType: whitespace,
			},
			{
				valueType: &sequenceValue{
					itemType: &stringValue{regex: regexIdentifier},
				},
			},
		},
	}
	formulaItemDefSequenceDelegated := &delegatedValue{}
	formulaItemDef := &oneofValue{
		options: []oneofValueOption{
			{
				name:      "formulaItemConstant",
				valueType: &stringValue{regex: "'.*?'"},
			},
			{
				name: "formulaItemVariable",
				valueType: &formulaValue{
					formula: []formulaItem{

						{
							valueType: &constantValue{value: "<"},
						},
						{
							name:      "name",
							valueType: &stringValue{regex: regexIdentifier},
						},
						{
							valueType: whitespace,
						},
						{
							name:      "type",
							valueType: &delegatedValue{target: typeDef},
						},
						{
							valueType: &constantValue{value: ">"},
						},
					},
				},
			},
			{
				name: "formulaItemOptional",
				valueType: &formulaValue{
					formula: []formulaItem{

						{
							valueType: &constantValue{value: "["},
						},
						{
							name:      "items",
							valueType: formulaItemDefSequenceDelegated,
						},
						{
							valueType: &constantValue{value: "]"},
						},
					},
				},
			},
		},
	}
	formulaItemDefSequenceDelegated.target = &sequenceValue{itemType: formulaItemDef}
	formulaDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "formula"},
			},
			{
				valueType: whitespace,
			},
			{
				name: "items",
				valueType: &sequenceValue{
					itemType: formulaItemDef,
				},
			},
		},
	}
	sequenceDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "sequence"},
			},
			{
				valueType: whitespace,
			},
			{
				name:      "itemType",
				valueType: &delegatedValue{target: typeDef},
			},
		},
	}
	listDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "list"},
			},
			{
				valueType: whitespace,
			},
			{
				name:      "itemType",
				valueType: &delegatedValue{target: typeDef},
			},
		},
	}
	mapDef := &complexValue{
		textValue: &formulaValue{
			formula: []formulaItem{
				{
					valueType: &stringValue{regex: "map"},
				},
				{
					valueType: &constantValue{value: "["},
				},
				{
					name:      "keyType",
					valueType: &delegatedValue{target: typeDef},
				},
				{
					valueType: &constantValue{value: "]"},
				},
				{
					optional:  true,
					name:      "valueType",
					valueType: &delegatedValue{target: typeDef},
				},
			},
		},
		structValue: &mapValue{
			keyType:   keyType,
			valueType: typeDef,
		},
	}
	customTypeRef := &formulaValue{
		formula: []formulaItem{
			{
				name:      "typeName",
				valueType: &stringValue{regex: regexIdentifier},
			},
		},
	}
	oneofDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "oneof"},
			},
			{
				valueType: whitespace,
			},
			{
				name: "options",
				valueType: &sequenceValue{
					itemType: &stringValue{regex: regexIdentifier},
				},
			},
		},
	}
	complexDef := &complexValue{
		textValue: &formulaValue{
			formula: []formulaItem{
				{
					valueType: &stringValue{regex: "complex"},
				},
			},
		},
		structValue: &structValue{
			children: map[string]valueType{
				"text":      typeDef,
				"structure": typeDef,
			},
		},
	}
	typeDef.options = []oneofValueOption{
		{
			name:      "stringDef",
			valueType: stringDef,
		},
		{
			name:      "identifierDef",
			valueType: identifierDef,
		},
		{
			name:      "intDef",
			valueType: intDef,
		},
		{
			name:      "boolDef",
			valueType: boolDef,
		},
		{
			name:      "numberDef",
			valueType: numberDef,
		},
		{
			name:      "enumDef",
			valueType: enumDef,
		},
		{
			name:      "formulaDef",
			valueType: formulaDef,
		},
		{
			name:      "sequenceDef",
			valueType: sequenceDef,
		},
		{
			name:      "listDef",
			valueType: listDef,
		},
		{
			name:      "mapDef",
			valueType: mapDef,
		},
		{
			name:      "structDef",
			valueType: structType,
		},
		{
			name:      "oneofDef",
			valueType: oneofDef,
		},
		{
			name:      "complexDef",
			valueType: complexDef,
		},
		{
			name:      "customTypeRef",
			valueType: customTypeRef,
		},
	}
	typesDef := &mapValue{
		keyType:   keyType,
		valueType: typeDef,
	}

	return &dadlSchemaImpl{root: &structValue{
		children: map[string]valueType{
			"types":     typesDef,
			"structure": typesDef,
		},
	}}
}

func decodeHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if to.String() == "parser.abstractTypeDef" {
		return mapType(data.(map[string]interface{}))
	} else if to.String() == "parser.abstractFormulaItem" {
		return mapFormulaItem(data.(map[string]interface{}))
	}
	return data, nil
}

func mapFormulaItem(data map[string]interface{}) (abstractFormulaItem, error) {
	option := data["@type"].(string)
	var result interface{}
	switch option {
	case "formulaItemVariable":
		value := data["value"].(map[string]interface{})
		mappedType, err := mapType(value["type"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		result = &formulaItemVariable{
			Name: value["name"].(string),
			Type: mappedType,
		}
	case "formulaItemConstant":
		result = &formulaItemConstant{Value: strings.Trim(data["value"].(string), "'")}
	case "formulaItemOptional":
		var err error
		value := data["value"].(map[string]interface{})
		itemsDef := value["items"].([]interface{})
		items := make([]abstractFormulaItem, len(itemsDef))
		for i, item := range itemsDef {
			items[i], err = mapFormulaItem(item.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
		}
		result = &formulaItemOptional{Items: items}
	default:
		return nil, errors.New("Unknown item type: " + option)
	}
	return result, nil
}

func mapType(data map[string]interface{}) (abstractTypeDef, error) {
	option := data["@type"].(string)
	var result interface{}
	switch option {
	case "identifierDef":
		result = &identifierTypeDef{}
	case "stringDef":
		result = &stringTypeDef{}
	case "intDef":
		result = &intTypeDef{}
	case "boolDef":
		result = &boolTypeDef{}
	case "numberDef":
		result = &numberTypeDef{}
	case "enumDef":
		result = &enumTypeDef{}
	case "listDef":
		result = &listTypeDef{}
	case "sequenceDef":
		result = &sequenceTypeDef{}
	case "customTypeRef":
		result = &customTypeRef{}
	case "formulaDef":
		result = &formulaTypeDef{Items: []abstractFormulaItem{}}
	case "oneofDef":
		result = &oneofTypeDef{}
	case "mapDef":
		keyType, err := mapType(data["value"].(map[string]interface{})["value"].(map[string]interface{})["keyType"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		valueTypeDef := data["value"].(map[string]interface{})["value"].(map[string]interface{})["valueType"].(map[string]interface{})
		//TODO fix hack
		if valueTypeDef["@type"] == "structDef" {
			valueTypeDef["value"].(map[string]interface{})["children"] = data["value"].(map[string]interface{})["children"]
		}
		valueType, err := mapType(valueTypeDef)
		if err != nil {
			return nil, err
		}
		return &mapTypeDef{
			keyType,
			valueType,
		}, nil
	case "structDef":
		structure := &structTypeDef{Children: map[string]abstractTypeDef{}}
		if children, ok := data["value"].(map[string]interface{})["children"]; ok && children != nil {
			for key, value := range children.(map[string]interface{}) {
				childType, err := mapType(value.(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				structure.Children[key] = childType
			}
		}
		return structure, nil
	case "complexDef":
		children := data["value"].(map[string]interface{})["children"].(map[string]interface{})
		textType, err := mapType(children["text"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		structureType, err := mapType(children["structure"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return &complexTypeDef{textType, structureType}, nil
	default:
		return nil, errors.New("Unknown option: " + option)
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: decodeHook,
		Result:     &result,
	})
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(data["value"])
	if err != nil {
		return nil, err
	}
	return result, nil
}

func parseSchema(schemaName string, resources ResourceProvider) (DadlSchema, error) {

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

	log.Printf("Parsed schema tree: %v\n", tree)

	var schemaRoot schemRoot

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: decodeHook,
		Result:     &schemaRoot,
	})
	err = decoder.Decode(tree)
	if err != nil {
		return nil, err
	}
	log.Printf("Schema as object: %+v\n", schemaRoot)

	resolver := newResolver(schemaRoot.Types)

	root, err := resolver.buildType(&structTypeDef{
		Children: schemaRoot.Structure,
	})

	if err != nil {
		return nil, err
	}
	log.Printf("Schema: %+v\n", root)
	return &dadlSchemaImpl{root: root}, nil
}

type schemRoot struct {
	Types     map[string]abstractTypeDef
	Structure map[string]abstractTypeDef
}

type abstractTypeDef interface{}

type stringTypeDef struct {
	Regex string
}
type identifierTypeDef struct{}
type intTypeDef struct{}
type numberTypeDef struct{}
type boolTypeDef struct{}
type structTypeDef struct {
	Children map[string]abstractTypeDef
}
type customTypeRef struct {
	TypeName string
}

type abstractFormulaItem interface {
}

type formulaItemVariable struct {
	Name string
	Type abstractTypeDef
}

type formulaItemConstant struct {
	Value string
}

type formulaItemOptional struct {
	Items []abstractFormulaItem
}

type formulaTypeDef struct {
	Items []abstractFormulaItem
}

type oneofTypeDef struct {
	Options []string
}

type complexTypeDef struct {
	TextType      abstractTypeDef
	StructureType abstractTypeDef
}

type enumTypeDef struct {
	Values []string
}
type mapTypeDef struct {
	KeyType   abstractTypeDef
	ValueType abstractTypeDef
}
type listTypeDef struct {
	ItemType abstractTypeDef
}
type sequenceTypeDef struct {
	ItemType abstractTypeDef
}
