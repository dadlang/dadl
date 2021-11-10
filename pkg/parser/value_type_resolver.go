package parser

import (
	"errors"
	"reflect"
)

type typeResolver struct {
	typesDefs     map[string]abstractTypeDef
	resolvedTypes map[string]valueType
}

func newResolver(typesDefs map[string]abstractTypeDef) *typeResolver {
	return &typeResolver{typesDefs: typesDefs, resolvedTypes: map[string]valueType{}}
}

func (r *typeResolver) buildFormulaItem(item abstractFormulaItem) (formulaItem, error) {
	switch item := item.(type) {
	case *formulaItemVariable:
		itemType, err := r.buildType(item.Type)
		if err != nil {
			return formulaItem{}, err
		}
		return formulaItem{name: item.Name, valueType: itemType, asStructType: item.StructValue}, nil
	case *formulaItemConstant:
		return formulaItem{valueType: &constantValue{value: item.Value}}, nil
	case *formulaItemOptional:
		childrenDef := item.Items
		children := make([]formulaItem, len(childrenDef))
		for i, itemDef := range childrenDef {
			childItem, err := r.buildFormulaItem(itemDef)
			if err != nil {
				return formulaItem{}, err
			}
			children[i] = childItem
		}
		return formulaItem{optional: true, composite: true, children: children}, nil
	default:
		return formulaItem{}, errors.New("unknown item type")
	}
}

func (r *typeResolver) buildType(typeDef abstractTypeDef) (valueType, error) {
	switch typeDef := typeDef.(type) {
	case *stringTypeDef:
		regex := typeDef.Regex
		if regex != "" {
			return &stringValue{regex: regex}, nil
		}
		return &stringValue{}, nil
	case *identifierTypeDef:
		return &stringValue{regex: "[A-Za-z-0-9_-]+"}, nil
	case *intTypeDef:
		return &intValue{min: typeDef.Min, max: typeDef.Max}, nil
	case *numberTypeDef:
		return &numberValue{}, nil
	case *boolTypeDef:
		return &boolValue{}, nil
	case *enumTypeDef:
		var valueType valueType
		var err error
		if typeDef.ValueType != nil {
			valueType, err = r.buildType(typeDef.ValueType)
			if err != nil {
				return nil, err
			}
		} else {
			valueType = &stringValue{}
		}
		enumValue := &enumValue{values: map[string]string{}, valueType: valueType}
		for _, value := range typeDef.Values {
			if value.MappedValue != "" {
				enumValue.values[value.TextValue] = value.MappedValue
			} else {
				enumValue.values[value.TextValue] = value.TextValue
			}
		}
		return enumValue, nil
	case *formulaTypeDef:
		items := []formulaItem{}
		for _, itemDef := range typeDef.Items {
			item, err := r.buildFormulaItem(itemDef)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return &formulaValue{formula: items}, nil
	case *sequenceTypeDef:
		itemType, err := r.buildType(typeDef.ItemType)
		if err != nil {
			return nil, err
		}
		return &sequenceValue{itemType: itemType}, nil
	// case "binaryDef":
	// 	return &binaryValue{}, nil
	case *listTypeDef:
		valueType, err := r.buildType(typeDef.ItemType)
		if err != nil {
			return nil, err
		}
		return &listValue{childType: valueType}, nil
	case *mapTypeDef:
		keyType, err := r.buildType(typeDef.KeyType)
		if err != nil {
			return nil, err
		}
		valueType, err := r.buildType(typeDef.ValueType)
		if err != nil {
			return nil, err
		}
		return &mapValue{keyType: keyType, valueType: valueType}, nil
	case *structTypeDef:
		var err error
		children := map[string]valueType{}
		childerenDef := typeDef.Children
		for key, def := range childerenDef {
			children[key], err = r.buildType(def)
			if err != nil {
				return nil, err
			}
		}
		return &structValue{children: children}, nil
	case *oneofTypeDef:
		optionsDef := typeDef.Options
		options := make([]oneofValueOption, len(optionsDef))
		for i, typeName := range optionsDef {
			valueType, err := r.resolveType(typeName)
			if err != nil {
				return nil, err
			}
			options[i] = oneofValueOption{Name: typeName, ValueType: valueType}
		}
		return &oneofValue{options: options}, nil
	case *complexTypeDef:
		textType, err := r.buildType(typeDef.ValueType)
		if err != nil {
			return nil, err
		}
		structureType, err := r.buildType(typeDef.ChildType)
		if err != nil {
			return nil, err
		}
		textValueKey := "value"
		if typeDef.SpreadValue {
			textValueKey = ""
		}
		structValueKey := "children"
		if typeDef.SpreadChildren {
			structValueKey = ""
		}
		return &complexValue{textValue: textType, structValue: structureType, textValueKey: textValueKey, structValueKey: structValueKey}, nil
	case *customTypeRef:
		return r.resolveType(typeDef.TypeName)
	}
	return nil, errors.New("Unsupported type: " + reflect.TypeOf(typeDef).Name())
}

func (r *typeResolver) resolveType(typeName string) (valueType, error) {
	if resolved, ok := r.resolvedTypes[typeName]; ok {
		return resolved, nil
	}
	typeDef, ok := r.typesDefs[typeName]
	if !ok {
		return nil, errors.New("Unknown type, no definition for: " + typeName)
	}
	delegate := delegatedValue{}
	// register delegated type to support circural dependencies
	r.resolvedTypes[typeName] = &delegate
	resolvedType, err := r.buildType(typeDef)
	if err != nil {
		return nil, err
	}
	delegate.target = resolvedType
	r.resolvedTypes[typeName] = resolvedType
	return r.resolvedTypes[typeName], nil
}
