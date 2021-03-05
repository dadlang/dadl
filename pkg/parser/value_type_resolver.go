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
	switch item.(type) {
	case *formulaItemVariable:
		itemType, err := r.buildType(item.(*formulaItemVariable).Type)
		if err != nil {
			return formulaItem{}, err
		}
		return formulaItem{name: item.(*formulaItemVariable).Name, valueType: itemType}, nil
	case *formulaItemConstant:
		return formulaItem{valueType: &constantValue{value: item.(*formulaItemConstant).Value}}, nil
	case *formulaItemOptional:
		childrenDef := item.(*formulaItemOptional).Items
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
		return formulaItem{}, errors.New("Unknown item type")
	}
}

func (r *typeResolver) buildType(typeDef abstractTypeDef) (valueType, error) {
	switch typeDef.(type) {
	case *stringTypeDef:
		regex := typeDef.(*stringTypeDef).Regex
		if regex != "" {
			return &stringValue{regex: regex}, nil
		}
		return &stringValue{}, nil
	case *identifierTypeDef:
		return &stringValue{regex: "[A-Za-z-0-9_-]+"}, nil
	case *intTypeDef:
		return &intValue{}, nil
	case *numberTypeDef:
		return &numberValue{}, nil
	case *boolTypeDef:
		return &boolValue{}, nil
	case *enumTypeDef:
		enumValue := &enumValue{values: map[string]bool{}}
		for _, value := range typeDef.(*enumTypeDef).Values {
			enumValue.values[value] = true
		}
		return enumValue, nil
	case *formulaTypeDef:
		items := []formulaItem{}
		for _, itemDef := range typeDef.(*formulaTypeDef).Items {
			item, err := r.buildFormulaItem(itemDef)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return &formulaValue{formula: items}, nil
	case *sequenceTypeDef:
		itemType, err := r.buildType(typeDef.(*sequenceTypeDef).ItemType)
		if err != nil {
			return nil, err
		}
		return &sequenceValue{itemType: itemType}, nil
	// case "binaryDef":
	// 	return &binaryValue{}, nil
	case *listTypeDef:
		valueType, err := r.buildType(typeDef.(*listTypeDef).ItemType)
		if err != nil {
			return nil, err
		}
		return &listValue{childType: valueType}, nil
	case *mapTypeDef:
		keyType, err := r.buildType(typeDef.(*mapTypeDef).KeyType)
		if err != nil {
			return nil, err
		}
		valueType, err := r.buildType(typeDef.(*mapTypeDef).ValueType)
		if err != nil {
			return nil, err
		}
		return &mapValue{keyType: keyType, valueType: valueType}, nil
	case *structTypeDef:
		var err error
		children := map[string]valueType{}
		childerenDef := typeDef.(*structTypeDef).Children
		for key, def := range childerenDef {
			children[key], err = r.buildType(def)
			if err != nil {
				return nil, err
			}
		}
		return &structValue{children: children}, nil
	case *oneofTypeDef:
		optionsDef := typeDef.(*oneofTypeDef).Options
		options := make([]oneofValueOption, len(optionsDef))
		for i, typeName := range optionsDef {
			valueType, err := r.resolveType(typeName)
			if err != nil {
				return nil, err
			}
			options[i] = oneofValueOption{name: typeName, valueType: valueType}
		}
		return &oneofValue{options: options}, nil
	case *complexTypeDef:
		textType, err := r.buildType(typeDef.(*complexTypeDef).TextType)
		if err != nil {
			return nil, err
		}
		structureType, err := r.buildType(typeDef.(*complexTypeDef).StructureType)
		if err != nil {
			return nil, err
		}
		return &complexValue{textValue: textType, structValue: structureType}, nil
	case *customTypeRef:
		return r.resolveType(typeDef.(*customTypeRef).TypeName)
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
