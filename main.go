package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/althink/dadl-core/parser"
)

func main() {
	file, err := os.Open("./sample/project.dad")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	schema, _ := parser.ParseSchema()
	parser := parser.NewParser(schema)
	tree, err := parser.Parse(file)

	if err != nil {
		println("Err: ", err.Error())
		return
	}

	j, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(j))
}
