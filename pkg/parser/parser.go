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

func newParserWithSchema(schema DadlSchema) Parser {
	return Parser{schema: schema}
}

type nodeInfo struct {
	valueType valueType
	builder   valueBuilder
}

type parseResult struct {
	root Node
}

type parseContext struct {
	// parser       NodeParser
	parent         Node
	last           Node
	indentWeight   int
	parentNodeInfo *nodeInfo
	lastNodeInfo   *nodeInfo
}

var groupRe = regexp.MustCompile("^\\[(?P<treePath>[a-zA-Z0-9-_.]*)\\s*(?:<<\\s*(?P<importPath>.+))?\\]$")

func (p *Parser) Parse(reader io.Reader, resources ResourceProvider) (Node, error) {
	//root := Node{}
	result := parseResult{root: Node{}}

	ctxByIndent := make([]*parseContext, 100)
	ctx := &parseContext{parent: result.root}

	if p.schema != nil {
		ctx.parentNodeInfo = &nodeInfo{
			valueType: p.schema.getRoot(),
			builder: &delegatedValueBuilder{
				set: func(value interface{}) {
					panic("not supported")
				},
				get: func() interface{} {
					return result.root
				},
			},
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
			ctx, err = p.processGroup(line, result.root, resources)
			if err != nil {
				return nil, err
			}
			ctxByIndent = make([]*parseContext, 100)
			ctxByIndent[0] = ctx
		} else if strings.TrimSpace(line) != "" {
			if strings.HasPrefix(line, "#") {
				//	fmt.Println("skip comment:", line)
			} else if strings.HasPrefix(line, "@") {
				//		fmt.Println("magic ->", line)
				err := p.parseMagic(ctx, line, resources)
				if err != nil {
					return nil, err
				}
			} else {
				if indentWeight > ctx.indentWeight {
					//		println("Indent found: " + line)
					nextParentInfo := ctx.lastNodeInfo
					ctx = &parseContext{
						indentWeight:   indentWeight,
						parentNodeInfo: nextParentInfo}
					//		fmt.Printf("CTX[%v] -> %v\n", indentWeight, ctx)
					ctxByIndent[indentWeight] = ctx
				} else if indentWeight < ctx.indentWeight {
					//		println("Find by indent: ", indentWeight)

					for j := indentWeight; j >= 0; j-- {
						if ctxByIndent[j] != nil {
							//				println("Return by indent: ", j)
							ctx = ctxByIndent[j]
							break
						}
					}
					//		fmt.Printf("CTX -> %+v\n", ctx)
				}

				// println("Parse line:", line)
				ctx.lastNodeInfo, err = ctx.parentNodeInfo.valueType.parseChild(ctx.parentNodeInfo.builder, line)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return result.root, nil
}

func (p *Parser) processGroup(line string, tree map[string]interface{}, resources ResourceProvider) (*parseContext, error) {
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
		if p.schema == nil {
			return nil, errors.New("Missing schema info")
		}
		schemaNode, valueBuilder, err := p.schema.getNode(treePath, &delegatedValueBuilder{
			set: func(value interface{}) {
				panic("not supported")
			},
			get: func() interface{} {
				return tree
			},
		})
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
				valueBuilder.setValue(string(data))
			} else {
				parser := newParserWithSchema(&dadlSchemaImpl{root: schemaNode})
				value, err := parser.Parse(file, resources)
				if err != nil {
					return nil, err
				}
				valueBuilder.setValue(value)
			}

		}

		return &parseContext{parentNodeInfo: &nodeInfo{valueType: schemaNode, builder: valueBuilder}, indentWeight: 0}, nil
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

func (p *Parser) parseMagic(ctx *parseContext, line string, resources ResourceProvider) error {
	if strings.HasPrefix(line, "@schema ") {
		var err error
		parts := strings.Split(line[8:], " ")
		p.schema, err = parseSchema2(parts[0], resources)
		if err != nil {
			return err
		}

		rootBuilder := &delegatedValueBuilder{
			set: func(value interface{}) {
				panic("not supported")
			},
			get: func() interface{} {
				return ctx.parent
			},
		}

		if len(parts) == 2 && strings.HasPrefix(parts[1], "[") && strings.HasSuffix(parts[1], "]") {
			valueType, _, err := p.schema.getNode(parts[1][1:len(parts[1])-1], rootBuilder)
			if err != nil {
				return err
			}
			ctx.parentNodeInfo = &nodeInfo{valueType: valueType, builder: rootBuilder}
		} else {
			ctx.parentNodeInfo = &nodeInfo{valueType: p.schema.getRoot(), builder: rootBuilder}
		}
		return nil
	} else {
		return errors.New("Unknown magic line: " + line)
	}
}

//Parser - parses DADL files
type Parser struct {
	schema DadlSchema
}

//Node alias for map of string to interface
type Node = map[string]interface{}

//NodeParser parses node.
type NodeParser interface {
	parse(ctx *parseContext, line string) error
}
