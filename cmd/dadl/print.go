package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dadlang/dadl/pkg/parser"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

func init() {
	rootCmd.AddCommand(printCmd)
}

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Prints data from given file",
	Long:  `Prints data from given file.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires optional path and a file name")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			printHandler(args[0], ".")
		} else {
			printHandler(args[1], args[0])
		}
	},
}

func printHandler(filePath string, treePath string) {

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	p := parser.NewParser()
	tree, err := p.Parse(file, parser.NewFSResourceProvider(filepath.Dir(filePath)))

	if err != nil {
		println("Err: ", err.Error())
		return
	}

	tree, _ = filterTree(tree, treePath)
	printTree(tree, treePath)
}

func filterTree(root map[string]interface{}, filterPath string) (map[string]interface{}, error) {
	if filterPath == "." {
		return root, nil
	}
	node := root
	pathElements := strings.Split(filterPath, ".")
	for _, pathElement := range pathElements {
		node = node[pathElement].(map[string]interface{})
	}
	return node, nil
}

func printTree(root map[string]interface{}, rootName string) {
	tree := treeprint.NewWithRoot(rootName)
	buildMapChildren(tree, root)
	fmt.Println(tree.String())
}

func buildMapChildren(tree treeprint.Tree, children map[string]interface{}) {
	keys := make([]string, 0)
	for k := range children {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		addChild(tree, key, children[key])
	}
}

func buildSliceChildren(tree treeprint.Tree, children []interface{}) {
	for key, value := range children {
		addChild(tree, fmt.Sprintf("[%v]", key), value)
	}
}

func addChild(tree treeprint.Tree, key string, value interface{}) {
	if asMap, ok := value.(map[string]interface{}); ok {
		branch := tree.AddBranch(key)
		buildMapChildren(branch, asMap)
	} else if asSlice, ok := value.([]interface{}); ok {
		branch := tree.AddBranch(key)
		buildSliceChildren(branch, asSlice)
	} else if asString, ok := value.(string); ok {
		if strings.Contains(asString, "\n") {
			tree.AddNode(key + ":\n" + asString)
		} else {
			tree.AddNode(key + ": " + asString)
		}
	} else if isInt, ok := value.(int); ok {
		tree.AddNode(fmt.Sprintf("%s: %v", key, isInt))
	}
}
