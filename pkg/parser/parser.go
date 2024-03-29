package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

//NewParser - creates new Parser instance
func NewParser() Parser {
	return Parser{}
}

//ParseError describes a parsing error
type ParseError interface {
	error
	GetLine() int
	GetColumn() int
	GetReason() string
}

//DefaultParseError default ParseError
type defaultParseError struct {
	line   int
	column int
	reason string
}

func newParseError(line int, col int, reason string) error {
	return defaultParseError{line, col, reason}
}

func (e defaultParseError) Error() string {
	return fmt.Sprintf("Parse error [line: %v, col: %v]: %v", e.line, e.column, e.reason)
}

func (e defaultParseError) GetLine() int {
	return e.line
}

func (e defaultParseError) GetColumn() int {
	return e.column
}

func (e defaultParseError) GetReason() string {
	return e.reason
}

type nodeInfo struct {
	valueType valueType
	builder   valueBuilder
	valueMeta *valueMeta
}

type parseContext struct {
	schema         DadlSchema
	indentWeight   int
	parentNodeInfo *nodeInfo
	lastNodeInfo   *nodeInfo
	lineNo         int
}

var groupRe = regexp.MustCompile(`^\[(?P<treePath>[a-zA-Z0-9-_.$]*)\s*(?:<\s*(?P<importPath>.+))?\]$`)

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

	ctx := &parseContext{schema: schema}

	if schema != nil {
		ctx.parentNodeInfo = &nodeInfo{
			valueType: schema.getRoot(),
			builder:   builder,
		}
	}

	var err error
	lineNo := 1

	//TODO
	ctxByIndent := make([]*parseContext, 100)
	ctxByIndent[0] = ctx

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		line := strings.TrimRight(scanner.Text(), "\t \n")
		ctx.lineNo = lineNo

		indentWeight := calcIndentWeight(line)

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			log.Println("process group:", line)
			ctx, err = p.processGroup(line, ctx, builder, resources)
			if err != nil {
				return err
			}
			ctxByIndent = make([]*parseContext, 100)
			ctxByIndent[0] = ctx
		} else if strings.TrimSpace(line) != "" {
			if strings.HasPrefix(line, "#") {
				log.Println("skip comment:", line)
			} else if strings.HasPrefix(line, "@") {
				err := p.parseMagic(ctx, builder, line, parseMetadata{lineNo: ctx.lineNo, colNo: 0}, resources)
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
				ctx.lineNo = lineNo

				ctx.lastNodeInfo, err = ctx.parentNodeInfo.valueType.parseChild(ctx.parentNodeInfo.builder, line, ctx.parentNodeInfo.valueMeta, parseMetadata{lineNo: ctx.lineNo, colNo: 0})
				if err != nil {
					return err
				}
			}
		}
		lineNo++
	}
	return nil
}

func (p *Parser) processGroup(line string, ctx *parseContext, rootBuilder valueBuilder, resources ResourceProvider) (*parseContext, error) {
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
			return nil, errors.New("missing schema info")
		}

		var valueMeta *valueMeta
		var schemaNode valueType
		var valueBuilder valueBuilder
		var err error

		if importPath != "" {
			paths, err := resources.FindResources(importPath)
			if err != nil {
				return nil, err
			}
			log.Println("Resources: ", resources)

			if len(paths) == 0 {
				return nil, fmt.Errorf("no file matches given path: %s", importPath)
			}

			for _, path := range paths {
				file, err := resources.GetResource(path)
				if err != nil {
					return nil, err
				}
				defer file.Close()

				_, fileName := filepath.Split(path)

				targetPath := treePath
				if strings.HasSuffix(treePath, "._") {
					targetPath = strings.TrimRight(treePath, "_") + strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
				}

				schemaNode, valueBuilder, err = ctx.schema.getNode(targetPath, rootBuilder, parseMetadata{lineNo: ctx.lineNo, colNo: 0})
				if err != nil {
					return nil, err
				}

				if _, ok := schemaNode.(*stringValue); ok {
					data, err := ioutil.ReadAll(file)
					if err != nil {
						return nil, err
					}
					valueBuilder.setSimpleValue(string(data))
				} else {
					err := p.ParseWithBuilderAndSchema(file, resources.ForResource(path), valueBuilder, &dadlSchemaImpl{root: schemaNode})
					if err != nil {
						return nil, err
					}
				}
			}
		} else {
			schemaNode, valueBuilder, err = ctx.schema.getNode(treePath, rootBuilder, parseMetadata{lineNo: ctx.lineNo, colNo: 0})
			if err != nil {
				return nil, err
			}

			valueMeta, err = schemaNode.parse(valueBuilder, "", parseMetadata{lineNo: ctx.lineNo, colNo: 0})
			if err != nil {
				return nil, err
			}
		}
		return &parseContext{schema: ctx.schema, parentNodeInfo: &nodeInfo{valueType: schemaNode, builder: valueBuilder, valueMeta: valueMeta}, indentWeight: 0}, nil
	}
	return nil, errors.New("invalid group definition")
}

func calcIndentWeight(line string) int {
	for idx, c := range line {
		if !unicode.IsSpace(c) {
			return idx
		}
	}
	return len(line)
}

func (p *Parser) parseMagic(ctx *parseContext, rootBuilder valueBuilder, line string, meta parseMetadata, resources ResourceProvider) error {
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
			tmpRootBuilder := &dynamicMapOrListValueBuilder{value: Node{}}
			valueType, _, err := ctx.schema.getNode(parts[1][1:len(parts[1])-1], tmpRootBuilder, parseMetadata{lineNo: ctx.lineNo, colNo: 0})
			if err != nil {
				return err
			}
			ctx.schema = &dadlSchemaImpl{root: valueType}
			ctx.parentNodeInfo = &nodeInfo{valueType: valueType, builder: rootBuilder}
		} else {
			ctx.parentNodeInfo = &nodeInfo{valueType: ctx.schema.getRoot(), builder: rootBuilder}
		}
		return nil
	}
	return newParseError(meta.lineNo, meta.colNo, "Unknown magic line: "+line)
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
