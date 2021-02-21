package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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

var groupRe = regexp.MustCompile("^\\[(?P<treePath>[a-zA-Z0-9-_.]*)\\s*(?:<<\\s*(?P<importPath>.+))?\\]$")

//Parse - scans given stream
func (p *Parser) Parse(reader io.Reader, resources ResourceProvider) (Node, error) {
	tree := Node{}
	ctxByIndent := make([]*parseContext, 100)
	ctx := &parseContext{parent: tree}
	if p.schema != nil {
		ctx.parentSchema = p.schema.getRoot()
	}
	//fmt.Printf("CTX -> %x\n", ctx)
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
			ctx, err = p.processGroup(line, tree, resources)
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
					println("Indent found: " + line)
					nextParent := ctx.last
					if nextParent == nil {
						nextParent = ctx.parent
					}
					ctx = &parseContext{parent: nextParent, parentSchema: ctx.lastSchema, indentWeight: indentWeight}
					fmt.Printf("CTX[%v] -> %v\n", indentWeight, ctx)
					ctxByIndent[indentWeight] = ctx
				} else if indentWeight < ctx.indentWeight {
					println("Find by indent: ", indentWeight)

					for j := indentWeight; j >= 0; j-- {
						if ctxByIndent[j] != nil {
							println("Return by indent: ", j)
							ctx = ctxByIndent[j]
							break
						}
					}
					fmt.Printf("CTX -> %+v\n", ctx)
				}

				println("Parse line:", line)
				parser, err := ctx.parentSchema.childParser()
				if err != nil {
					return nil, err
				}
				if err := parser.parse(ctx, line); err != nil {
					return nil, err
				}

			}
		}
	}
	return tree, nil
}

type parseContext struct {
	// parser       NodeParser
	parentSchema SchemaNode
	lastSchema   SchemaNode
	parent       Node
	last         Node
	indentWeight int
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
		parts := strings.Split(treePath, ".")
		node := tree
		var nodeParent map[string]interface{}
		var key string
		for _, part := range parts {
			nodeParent = node
			key = part
			next, ok := node[part]
			if !ok {
				next = map[string]interface{}{}
				node[part] = next
			}
			node = next.(map[string]interface{})
		}
		if p.schema == nil {
			return nil, errors.New("Missing schema info for path: " + treePath)
		}
		schemaNode, err := p.schema.getNode(treePath)
		if err != nil {
			return nil, err
		}

		if importPath != "" {
			file, err := resources.GetResource(importPath)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			parser := newParserWithSchema(&dadlSchemaImpl{root: schemaNode})
			value, err := parser.Parse(file, resources)
			if err != nil {
				return nil, err
			}
			if schemaNode.isSimple() {
				nodeParent[key] = value[key]
			} else {
				for k, v := range value {
					node[k] = v
				}
			}
		}

		// childParser, err := schemaNode.childParser()
		// if err != nil {
		// 	return parseContext{}, err
		// }

		return &parseContext{parent: node, parentSchema: schemaNode, indentWeight: 0}, nil
	}
	return &parseContext{}, nil
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
		p.schema, err = parseSchema(parts[0], resources)
		if err != nil {
			return err
		}

		ctx.parentSchema = p.schema.getRoot()
		if len(parts) == 2 && strings.HasPrefix(parts[1], "[") && strings.HasSuffix(parts[1], "]") {
			ctx.parentSchema, err = p.schema.getNode(parts[1][1 : len(parts[1])-1])
			if err != nil {
				return err
			}
		}
	} else {
		println("UNKNOW MAGIC")
	}
	return nil
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
