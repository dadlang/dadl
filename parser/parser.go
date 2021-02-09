package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"unicode"
)

//NewParser - creates new Parser instance
func NewParser(schema DadlSchema) Parser {
	return Parser{schema: schema}
}

var groupRe = regexp.MustCompile("^\\[(?P<treePath>[a-zA-Z0-9-_.]*)\\s*(?:<<\\s*(?P<importPath>.+))?\\]$")

//Parse - scans given stream
func (p *Parser) Parse(reader io.Reader) (Node, error) {
	tree := Node{}
	ctxByIndent := make([]parseContext, 100)
	ctx := parseContext{parent: tree, parser: p.schema.rootParser()}
	var err error

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		line := strings.TrimRight(scanner.Text(), "\t \n")
		println("Process line:", line)

		indentWeight := calcIndentWeight(line)

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			ctx, err = p.processGroup(line, tree)
			if err != nil {
				return nil, err
			}
			ctxByIndent = make([]parseContext, 100)
			ctxByIndent[0] = ctx
		} else if strings.TrimSpace(line) != "" {
			if strings.HasPrefix(line, "#") {
				fmt.Println("skip comment:", line)
			} else if strings.HasPrefix(line, "@") {
				fmt.Println("magic ->", line)
			} else {
				if indentWeight > ctx.indentWeight {
					println("Indent found: " + line)
					next, err := ctx.parser.childParser()
					if err != nil {
						return nil, err
					}
					ctx = parseContext{parser: next, parent: ctx.current, indentWeight: indentWeight}
					ctxByIndent[indentWeight] = ctx
				} else if indentWeight < ctx.indentWeight {
					println("Find by indent: ", indentWeight)
					ctx = ctxByIndent[indentWeight]
				}

				if err := ctx.parser.parse(&ctx, line); err != nil {
					return nil, err
				}

			}
		}
	}
	return tree, nil
}

type parseContext struct {
	parser       NodeParser
	parent       Node
	current      Node
	indentWeight int
}

func (p *Parser) processGroup(line string, tree map[string]interface{}) (parseContext, error) {
	fmt.Println(line)
	match := groupRe.FindStringSubmatch(line)
	if match != nil {
		result := make(map[string]string)
		for i, name := range groupRe.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
		treePath := result["treePath"]
		parts := strings.Split(treePath, ".")
		node := tree
		for _, part := range parts {
			next, ok := node[part]
			if !ok {
				next = map[string]interface{}{}
				node[part] = next
			}
			node = next.(map[string]interface{})
		}
		groupParser, err := p.schema.parser(treePath)
		if err != nil {
			return parseContext{}, err
		}

		return parseContext{parent: node, parser: groupParser, indentWeight: 0}, nil
	}
	return parseContext{}, nil
}

func calcIndentWeight(line string) int {
	for idx, c := range line {
		if !unicode.IsSpace(c) {
			return idx
		}
	}
	return len(line)
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
	childParser() (NodeParser, error)
}

// type genericKeyValueParser struct{}

// func (p *genericKeyValueParser) parse(ctx *parseContext, value string) error {
// 	println("KV: ", value)
// 	value = strings.TrimSpace(value)
// 	vals := strings.SplitN(value, " ", 2)
// 	if len(vals) > 1 {
// 		ctx.parent[vals[0]] = vals[1]
// 	}
// 	return nil
// }

// func (p *genericKeyValueParser) childParser() (NodeParser, error) {
// 	return nil, nil
// }

// type tokenSequenceParser struct{}

// func (p *tokenSequenceParser) parse(ctx *parseContext, value string) error {
// 	value = strings.TrimSpace(value)
// 	vals := strings.SplitN(value, " ", 3)
// 	if len(vals) > 2 {
// 		ctx.parent[vals[0]] = Node{
// 			"type": vals[1],
// 			"desc": vals[2][1:],
// 		}
// 	}
// 	return nil
// }

// func (p *tokenSequenceParser) childParser() (NodeParser, error) {
// 	return nil, nil
// }
