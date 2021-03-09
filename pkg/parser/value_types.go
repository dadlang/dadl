package parser

import (
	"errors"
	"log"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type regexBuildContext struct {
	usage map[valueType]int
}

type valueType interface {
	toRegex(ctx regexBuildContext) string
	parse(builder valueBuilder, value string) error
	parseChild(builder valueBuilder, value string) (*nodeInfo, error)
	getChild(name string, builder valueBuilder) (valueType, valueBuilder, error)
	supportsChildren() bool
}

type binaryAsTextFormat int

const (
	binaryFormatBase64 binaryAsTextFormat = iota
	binaryFormatHex
)

type delegatedValueBuilder struct {
	metaBuilder
	set func(value interface{})
	get func() interface{}
}

func (b *delegatedValueBuilder) getSimpleValue() interface{} {
	return b.get()
}

// func (b *delegatedValueBuilder) setValue(value interface{}) {
// 	b.set(value)
// }

func (b *delegatedValueBuilder) setSimpleValue(value interface{}) {
	b.set(value)
}
func (b *delegatedValueBuilder) getField(name string) interface{} {
	if b.get() == nil {
		b.set(map[string]interface{}{})
	}
	return b.get().(map[string]interface{})[name]
}
func (b *delegatedValueBuilder) setField(name string, value interface{}) {
	if b.get() == nil {
		b.set(map[string]interface{}{})
	}
	b.get().(map[string]interface{})[name] = value
}
func (b *delegatedValueBuilder) addListElement(value interface{}) int {
	if b.get() == nil {
		b.set([]interface{}{})
	}
	newSequence := append(b.get().([]interface{}), value)
	b.set(newSequence)
	return len(newSequence) - 1
}

func (b *delegatedValueBuilder) getListElement(idx int) interface{} {
	if b.get() == nil {
		b.set([]interface{}{})
	}
	return b.get().([]interface{})[idx]
}

func (b *delegatedValueBuilder) setListElement(idx int, value interface{}) {
	if b.get() == nil {
		b.set([]interface{}{})
	}
	if len(b.get().([]interface{})) == idx {
		b.set(append(b.get().([]interface{}), value))
	} else {
		b.get().([]interface{})[idx] = value
	}
}

type stringValue struct {
	regex      string
	indentLock int
}

func (v *stringValue) parse(builder valueBuilder, value string) error {
	log.Println("stringValue [parse]:", value)
	builder.setSimpleValue(strings.TrimSpace(value))
	v.indentLock = -1
	return nil
}

func (v *stringValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	log.Println("stringValue [parseChild]:", value)
	if v.indentLock < 0 {
		v.indentLock = calcIndentWeight(value)
	}
	value = value[v.indentLock:]
	if existingVal := builder.getSimpleValue(); existingVal != nil && existingVal != "" {
		builder.setSimpleValue(existingVal.(string) + "\n" + value)
	} else {
		builder.setSimpleValue(value)
	}
	return &nodeInfo{valueType: v, builder: builder}, nil
}

func (v *stringValue) toRegex(ctx regexBuildContext) string {
	if v.regex != "" {
		return v.regex
	}
	return ".*?"
}

func (v *stringValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("[stringValue] Not supported")
}

func (v *stringValue) supportsChildren() bool {
	return false
}

type boolValue struct {
}

func (v *boolValue) parse(builder valueBuilder, value string) error {
	boolVal, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return err
	}
	builder.setSimpleValue(boolVal)
	return nil
}

func (v *boolValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("[boolValue] Not supported")
}

func (v *boolValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *boolValue) toRegex(ctx regexBuildContext) string {
	return "(?:true)|(?:false)"
}

func (v *boolValue) supportsChildren() bool {
	return false
}

type constantValue struct {
	value string
}

func (v *constantValue) parse(builder valueBuilder, value string) error {
	builder.setSimpleValue(value == v.value)
	return nil
}

func (v *constantValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *constantValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *constantValue) toRegex(ctx regexBuildContext) string {
	return regexp.QuoteMeta(v.value)
}

