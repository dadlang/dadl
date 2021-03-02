package parser

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"unicode"
)

//NewParser - creates new Parser instance
func NewParser() Parser {
	return Parser{}
}

type nodeInfo struct {
	valueType valueType
	builder   valueBuilder
}

type parseResult struct {
	root Node
}

type parseContext struct {
	schema         DadlSchema
	indentWeight   int
	parentNodeInfo *nodeInfo
	lastNodeInfo   *nodeInfo
}

var groupRe = regexp.MustCompile("^\\[(?P<treePath>[a-zA-Z0-9-_.]*)\\s*(?:<<\\s*(?P<importPath>.+))?\\]$")

func (p *Parser) Parse(reader io.Reader, resources ResourceProvider) (Node, error) {

	root := Node{}
	rootBuilder := &dynamicMapOrListValueBuilder{value: root}

	err := p.ParseWithBuilderAndSchema(reader, resources, rootBuilder, nil)
	if err != nil {
		return nil, err
	}
	return root, nil
}

func (p *Parser) ParseWithBuilderAndSchema(reader io.Reader, resources ResourceProvider, builder valueBuilder, schema DadlSchema) error {

	ctxByIndent := make([]*parseContext, 100)
	ctx := &parseContext{schema: schema}

	if schema != nil {
		ctx.parentNodeInfo = &nodeInfo{
			valueType: schema.getRoot(),
			builder:   builder,
		}
	}

	var err error

	ctxByIndent = make([]*parseContext, 100)
	ctxByIndent[0] = ctx

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		line := strings.TrimRight(scanner.Text(), "\t \n")

		indentWeight := calcIndentWeight(line)

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			//	println("Process group:", line)
			ctx, err = p.processGroup(line, ctx, builder, resources)
			if err != nil {
				return err
			}
			ctxByIndent = make([]*parseContext, 100)
			ctxByIndent[0] = ctx
		} else if strings.TrimSpace(line) != "" {
			if strings.HasPrefix(line, "#") {
				//	fmt.Println("skip comment:", line)
			} else if strings.HasPrefix(line, "@") {
				err := p.parseMagic(ctx, builder, line, resources)
				if err != nil {
					return err
				}
			} else {
				if indentWeight > ctx.indentWeight {
					nextParentInfo := ctx.lastNodeInfo
					ctx = &parseContext{
						schema:         ctx.schema,
						indentWeight:   indentWeight,
						parentNodeInfo: nextParentInfo}
					ctxByIndent[indentWeight] = ctx
				} else if indentWeight < ctx.indentWeight {

					for j := indentWeight; j >= 0; j-- {
						if ctxByIndent[j] != nil {
							ctx = ctxByIndent[j]
							break
						}
					}
				}

				ctx.lastNodeInfo, err = ctx.parentNodeInfo.valueType.parseChild(ctx.parentNodeInfo.builder, line)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *Parser) processGroup(line string, ctx *parseContext, rootBuilder valueBuilder, resources ResourceProvider) (*parseContext, error) {
	//fmt.Println(line)
	match := groupRe.FindStringSubmatch(line)
	if match != nil {
		result := make(map[string]string)
		for i, name := range groupRe.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
		treePath := result["treePath"]
		importPath := result["importPath"]
		if ctx.schema == nil {
			return nil, errors.New("Missing schema info")
		}
		schemaNode, valueBuilder, err := ctx.schema.getNode(treePath, rootBuilder)
		if err != nil {
			return nil, err
		}

		schemaNode.parse(valueBuilder, "")

		if importPath != "" {
			file, err := resources.GetResource(importPath)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			if _, ok := schemaNode.(*stringValue); ok {
				data, err := ioutil.ReadAll(file)
				if err != nil {
					return nil, err
				}
				valueBuilder.setSimpleValue(string(data))
			} else {
				err := p.ParseWithBuilderAndSchema(file, resources, valueBuilder, &dadlSchemaImpl{root: schemaNode})
				if err != nil {
					return nil, err
				}
			}
		}
		return &parseContext{schema: ctx.schema, parentNodeInfo: &nodeInfo{valueType: schemaNode, builder: valueBuilder}, indentWeight: 0}, nil
	}
	return nil, errors.New("Invalid group definition")
}

func calcIndentWeight(line string) int {
	for idx, c := range line {
		if !unicode.IsSpace(c) {
			return idx
		}
	}
	return len(line)
}

func (p *Parser) parseMagic(ctx *parseContext, rootBuilder valueBuilder, line string, resources ResourceProvider) error {
	if strings.HasPrefix(line, "@schema ") {
		var err error
		parts := strings.Split(line[8:], " ")

		if ctx.schema != nil {
			//TODO compares chema
			return nil
		}

		ctx.schema, err = parseSchema(parts[0], resources)
		if err != nil {
			return err
		}

		if len(parts) == 2 && strings.HasPrefix(parts[1], "[") && strings.HasSuffix(parts[1], "]") {
			valueType, _, err := ctx.schema.getNode(parts[1][1:len(parts[1])-1], rootBuilder)
			if err != nil {
				return err
			}
			ctx.schema = &dadlSchemaImpl{root: valueType}
			ctx.parentNodeInfo = &nodeInfo{valueType: valueType, builder: rootBuilder}
		} else {
			ctx.parentNodeInfo = &nodeInfo{valueType: ctx.schema.getRoot(), builder: rootBuilder}
		}
		return nil
	} else {
		return errors.New("Unknown magic line: " + line)
	}
}

//Parser - parses DADL files
type Parser struct {
	// schema DadlSchema
}

//Node alias for map of string to interface
type Node = map[string]interface{}

//NodeParser parses node.
type NodeParser interface {
	parse(ctx *parseContext, line string) error
}
