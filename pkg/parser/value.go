package parser

import (
	"errors"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type valueType interface {
	toRegex() string
	parse(builder valueBuilder, value string) error
	parseChild(builder valueBuilder, value string) (*nodeInfo, error)
	getChild(name string, builder valueBuilder) (valueType, valueBuilder, error)
}

type binaryAsTextFormat int

const (
	binaryFormatBase64 binaryAsTextFormat = iota
	binaryFormatHex
)

type valueBuilder interface {
	getValue() interface{}
	setValue(value interface{})
}

type delegatedValueBuilder struct {
	set func(value interface{})
	get func() interface{}
}

func (b *delegatedValueBuilder) getValue() interface{} {
	return b.get()
}

func (b *delegatedValueBuilder) setValue(value interface{}) {
	b.set(value)
}

type stringValue struct {
	regex      string
	indentLock int
}

func (v *stringValue) parse(builder valueBuilder, value string) error {
	println("stringValue [parse]:", value)
	builder.setValue(strings.TrimSpace(value))
	v.indentLock = -1
	return nil
}

func (v *stringValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	println("stringValue [parseChild]:", value)
	if v.indentLock < 0 {
		v.indentLock = calcIndentWeight(value)
	}
	value = value[v.indentLock:]
	if existingVal := builder.getValue(); existingVal != nil && existingVal != "" {
		builder.setValue(existingVal.(string) + "\n" + value)
	} else {
		builder.setValue(value)
	}
	return &nodeInfo{valueType: v, builder: builder}, nil
}

func (v *stringValue) toRegex() string {
	if v.regex != "" {
		return v.regex
	}
	return ".*"
}

func (v *stringValue) isSimple() bool {
	return true
}

func (v *stringValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

type boolValue struct {
}

func (v *boolValue) parse(builder valueBuilder, value string) error {
	boolVal, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return err
	}
	builder.setValue(boolVal)
	return nil
}

func (v *boolValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *boolValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *boolValue) toRegex() string {
	return "(?:true)|(?:false)"
}

func (v *boolValue) isSimple() bool {
	return true
}

type constantValue struct {
	value string
}

func (v *constantValue) parse(builder valueBuilder, value string) error {
	return nil
}

func (v *constantValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *constantValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *constantValue) toRegex() string {
	return regexp.QuoteMeta(v.value)
}

type intValue struct {
	min big.Int
	max big.Int
}

func (v *intValue) parse(builder valueBuilder, value string) error {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	builder.setValue(intValue)
	return nil
}

func (v *intValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("not supported")
}

func (v *intValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("not supported")
}

func (v *intValue) toRegex() string {
	return "(-)?\\d+"
}

type numberValue struct {
	min big.Float
	max big.Float
}

func (v *numberValue) parse(builder valueBuilder, value string) error {
	builder.setValue(strings.TrimSpace(value))
	return nil
}

func (v *numberValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *numberValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *numberValue) toRegex() string {
	return "-?(?:\\d+)|(?:\\d*\\.\\d+)"
}

type enumValue struct {
	values map[string]bool
}

func (v *enumValue) parse(builder valueBuilder, value string) error {
	builder.setValue(strings.TrimSpace(value))
	return nil
}

func (v *enumValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *enumValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *enumValue) toRegex() string {
	return "\\S+"
}

type formulaItem struct {
	name      string
	valueType valueType
}

type formulaValue struct {
	formula []formulaItem
	re      *regexp.Regexp
}

func (v *formulaValue) parse(builder valueBuilder, value string) error {

	if v.re == nil {
		var sb strings.Builder
		for _, token := range v.formula {
			if token.name != "" {
				sb.WriteString("(?P<" + token.name + ">" + token.valueType.toRegex() + ")")
			} else {
				sb.WriteString("(?:" + token.valueType.toRegex() + ")")
			}
		}
		v.re = regexp.MustCompile(sb.String())
	}
	var err error
	parsed := map[string]string{}
	match := v.re.FindStringSubmatch(strings.TrimSpace(value))
	// var keyValue string
	if match != nil {
		for i, name := range v.re.SubexpNames() {
			if i != 0 && name != "" {
				parsed[name] = match[i]
			}
		}
	}

	result := map[string]interface{}{}
	for _, token := range v.formula {
		if token.name != "" {
			//	result[token.name],
			err = token.valueType.parse(&delegatedValueBuilder{
				set: func(value interface{}) {
					result[token.name] = value
				},
				get: func() interface{} {
					return result[token.name]
				},
			}, parsed[token.name])
			if err != nil {
				return err
			}
		}
	}
	builder.setValue(result)
	return nil
}

func (v *formulaValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *formulaValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *formulaValue) toRegex() string {
	var sb strings.Builder
	for _, token := range v.formula {
		sb.WriteString("(?:" + token.valueType.toRegex() + ")")
	}
	return sb.String()
}

type sequenceValue struct {
	itemType  valueType
	separator string
	re        *regexp.Regexp
}

func (v *sequenceValue) parse(builder valueBuilder, value string) error {
	if v.re == nil {
		sep := v.separator
		if sep == "" {
			sep = "\\s"
		}
		v.re = regexp.MustCompile("(?:" + sep + ")?(" + v.itemType.toRegex() + ")")
	}

	var err error
	matches := v.re.FindAllStringSubmatch(strings.TrimSpace(value), -1)
	result := make([]interface{}, len(matches))
	for i, match := range matches {
		err = v.itemType.parse(&delegatedValueBuilder{
			get: func() interface{} {
				return result[i]
			},
			set: func(value interface{}) {
				result[i] = value
			},
		}, match[1])
		if err != nil {
			return err
		}
	}
	builder.setValue(result)
	return nil
}

func (v *sequenceValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *sequenceValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *sequenceValue) toRegex() string {
	return "(?:" + v.itemType.toRegex() + ")(?:" + v.separator + "(?:" + v.itemType.toRegex() + "))*"
}

type binaryValue struct {
	textFormat binaryAsTextFormat
}

func (v *binaryValue) parse(builder valueBuilder, value string) error {
	builder.setValue(strings.TrimSpace(value))
	return nil
}

func (v *binaryValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *binaryValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *binaryValue) toRegex() string {
	return ".*"
}

type listValue struct {
	childType valueType
}

func (v *listValue) parse(builder valueBuilder, value string) error {
	println("listValue [parse]:", value)
	builder.setValue([]interface{}{})
	return nil
}

func (v *listValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	println("listValue [parseChild]:", value)

	childvalueBuilder := &delegatedValueBuilder{
		get: func() interface{} {
			return nil
		},
		set: func(value interface{}) {
			builder.setValue(append(builder.getValue().([]interface{}), value))
		},
	}
	v.childType.parse(childvalueBuilder, value)

	builder.getValue()
	return &nodeInfo{
		valueType: v.childType,
		builder:   childvalueBuilder,
	}, nil
}

func (v *listValue) toRegex() string {
	return ""
}

func (v *listValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not found: " + name)
}

type mapValue struct {
	keyType   valueType
	valueType valueType
}

func (v *mapValue) parse(builder valueBuilder, value string) error {
	println("mapValue [parse]:", value)
	builder.setValue(map[string]interface{}{})
	return nil
}

func (v *mapValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	println("mapValue [parseChild]:", value)

	//TODO
	parts := strings.SplitN(strings.TrimSpace(value), " ", 2)

	childValueBuilder := &delegatedValueBuilder{
		get: func() interface{} {
			return builder.getValue().(map[string]interface{})[parts[0]]
		},
		set: func(value interface{}) {
			if builder.getValue() == nil {
				builder.setValue(map[string]interface{}{})
			}
			builder.getValue().(map[string]interface{})[parts[0]] = value
		},
	}
	if len(parts) > 1 {
		v.valueType.parse(childValueBuilder, parts[1])
	} else {
		//TODO
		v.valueType.parse(childValueBuilder, "")
	}

	builder.getValue()
	return &nodeInfo{
		valueType: v.valueType,
		builder:   childValueBuilder,
	}, nil
}

func (v *mapValue) toRegex() string {
	//TODO
	return v.keyType.toRegex()
}

func (v *mapValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	if builder.getValue() == nil {
		//	builder.setValue(map[string]interface{}{})
	}

	return v.valueType, &delegatedValueBuilder{
		get: func() interface{} {
			return builder.getValue().(map[string]interface{})[name]
		},
		set: func(value interface{}) {
			builder.getValue().(map[string]interface{})[name] = value
		},
	}, nil
}

type structValue struct {
	children map[string]valueType
}

func (v *structValue) parse(builder valueBuilder, value string) error {
	println("structValue [parse]:", value)
	builder.setValue(map[string]interface{}{})
	return nil
}

func (v *structValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	println("structValue [parseChild]:", value)

	res := map[string]string{}
	match := keyWithDelegatedValueRe.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		return nil, errors.New("invalid format")
	}
	if match != nil {
		for i, name := range keyWithDelegatedValueRe.SubexpNames() {
			if i != 0 && name != "" {
				res[name] = match[i]
			}
		}
	}

	key := removeQuotes(res["key"])

	if childType, ok := v.children[key]; ok { //valueBuilder.getValue()[key]; ok {
		childvalueBuilder := &delegatedValueBuilder{
			get: func() interface{} {
				return builder.getValue().(map[string]interface{})[key]
			},
			set: func(value interface{}) {
				if builder.getValue() == nil {
					builder.setValue(map[string]interface{}{})
				}
				builder.getValue().(map[string]interface{})[key] = value
			},
		}
		childType.parse(childvalueBuilder, res["rest"])

		println("key: ", key)

		builder.getValue()
		return &nodeInfo{
			valueType: childType,
			builder:   childvalueBuilder,
		}, nil
	}
	return nil, errors.New("No child with name: " + key)
}

func (v *structValue) toRegex() string {
	return ""
}

func (v *structValue) isSimple() bool {
	return false
}

func (v *structValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	if c, ok := v.children[name]; ok {

		if builder.getValue() == nil {
			builder.setValue(map[string]interface{}{})
		}

		return c, &delegatedValueBuilder{
			get: func() interface{} {
				return builder.getValue().(map[string]interface{})[name]
			},
			set: func(value interface{}) {
				builder.getValue().(map[string]interface{})[name] = value
			},
		}, nil
	}
	return nil, nil, errors.New("Not found: " + name)
}