func (v *constantValue) supportsChildren() bool {
	return false
}

type intValue struct {
	min big.Int
	max big.Int
}

func (v *intValue) parse(builder valueBuilder, value string) error {
	log.Println("intValue [parse]:", value)
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	builder.setSimpleValue(intValue)
	return nil
}

func (v *intValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	log.Println("intValue [parseChild]:", value)
	return nil, errors.New("not supported")
}

func (v *intValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("not supported")
}

func (v *intValue) toRegex(ctx regexBuildContext) string {
	return "(?:-)?\\d+"
}

func (v *intValue) supportsChildren() bool {
	return false
}

type numberValue struct {
	min big.Float
	max big.Float
}

func (v *numberValue) parse(builder valueBuilder, value string) error {
	builder.setSimpleValue(strings.TrimSpace(value))
	return nil
}

func (v *numberValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *numberValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *numberValue) toRegex(ctx regexBuildContext) string {
	return "-?(?:\\d+)|(?:\\d*\\.\\d+)"
}

func (v *numberValue) supportsChildren() bool {
	return false
}

type enumValue struct {
	valueType valueType
	values    map[string]string
}

func (v *enumValue) parse(builder valueBuilder, value string) error {
	mappedValue, ok := v.values[value]
	if ok {
		return v.valueType.parse(builder, mappedValue)
	} else {
		return errors.New("Unsupported enum value: " + value)
	}
}

func (v *enumValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *enumValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *enumValue) toRegex(ctx regexBuildContext) string {
	return "\\S+"
}

func (v *enumValue) supportsChildren() bool {
	return false
}

type formulaItem struct {
	name         string
	spread       bool
	valueType    valueType
	optional     bool
	composite    bool
	children     []formulaItem
	asStructType bool
}

type formulaValue struct {
	formula     []formulaItem
	_re         *regexp.Regexp
	_mapping    []formulaItem
	_structType valueType
}

func buildMapping(mapping *[]formulaItem, items []formulaItem, structType *valueType) {
	for _, item := range items {
		if item.composite {
			buildMapping(mapping, item.children, structType)
		} else if item.name != "" || item.spread {
			*mapping = append(*mapping, item)
			if item.asStructType {
				*structType = item.valueType
			}
		}
	}
}

func (v *formulaValue) initIfRequired() error {
	if v._re == nil {
		var sb strings.Builder
		sb.WriteString("^")
		sb.WriteString(buildItemsRegex(v.formula, true, newRegexBuildContext()))
		sb.WriteString("$")
		log.Println("formulaValue [regex]:", sb.String())
		v._re = regexp.MustCompile(sb.String())
		v._mapping = []formulaItem{}
		buildMapping(&v._mapping, v.formula, &v._structType)
	}
	return nil
}

func (v *formulaValue) parse(builder valueBuilder, value string) error {
	log.Println("formulaValue [parse]:", value)
	v.initIfRequired()
	valueToParse := strings.TrimSpace(value)
	match := v._re.FindStringSubmatchIndex(valueToParse)
	if match != nil {
		for i, item := range v._mapping {
			itemIdxFrom := i*2 + 2
			if match[itemIdxFrom] < 0 {
				continue
			}
			matchValue := valueToParse[match[itemIdxFrom]:match[itemIdxFrom+1]]
			var newBuilder valueBuilder
			if item.spread {
				newBuilder = builder
			} else {
				newBuilder = builder.getFieldBuilder(item.name)
			}
			err := item.valueType.parse(newBuilder, matchValue)
			if err != nil {
				return err
			}
		}
	} else {
		return errors.New("No match for: " + value)
	}
	return nil
}

func (v *formulaValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	v.initIfRequired()
	if v._structType != nil {
		return v._structType.parseChild(builder, value)
	}
	return nil, errors.New("[formulaValue] Not supported: parseChild")
}

func (v *formulaValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("[formulaValue] Not supported: getChild")
}

