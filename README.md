# DADL Quickstart Guide
<img src="./.github/gopher.png" style="align: right" width="400">

DADL is a configuration language and a foundation for creating custom domain specific declarative languages. It simplifies splitting data across multiple files while allowing to verifiy data correctness and exporting it to other formats like JSON or YAML.

## Install
Binary downloads of dadl can be found on [the Releases page](https://github.com/dadlang/dadl/releases/latest).

Download the binary for your system, rename it to `dadl`, add it to your PATH and you are good to go!

## Example
Dadl is schema first language. File containing data can be parsed correctly only if schema file is provided.
Let's create simple schema file `config.dads`

    @schema dadl 0.1

    [types]
    hostname string `\S+`
    networkPort int 0..65535
    address formula <host hostname> ':' <port networkPort>

    [structure]
    cassandra
        nodes sequence[address]
        pass string

Then create the data file `config.dad`

    @schema config.dads
    
    [cassandra]
    nodes node1:9042 node2:9042 node2:9042
    pass admin123

Now we can preview data using `dadl print config.dad` to see content of the file.

    .
    └── cassandra
        ├── nodes
        │   ├── [0]
        │   │   ├── host: node1
        │   │   └── port: 9042
        │   ├── [1]
        │   │   ├── host: node2
        │   │   └── port: 9042
        │   └── [2]
        │       ├── host: node2
        │       └── port: 9042
        └── pass: admin123

Or export it to JSON using `dadl export config.dad`

    {
    "cassandra": {
        "nodes": [
        {
            "host": "node1",
            "port": 9042
        },
        {
            "host": "node2",
            "port": 9042
        },
        {
            "host": "node2",
            "port": 9042
        }
        ],
        "pass": "admin123"
    }

Or to YAML using `dadl export config.dad -f yaml`

    cassandra:
      nodes:
      - host: node1
        port: 9042
      - host: node2
        port: 9042
      - host: node2
        port: 9042
      pass: admin123


More complex examples can be found on [the Samples page](https://github.com/dadlang/dadl/blob/main/SAMPLES.md).

## Structural syntax
In simplest form dadl format is very similar to YAML with indentation used to describe hirarchy. Every line prefixed with whitespaces is considered a child of a first preceding line with lower number of whitespaces.

For example:

    root
        child-1
            grandchild-1-1 value1
        child-2
            grandchild-2-1 value2
            grandchild-2-2 value3

in the above example both child-1 and child-2 are considered a child of a root node while grandchild-2-1 and grandchild-2-2 are children of child-2 node.

Additionaly it's possible to limit the use of indention and "teleport" to given node:

    [root.child-1]
    grandchild-1-1 value1

    [root.child-2]
    grandchild-2-1 value2
    grandchild-2-2 value3

This sample definition is equvalent to the previous one that used indentation. After teleporting to the node with square brackets it's not required (but possible) to use indentation. Every following line is considerd to be a child  of the node defined in the square brackets.

## Schema definition
Dadl requires schema file to correctly parse data. Schema file uses exaclty the same sytax and concepts as data file. Schema file consists of only two root elements:

    types
    structure

where `types` contains list of custom type definions while `structure` contains a definition of the structure of the data file. Additionaly every dadl file must start with a schema information. For data file it's a name of the schema file, for schema file it's a constant definition:

    @schema dadl 0.1

Let's define a schema file with simple structure:

    @schema dadl 0.1

    [structure]
    someRoot
        firstChild string
        secondChild
            nestedChild int

Every node under the `structure` node consists of the node name, whitespace and node type name. When type is not provided it's considered to be a `struct` type. Both of those definitions are considered equal:

    node1
    node2 struct

In the last schema exemple we expect one root node called `someRoot` that's type of `struct`. This node has 2 children, first called `firstChild` of type `string` and second one called `secondChild` of type `struct` (lack of type definition defaults to `struct` type). Additionaly `secondChild` has one child of type `nestedChild` that type is `int`.

## Data file definition
Data file starts with schema definition the same way as a schema file but instead of constant `dadl` it should contain name of the schema file. Let's assume that we created schema file called `simple.dads` that contains sample schema definiton mentioned in previous chapter. Then we can create sample data file:

    @schema ./sample.dads

    someRoot
        firstChild some long string value with spaces
        secondChild
            nestedChild 7  

In that file we define that node `firstChild` contains string value `some long string value with spaces` while nestedChild node value is `7`.

## Node types definitions
Dadl supportes followind node types:

### **string**
String type expects any text value but can be limited by passing a regex that is enclosed between ` chars.

    sampleDef1 string

    sampleDef2 string `\S+`

    sampleValue1 some text with spaces

    sampleValue2 onlyNonWhitespaceCharacters

### **identifier**
Identifier is a textual type similar to string with that difference that it supports only characters: A-Z, a-z, 0-9, - and _.

    sampleDef identifier

    sampleValue someIdentifier

### **int**
Int is a numerical type that expects an integer value. It's possible to define a range of allowed values.
    
    sampleDef1 int

    sampleDef2 int 0..65535

    sampleValue 7

### **number**
Number is a numerical type that accepts any real number with `.` as decimal part separator

    sampleDef number

    sampleValue 3.14

### **bool**
Bool is a boolean type that accepts values: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False

    sampleDef bool

    sampleValue true

### **enum**
Enum is a enumeration type that expects only one of defined possible values.

    sampleDef enum OPTION1 OPTION2 OPTION3

    sampleValue OPTION1

It's possbile to map enum value to some other type like boolean or int.

    sampleDef1 enum[bool] YES[true] NO[false]

    sampleDef2 enum[int] CAR[1] TRUCK[2] MOTO[3]

    sampleValue1 YES
    sampleValue2 MOTO

### **formula**
Formula is a textual value that expects given sequence of tokens. Possible tokens are:

- variable definition - \<varName typeDef\>
- constant - 'constant value'
- optional block using square brackets []
<!-- -->

    sampleDef formula <host hostname> [':' <port networkPort>]

    sampleValue localhost:8080

### **sequence**
Sequence is a textual sequence of values separated with whitespace.

    sampleDef sequence[identifier]

    sampleValue VAL1 VAL2 VAL3 

### **list**
List is a structural type where every child is considered to be a list item of given type

    sample list[identifier]

    value
        VAL1
        VAL2
        VAL3

### **map**
Map is a structural type where every child contains a key to value mapping. Value is separated from the key with whitespace.

    sampleDef1 map[int]string

    sampleDef2 map[identifier]
        childValue int

    sampleValue1
        1 Some value for key 1
        2 Some value for key 1
        3 Some value for key 1

    sampleValue2
        key1
            childValue 1
        key2
            childValue 2

### **struct**
Struct is a structural nodes that dont have a textual value but can contain children. Lack of type name defaults also to `struct` type.

    sampleDef1
        key1 string
        key2 int

    sampleDef2 struct
        key1 string
        key2 int

    sampleValue
        key1 text value
        key2 8

### **oneof**
Oneof accepts value that matches one of the provided definitions. The first type that matches is used to parse the value. Possible values are separated with `|` char.

    sampleDef oneof[path|operation]

### **complex**
Complex types allowes to merge a simple textual type with structural type.

    sampleDef1 complex[int]struct
        child1 string
        child2 int

    sampleDef2 complex[int]...
        child1 string
        child2 int

By default textual part is stored undel child with name `value` and the structural part is stored under node `children`. It's possible to use `...` prefix both before textual or structural type to move the values under the root node. 

## Custom data types
In schema file it's possible to define custom data types. Typeas are defined the same way as in `structure`, the only difference is that instead of defining type for given node we define an alias that can be used to point to that type definition. 

    @schema dadl 0.1

    [types]
    hostname string `\S+`
    networkPort int 0..65535
    address formula <host hostname> ':' <port networkPort>

    [structure]
    sampleNode address
    samplePort networkPort

In this sample file we define 3 custom types:
- hostname - that extends string type but limits allowed values to non-whitespace characters
- networkPort - that extends int type and limits allowed values to range 0..65535 (including)
- address - that is type of formula which expects hostname definition followed by`:` constant and networkPort definition

After that we can use those custom types in `structure` definition or in definitions of other custom types.