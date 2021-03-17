### Simple schema and data file
Create sample schema file `sample.dads`

    @schema dadl 0.1

    [structure]
    someRoot
        firstChild string
        secondChild
            nestedChild int

with given schema then we can create data file `sample.dad`

    @schema ./sample.dads

    someRoot
        firstChild some long string value with spaces
        secondChild
            nestedChild 7  

### Embedding multiline text data `embedded_text_sample.dads`

    @schema dadl 0.1

    [structure]
    someJson string
    someYaml string
    someDadl string
    someBrainfuck string

Then you can just put multiline text data as a value

    @schema embedded_text_sample.dads

    someJson
        {
            "martin": {
                "name": "Martin D'vloper",
                "job": "Developer"
            }
        }
    someYaml
        martin:
            name: Martin D'vloper
            job: Developer
    someDadl
        [martin]
        name Martin D'vloper
        job Developer
    someBrainfuck
        ++++++++++[>+>+++>+++++++>+++++
        +++++<<<<-]>>>++.>+.+++++++..++
        +.<<++.>----.---.+++.++++++++.

### Teleporting to given node
Create sample schema file `teleport.dads`

    @schema dadl 0.1

    [structure]
    someRoot
        firstChild
            nestedChild
                evenMoreNasted string

Then to save yourself from indention hell you can just teleport to given node using braces []. Every line following teleport is considerd as a child of that node.

    @schema teleport.dads

    [someRoot.firstChild.nestedChild]     
    evenMoreNasted some value


## Embedding data from external file
It's possible to import external file as a node value. Create `import_text_file.dads'

    @schema dadl 0.1

    [structure]
    someBrainfuck string

Then create attachment file `brainfuck.bf`

    ++++++++++[>+>+++>+++++++>+++++
    +++++<<<<-]>>>++.>+.+++++++..++
    +.<<++.>----.---.+++.++++++++.

And create data file `import_file.dad`

    @schema import_text_file.dads

    [someBrainfuck << ./brainfuck.bf]     

## Custom types
Create `custom_types.dads'

    @schema dadl 0.1

    [types]
    httpVerb enum GET POST PUT PATCH DELETE
    hostname string \S+
    networkPort int 0..65535 #Network port number
    address formula <host hostname> ':' <port networkPort>
    addresses sequence address

    [structure]
    sampleEnum httpVerb
    sampleHostname hostname
    samplePort networkPort
    sampleNode address
    sampleNodes addresses

And create data file `custom_types.dad`

    @schema custom_types.dads
    
    sampleEnum GET
    sampleHostname node1
    samplePort 9042
    sampleNode node1:9042
    sampleNodes node1:9042 node2:9042 node2:9042

         