func (v *formulaValue) toRegex(ctx regexBuildContext) string {
	nextCtx, result := ctx.With(v)
	if nextCtx == nil {
		return result
	}
	return buildItemsRegex(v.formula, false, *nextCtx)
}

func buildItemsRegex(items []formulaItem, nonCapture bool, ctx regexBuildContext) string {
	var sb strings.Builder
	for _, token := range items {
		sb.WriteString(buildItemRegex(token, nonCapture, ctx))
	}
	return sb.String()
}

func buildItemRegex(item formulaItem, capture bool, ctx regexBuildContext) string {
	var sb strings.Builder
	if item.composite {
		sb.WriteString("(?:")
		sb.WriteString(buildItemsRegex(item.children, capture, ctx))
		sb.WriteString(")")
	} else {
		if !capture || (item.name == "" && !item.spread) {
			sb.WriteString("(?:" + item.valueType.toRegex(ctx) + ")")
		} else {
			sb.WriteString("(" + item.valueType.toRegex(ctx) + ")")
		}
	}
	if item.optional {
		sb.WriteString("?")
	}
	return sb.String()
}

func (v *formulaValue) supportsChildren() bool {
	return false
}

type sequenceValue struct {
	itemType  valueType
	separator string
	re        *regexp.Regexp
}

func (v *sequenceValue) parse(builder valueBuilder, value string) error {
	log.Println("sequenceValue [parse]:", value)
	if v.re == nil {
		sep := v.separator
		if sep == "" {
			sep = "\\s"
		}
		ctx := newRegexBuildContext()
		re := "^(" + v.itemType.toRegex(ctx) + ")(?:(?:" + sep + ")((?:" + v.itemType.toRegex(ctx) + ")(?:(?:" + sep + ")(?:" + v.itemType.toRegex(ctx) + "))*))?$"
		log.Println("sequenceValue [regex]:", re)
		v.re = regexp.MustCompile(re)
	}

	match := v.re.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		return errors.New("sequenceValue [parse]: No match")
	}
	matches := []string{}
	matches = append(matches, match[1])
	for match[2] != "" {
		match = v.re.FindStringSubmatch(match[2])
		if match == nil {
			return errors.New("sequenceValue [parse]: No match")
		}
		matches = append(matches, match[1])
	}
	for _, match := range matches {
		err := v.itemType.parse(builder.getListItemBuilder(), match)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *sequenceValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("[sequenceValue] Not supported")
}

func (v *sequenceValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("[sequenceValue] Not supported")
}

func (v *sequenceValue) toRegex(ctx regexBuildContext) string {
	nextCtx, result := ctx.With(v)
	if nextCtx == nil {
		return result
	}
	sep := v.separator
	if sep == "" {
		sep = "\\s"
	}
	return "(?:" + v.itemType.toRegex(*nextCtx) + ")(?:" + sep + "(?:" + v.itemType.toRegex(*nextCtx) + "))*"
}

func (v *sequenceValue) supportsChildren() bool {
	return false
}

type binaryValue struct {
	textFormat binaryAsTextFormat
}

func (v *binaryValue) parse(builder valueBuilder, value string) error {
	log.Println("binaryValue [parse]:", value)
	builder.setSimpleValue(strings.TrimSpace(value))
	return nil
}

func (v *binaryValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return nil, errors.New("Not supported")
}

func (v *binaryValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not supported")
}

func (v *binaryValue) toRegex(ctx regexBuildContext) string {
	return ".*"
}

func (v *binaryValue) supportsChildren() bool {
	return false
}

type listValue struct {
	childType valueType
}

func (v *listValue) parse(builder valueBuilder, value string) error {
	log.Println("listValue [parse]:", value)
	return nil
}

func (v *listValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	log.Println("listValue [parseChild]:", value)

	childBuilder := builder.getListItemBuilder()
	err := v.childType.parse(childBuilder, value)
	if err != nil {
		return nil, err
	}

	return &nodeInfo{
		valueType: v.childType,
		builder:   childBuilder,
	}, nil
}

