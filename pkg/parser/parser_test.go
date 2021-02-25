package parser

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParserE2E(t *testing.T) {
	parser := NewParser()
	for _, tc := range testCases {
		println("Test:", tc.name)
		fullPath := "../../samples/" + tc.testFile
		file, err := os.Open(fullPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		got, err := parser.Parse2(file, NewFSResourceProvider(filepath.Dir(fullPath)))
		if err != nil {
			t.Fatalf("could not parse %v, %v", tc.testFile, err)
		}
		if !reflect.DeepEqual(got, tc.expected) {
			t.Errorf("%s[%s]\nGOT:  %+v \nWANT: %+v", tc.name, tc.testFile, got, tc.expected)
		}
	}
}

type testCase struct {
	name     string
	testFile string
	expected Node
}

var testCases = []testCase{
	{
		name:     "simple test",
		testFile: "simple/simple.dad",
		expected: Node{
			"someRoot": Node{
				"firstChild": "some long string value with spaces",
				"secondChild": Node{
					"nestedChild": 7,
				},
			},
		},
	},
	{
		name:     "embedded text test",
		testFile: "embedded_text/embedded_text.dad",
		expected: Node{
			"someJson": `{
	"martin": {
		"name": "Martin D'vloper",
		"job": "Developer"
	}
}`,
			"someYaml": `martin:
	name: Martin D'vloper
	job: Developer`,
			"someDadl": `[martin]
name Martin D'vloper
job Developer`,
			"someBrainfuck": `++++++++++[>+>+++>+++++++>+++++
+++++<<<<-]>>>++.>+.+++++++..++
+.<<++.>----.---.+++.++++++++.`,
			"someBash": `#!/bin/bash
echo "Hello Dadl!"`,
		},
	},
	{
		name:     "teleport test",
		testFile: "teleport/teleport.dad",
		expected: Node{
			"someRoot": Node{
				"firstChild": Node{
					"nestedChild": Node{
						"evenMoreNasted": "some value",
					},
				},
			},
		},
	},
	{
		name:     "import text file test",
		testFile: "import_text_file/import_text_file.dad",
		expected: Node{
			"someBrainfuck": `++++++++++[>+>+++>+++++++>+++++
+++++<<<<-]>>>++.>+.+++++++..++
+.<<++.>----.---.+++.++++++++.`,
		},
	},
	{
		name:     "custom types test",
		testFile: "custom_types/custom_types.dad",
		expected: Node{
			"sampleEnum":       "GET",
			"sampleInlineEnum": "OK",
			"sampleHostname":   "node1",
			"samplePort":       9042,
			"sampleAddress": Node{
				"host": "node1",
				"port": 9042,
			},
			"sampleAddresses": []interface{}{
				Node{
					"host": "node1",
					"port": 9042,
				},
				Node{
					"host": "node2",
					"port": 9042,
				},
				Node{
					"host": "node3",
					"port": 9042,
				},
			},
		},
	},
	{
		name:     "maps test",
		testFile: "maps/maps.dad",
		expected: Node{
			"structMap": Node{
				"firstKey": Node{
					"intValue":  7,
					"textValue": "some text value",
				},
				"secondKey": Node{
					"intValue": 14,
				},
				"thirdKey": Node{
					"textValue": "third",
				},
				"fourthKey": Node{
					"textValue": "fourth",
				},
				"fifthKey": Node{
					"textValue": "fifth",
				},
			},
			"coordinatesMap": Node{
				"key1": Node{
					"x": 2,
					"y": 3,
				},
				"key2": Node{
					"x": 5,
					"y": 7,
				},
			},
			"intMap": Node{
				"key1": 5,
				"key2": 7,
			},
			"stringMap": Node{
				"key1": "value1",
				"key2": "value2",
			},
		},
	},
	{
		name:     "import subtree test",
		testFile: "import_subtree/import_subtree.dad",
		expected: Node{
			"modules": Node{
				"firstModule": Node{
					"active": true,
					"name":   "First module",
				},
				"secondModule": Node{
					"active": false,
					"name":   "Second module",
				},
				"thirdModule": Node{
					"active": false,
					"name":   "Third module",
				},
			},
		},
	},
}
