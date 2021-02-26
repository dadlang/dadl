package parser

import "errors"

type typeResolver struct {
	typesDefs     map[string]interface{}
	resolvedTypes map[string]valueType
}

func newResolver(typesDefs map[string]interface{}) *typeResolver {
	return &typeResolver{typesDefs: typesDefs, resolvedTypes: map[string]valueType{}}
}

func (r *typeResolver) buildType(typeDef map[string]interface{}) (valueType, error) {
	switch typeDef["baseType"] {
	case "string":
		if typeDef["regex"] != nil {
			return &stringValue{regex: typeDef["regex"].(string)}, nil
		}
		return &stringValue{}, nil
	case "int":
		return &intValue{}, nil
	case "number":
		return &numberValue{}, nil
	case "bool":
		return &boolValue{}, nil
	case "enum":
		enumValue := &enumValue{values: map[string]bool{}}
		for _, value := range typeDef["values"].([]string) {
			enumValue.values[value] = true
		}
		return enumValue, nil
	case "formula":
		items := []formulaItem{}
		for _, item := range typeDef["formula"].([]map[string]interface{}) {
			if item["type"] == "token" {
				itemType, err := r.buildType(item)
				if err != nil {
					return nil, err
				}
				items = append(items, formulaItem{name: item["name"].(string), valueType: itemType})
			} else if item["type"] == "constant" {
				items = append(items, formulaItem{valueType: &constantValue{value: item["value"].(string)}})
			}
		}
		return &formulaValue{formula: items}, nil
	case "sequence":
		sequenceDef := typeDef["sequence"].(map[string]string)
		itemType, err := r.resolveType(sequenceDef["itemType"])
		if err != nil {
			return nil, err
		}
		return &sequenceValue{itemType: itemType}, nil
	case "binary":
		return &binaryValue{}, nil
	case "list":
		valueType, err := r.buildType(typeDef["itemType"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return &listValue{childType: valueType}, nil
	case "map":
		keyType, err := r.buildType(typeDef["keyType"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		valueType, err := r.buildType(typeDef["valueType"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return &mapValue{keyType: keyType, valueType: valueType}, nil
	case "struct":
		var err error
		children := map[string]valueType{}
		childerenDef := typeDef["children"].(map[string]interface{})
		for key, def := range childerenDef {
			children[key], err = r.buildType(def.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
		}
		return &structValue{children: children}, nil
	}
	return r.resolveType(typeDef["baseType"].(string))
}

func (r *typeResolver) resolveType(typeName string) (valueType, error) {
	if resolved, ok := r.resolvedTypes[typeName]; ok {
		return resolved, nil
	}
	var err error
	typeDef := r.typesDefs[typeName]
	if typeDef == nil {
		return nil, errors.New("Unknown type, no definition for: " + typeName)
	}
	r.resolvedTypes[typeName], err = r.buildType(typeDef.(map[string]interface{}))
	if err != nil {
		return nil, err
	}
	return r.resolvedTypes[typeName], nil
}