func (v *listValue) toRegex(ctx regexBuildContext) string {
	return ""
}

func (v *listValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("Not found: " + name)
}

func (v *listValue) supportsChildren() bool {
	return false
}

type mapValue struct {
	keyType   valueType
	valueType valueType
}

func (v *mapValue) parse(builder valueBuilder, value string) error {
	log.Println("mapValue [parse]:", value)
	// if builder.getValue() == nil {
	// 	builder.setValue(map[string]interface{}{})
	// }
	return nil
}

func (v *mapValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	log.Println("mapValue [parseChild]:", value)

	//TODO
	parts := strings.SplitN(strings.TrimSpace(value), " ", 2)

	childBuilder := builder.getFieldBuilder(parts[0])
	if len(parts) > 1 {
		err := v.valueType.parse(childBuilder, parts[1])
		if err != nil {
			return nil, err
		}
	} else {
		//TODO
		err := v.valueType.parse(childBuilder, "")
		if err != nil {
			return nil, err
		}
	}
	return &nodeInfo{
		valueType: v.valueType,
		builder:   childBuilder,
	}, nil
}

func (v *mapValue) toRegex(ctx regexBuildContext) string {
	nextCtx, result := ctx.With(v)
	if nextCtx == nil {
		return result
	}
	return v.keyType.toRegex(*nextCtx)
}

func (v *mapValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return v.valueType, builder.getFieldBuilder(name), nil
}

func (v *mapValue) supportsChildren() bool {
	return true
}

type structValue struct {
	children map[string]valueType
}

func (v *structValue) parse(builder valueBuilder, value string) error {
	log.Println("structValue [parse]:", value)
	// builder.setValue(map[string]interface{}{})
	return nil
}

func (v *structValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	log.Println("structValue [parseChild]:", value)

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
		childValueBuilder := builder.getFieldBuilder(key)
		err := childType.parse(childValueBuilder, res["rest"])
		if err != nil {
			return nil, err
		}
		return &nodeInfo{
			valueType: childType,
			builder:   childValueBuilder,
		}, nil
	}
	return nil, errors.New("No child with name: " + key)
}

func (v *structValue) toRegex(ctx regexBuildContext) string {
	return ""
}

func (v *structValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	if c, ok := v.children[name]; ok {
		return c, builder.getFieldBuilder(name), nil
	}
	return nil, nil, errors.New("Not found: " + name)
}

func (v *structValue) supportsChildren() bool {
	return true
}

type oneofValueOption struct {
	Name      string
	ValueType valueType
	ValueKey  string
}

type oneofValue struct {
	TypeKey string
	options []oneofValueOption
	_res    []*regexp.Regexp
	// _optionsMap map[string]valueType
}

func (v *oneofValue) parse(builder valueBuilder, value string) error {
	log.Println("oneofValue [parse]:", value)
	if len(v._res) == 0 {
		v._res = make([]*regexp.Regexp, len(v.options))
		for i, option := range v.options {
			v._res[i] = regexp.MustCompile("^" + option.ValueType.toRegex(newRegexBuildContext()) + "$")
		}
	}

	// parsed := map[string]interface{}{}
	trimmedValue := strings.TrimSpace(value)
	builder.setMeta("lastMatch", nil)
	for i, re := range v._res {
		match := re.FindStringSubmatch(trimmedValue)
		if match != nil {
			matchedOption := v.options[i]
			optionName := matchedOption.Name
			log.Println("oneofValue [match]:", optionName)
			if v.TypeKey == "" {
				builder.getFieldBuilder("@type").setSimpleValue(optionName)
			} else {
				builder.getFieldBuilder(v.TypeKey).setSimpleValue(optionName)
			}
			valueBuilder := builder
			if matchedOption.ValueKey != "" {
				valueBuilder = builder.getFieldBuilder(matchedOption.ValueKey)
			}
			err := matchedOption.ValueType.parse(valueBuilder, trimmedValue)
			if err != nil {
				return err
			}
			builder.setMeta("lastMatch", i)
			return nil
		}
	}
	return errors.New("No match for: " + value)
}

