package parser

import (
	"errors"
	"log"
	"math/big"
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
	textualTypeDef := &oneofValue{}
	typeWithCommentDef := &formulaValue{
		formula: []formulaItem{
			{
				spread:       true,
				valueType:    typeDef,
				asStructType: true,
			},
			{
				optional:  true,
				composite: true,
				children: []formulaItem{
					{
						valueType: &stringValue{regex: "\\s*#"},
					},
					{
						name:      "comment",
						valueType: &stringValue{},
					},
				},
			},
		},
	}
	structType := &complexValue{
		textValue: &formulaValue{
			formula: []formulaItem{
				{
					optional:  true,
					valueType: &constantValue{value: "struct"},
				},
			},
		},
		structValue: &mapValue{
			keyType:   keyType,
			valueType: typeDef,
		},
		structValueKey: "children",
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
						valueType: &constantValue{value: "`"},
					},
					{
						name:      "regex",
						valueType: &stringValue{regex: ".*?"},
					},
					{
						valueType: &constantValue{value: "`"},
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
						name:      "min",
						valueType: &intValue{},
					},
					{
						valueType: &constantValue{value: ".."},
					},
					{
						name:      "max",
						valueType: &intValue{},
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
				optional:  true,
				composite: true,
				children: []formulaItem{
					{
						valueType: &constantValue{value: "["},
					},
					{
						name:      "valueType",
						valueType: textualTypeDef,
					},
					{
						valueType: &constantValue{value: "]"},
					},
				},
			},
			{
				valueType: whitespace,
			},
			{
				name: "values",
				valueType: &sequenceValue{
					itemType: &formulaValue{
						formula: []formulaItem{
							{
								name:      "textValue",
								valueType: &stringValue{regex: regexIdentifier},
							},
							{
								optional:  true,
								composite: true,
								children: []formulaItem{
									{
										valueType: &constantValue{value: "["},
									},
									{
										name:      "mappedValue",
										valueType: &stringValue{},
									},
									{
										valueType: &constantValue{value: "]"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	formulaItemDefSequenceDelegated := &delegatedValue{}
	formulaItemDef := &oneofValue{
		options: []oneofValueOption{
			{
				Name:      "formulaItemConstant",
				ValueType: &stringValue{regex: "'.*?'"},
				ValueKey:  "value",
			},
			{
				Name:      "formulaItemRegex",
				ValueType: &stringValue{regex: "`.*`"},
				ValueKey:  "regex",
			},
			{
				Name: "formulaItemVariable",
				ValueType: &formulaValue{
					formula: []formulaItem{

						{
							valueType: &constantValue{value: "<"},
						},
						{
							name:      "structType",
							optional:  true,
							valueType: &constantValue{value: "+"},
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
							valueType: typeDef,
						},
						{
							valueType: &constantValue{value: ">"},
						},
					},
				},
			},
			{
				Name: "formulaItemOptional",
				ValueType: &formulaValue{
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
				valueType: &constantValue{value: "["},
			},
			{
				name:      "itemType",
				valueType: typeDef,
			},
			{
				valueType: &constantValue{value: "]"},
			},
		},
	}
	listDef := &formulaValue{
		formula: []formulaItem{
			{
				valueType: &stringValue{regex: "list"},
			},
			{
				valueType: &constantValue{value: "["},
			},
			{
				name:      "itemType",
				valueType: typeDef,
			},
			{
				valueType: &constantValue{value: "]"},
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
					valueType: typeDef,
				},
				{
					valueType: &constantValue{value: "]"},
				},
				{
					optional:  true,
					name:      "valueType",
					valueType: typeDef,
				},
			},
		},
		structValue: &mapValue{
			keyType:   keyType,
			valueType: typeDef,
		},
		structValueKey: "valueType.children",
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
				valueType: &constantValue{value: "["},
			},
			{
				name: "options",
				valueType: &sequenceValue{
					itemType:  &stringValue{regex: regexIdentifier},
					separator: "|",
				},
			},
			{
				valueType: &constantValue{value: "]"},
			},
		},
	}
	complexDef := &complexValue{
		textValue: &formulaValue{
			formula: []formulaItem{
				{
					valueType: &stringValue{regex: "complex"},
				},
				{
					valueType: &constantValue{value: "["},
				},
				{
					name:      "spreadValue",
					optional:  true,
					valueType: &constantValue{value: "..."},
				},
				{
					name:      "valueType",
					valueType: typeDef,
				},
				{
					valueType: &constantValue{value: "]"},
				},
				{
					name:      "spreadChildren",
					optional:  true,
					valueType: &constantValue{value: "..."},
				},
				{
					name:      "childType",
					valueType: typeDef,
				},
			},
		},
		textValueKey: "",
		structValue: &mapValue{
			keyType:   keyType,
			valueType: typeDef,
		},
		structValueKey: "childType.children",
	}
	textualTypeDef.options = []oneofValueOption{
		{
			Name:      "stringDef",
			ValueType: stringDef,
		},
		{
			Name:      "identifierDef",
			ValueType: identifierDef,
		},
		{
			Name:      "intDef",
			ValueType: intDef,
		},
		{
			Name:      "boolDef",
			ValueType: boolDef,
		},
		{
			Name:      "numberDef",
			ValueType: numberDef,
		},
		{
			Name:      "enumDef",
			ValueType: enumDef,
		},
		{
			Name:      "formulaDef",
			ValueType: formulaDef,
		},
		{
			Name:      "sequenceDef",
			ValueType: sequenceDef,
		},
		{
			Name:      "customTypeRef",
			ValueType: customTypeRef,
		},
	}
	typeDef.options = []oneofValueOption{
		{
			Name:      "stringDef",
			ValueType: stringDef,
		},
		{
			Name:      "identifierDef",
			ValueType: identifierDef,
		},
		{
			Name:      "intDef",
			ValueType: intDef,
		},
		{
			Name:      "boolDef",
			ValueType: boolDef,
		},
		{
			Name:      "numberDef",
			ValueType: numberDef,
		},
		{
			Name:      "enumDef",
			ValueType: enumDef,
		},
		{
			Name:      "formulaDef",
			ValueType: formulaDef,
		},
		{
			Name:      "sequenceDef",
			ValueType: sequenceDef,
		},
		{
			Name:      "listDef",
			ValueType: listDef,
		},
		{
			Name:      "mapDef",
			ValueType: mapDef,
		},
		{
			Name:      "structDef",
			ValueType: structType,
		},
		{
			Name:      "oneofDef",
			ValueType: oneofDef,
		},
		{
			Name:      "complexDef",
			ValueType: complexDef,
		},
		{
			Name:      "customTypeRef",
			ValueType: customTypeRef,
		},
	}
	typesDef := &mapValue{
		keyType:   keyType,
		valueType: typeWithCommentDef,
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
		mappedType, err := mapType(data["type"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		result = &formulaItemVariable{
			Name:        data["name"].(string),
			Type:        mappedType,
			StructValue: data["structType"] == true,
		}
	case "formulaItemConstant":
		result = &formulaItemConstant{Value: strings.Trim(data["value"].(string), "'")}
	case "formulaItemRegex":
		result = &formulaItemRegex{Regex: strings.Trim(data["regex"].(string), "`")}
	case "formulaItemOptional":
		var err error
		itemsDef := data["items"].([]interface{})
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
		if data["min"] != nil {
			result.(*intTypeDef).Min = data["min"].(*big.Int)
		}
		if data["max"] != nil {
			result.(*intTypeDef).Max = data["max"].(*big.Int)
		}
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
		keyType, err := mapType(data["keyType"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		valueTypeDef := data["valueType"].(map[string]interface{})
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
		if children, ok := data["children"]; ok && children != nil {
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
		valueType, err := mapType(data["valueType"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		childType, err := mapType(data["childType"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		spreadValue := data["spreadValue"] == true
		spreadChildren := data["spreadChildren"] == true
		return &complexTypeDef{valueType, childType, spreadValue, spreadChildren}, nil
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
	err = decoder.Decode(data)
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
type intTypeDef struct {
	Min *big.Int
	Max *big.Int
}
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
	Name        string
	Type        abstractTypeDef
	StructValue bool
}

type formulaItemConstant struct {
	Value string
}

type formulaItemRegex struct {
	Regex string
}

type formulaItemOptional struct {
	Items []abstractFormulaItem
}

type formulaTypeDef struct {
	Items []abstractFormulaItem
}

type oneofTypeDef struct {
	Options     []string
	SpreadValue bool
}

type complexTypeDef struct {
	ValueType      abstractTypeDef
	ChildType      abstractTypeDef
	SpreadValue    bool
	SpreadChildren bool
}

type enumTypeDef struct {
	ValueType abstractTypeDef
	Values    []enumTypeValueDef
}

type enumTypeValueDef struct {
	TextValue   string
	MappedValue string
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
