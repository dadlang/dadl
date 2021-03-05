package parser

type metaBuilder struct {
	meta map[string]interface{}
}

func (b *metaBuilder) getMeta(name string) interface{} {
	return b.meta[name]
}

func (b *metaBuilder) setMeta(name string, value interface{}) {
	if b.meta == nil {
		b.meta = map[string]interface{}{}
	}
	b.meta[name] = value
}

type valueBuilder interface {
	getSimpleValue() interface{}
	setSimpleValue(value interface{})
	getFieldBuilder(name string) valueBuilder
	getListItemBuilder() valueBuilder
	setMeta(name string, value interface{})
	getMeta(name string) interface{}
}

type dynamicMapOrListValueBuilder struct {
	metaBuilder
	value interface{}
}

func (b *dynamicMapOrListValueBuilder) getSimpleValue() interface{} {
	panic("Not supported")
}

func (b *dynamicMapOrListValueBuilder) setSimpleValue(value interface{}) {
	panic("Not supported")
}

func (b *dynamicMapOrListValueBuilder) getFieldBuilder(name string) valueBuilder {
	if b.value == nil {
		b.value = map[string]interface{}{}
	}
	return &itemInMapValueBuilder{
		parent:    b.value.(map[string]interface{}),
		fieldName: name,
	}
}

func (b *dynamicMapOrListValueBuilder) getListItemBuilder() valueBuilder {
	if b.value == nil {
		b.value = []interface{}{}
	}
	idx := len(b.value.([]interface{}))
	b.value = append(b.value.([]interface{}), nil)
	return &itemInListValueBuilder{
		parent: b.value.([]interface{}),
		idx:    idx,
	}
}

type itemInMapValueBuilder struct {
	metaBuilder
	parent    map[string]interface{}
	fieldName string
}

func (b *itemInMapValueBuilder) getSimpleValue() interface{} {
	return b.parent[b.fieldName]
}

func (b *itemInMapValueBuilder) setSimpleValue(value interface{}) {
	b.parent[b.fieldName] = value
}

func (b *itemInMapValueBuilder) getFieldBuilder(name string) valueBuilder {
	if b.parent[b.fieldName] == nil {
		b.parent[b.fieldName] = map[string]interface{}{}
	}
	return &itemInMapValueBuilder{
		parent:    b.parent[b.fieldName].(map[string]interface{}),
		fieldName: name,
	}
}

func (b *itemInMapValueBuilder) getListItemBuilder() valueBuilder {
	if b.parent[b.fieldName] == nil {
		b.parent[b.fieldName] = []interface{}{}
	}
	idx := len(b.parent[b.fieldName].([]interface{}))
	b.parent[b.fieldName] = append(b.parent[b.fieldName].([]interface{}), nil)
	return &itemInListValueBuilder{
		parent: b.parent[b.fieldName].([]interface{}),
		idx:    idx,
	}
}

type itemInListValueBuilder struct {
	metaBuilder
	parent []interface{}
	idx    int
}

func (b *itemInListValueBuilder) getSimpleValue() interface{} {
	return b.parent[b.idx]
}

func (b *itemInListValueBuilder) setSimpleValue(value interface{}) {
	b.parent[b.idx] = value
}

func (b *itemInListValueBuilder) getFieldBuilder(name string) valueBuilder {
	if b.parent[b.idx] == nil {
		b.parent[b.idx] = map[string]interface{}{}
	}
	return &itemInMapValueBuilder{
		parent:    b.parent[b.idx].(map[string]interface{}),
		fieldName: name,
	}
}

func (b *itemInListValueBuilder) getListItemBuilder() valueBuilder {
	if b.parent[b.idx] == nil {
		b.parent[b.idx] = []interface{}{}
	}
	idx := len(b.parent[b.idx].([]interface{}))
	b.parent[b.idx] = append(b.parent[b.idx].([]interface{}), nil)
	return &itemInListValueBuilder{
		parent: b.parent[b.idx].([]interface{}),
		idx:    idx,
	}
}