func (v *oneofValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	log.Println("oneofValue [parseChild]:", value)

	lastMatch := builder.getMeta("lastMatch")
	if lastMatch != nil {
		lastOption := v.options[lastMatch.(int)]
		log.Println("oneofValue [lastOption]:", lastOption.Name)
		valueBuilder := builder
		if lastOption.ValueKey != "" {
			valueBuilder = builder.getFieldBuilder(lastOption.ValueKey)
		}
		return lastOption.ValueType.parseChild(valueBuilder, value)
	}
	return nil, errors.New("[oneofValue] Not supported")
}

func (v *oneofValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return nil, nil, errors.New("[oneofValue] Not supported")
}

func (v *oneofValue) toRegex(ctx regexBuildContext) string {
	nextCtx, result := ctx.With(v)
	if nextCtx == nil {
		return result
	}
	var sb strings.Builder
	for _, option := range v.options {
		if sb.Len() > 0 {
			sb.WriteString("|")
		}
		sb.WriteString("(?:" + option.ValueType.toRegex(*nextCtx) + ")")
	}
	return sb.String()
}

func (v *oneofValue) supportsChildren() bool {
	for _, option := range v.options {
		if option.ValueType.supportsChildren() {
			return true
		}
	}
	return false
}

type complexValue struct {
	textValue      valueType
	textValueKey   string
	structValue    valueType
	structValueKey string
}

func (v *complexValue) parse(builder valueBuilder, value string) error {
	log.Println("complexValue [parse]:", value)
	if v.textValueKey == "" {
		return v.textValue.parse(builder, value)
	}
	return v.textValue.parse(v.resolveBuilderFieldPath(builder, v.textValueKey), value)
}

func (v *complexValue) resolveBuilderFieldPath(builder valueBuilder, path string) valueBuilder {
	for _, part := range strings.Split(path, ".") {
		builder = builder.getFieldBuilder(part)
	}
	return builder
}

func (v *complexValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	log.Println("complexValue [parseChild]:", value)
	if v.structValueKey == "" {
		return v.structValue.parseChild(builder, value)
	}
	return v.structValue.parseChild(v.resolveBuilderFieldPath(builder, v.structValueKey), value)
}

func (v *complexValue) toRegex(ctx regexBuildContext) string {
	nextCtx, result := ctx.With(v)
	if nextCtx == nil {
		return result
	}
	return v.textValue.toRegex(*nextCtx)
}

func (v *complexValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return v.structValue.getChild(name, builder)
}

func (v *complexValue) supportsChildren() bool {
	return v.structValue.supportsChildren()
}

type delegatedValue struct {
	target valueType
}

func (v *delegatedValue) parse(builder valueBuilder, value string) error {
	return v.target.parse(builder, value)
}

func (v *delegatedValue) parseChild(builder valueBuilder, value string) (*nodeInfo, error) {
	return v.target.parseChild(builder, value)
}

func (v *delegatedValue) toRegex(ctx regexBuildContext) string {
	//TODO
	// return ".*?" //
	return v.target.toRegex(ctx)
}

func (v *delegatedValue) getChild(name string, builder valueBuilder) (valueType, valueBuilder, error) {
	return v.target.getChild(name, builder)
}

func (v *delegatedValue) supportsChildren() bool {
	return v.target.supportsChildren()
}

func newRegexBuildContext() regexBuildContext {
	return regexBuildContext{usage: map[valueType]int{}}
}

func (c regexBuildContext) With(currentType valueType) (*regexBuildContext, string) {
	if c.usage[currentType] > 0 {
		return nil, ".*?"
	}
	newMap := map[valueType]int{}
	for k, v := range c.usage {
		newMap[k] = v
	}
	newMap[currentType] = newMap[currentType] + 1
	return &regexBuildContext{usage: newMap}, ""
}
