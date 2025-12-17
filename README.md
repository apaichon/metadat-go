# MetaDat Go Library

A Go library for reading and writing MetaDat format files. MetaDat is a schema-first data serialization format that separates metadata from data for enhanced type safety and compression.

## Features

- **Read and Write Support**: Full support for parsing and generating MetaDat format
- **Struct Serialization**: Convert Go structs directly to MetaDat format
- **Single and Separated Files**: Support for both combined and separated schema/data files
- **Type Safety**: Schema validation and type checking
- **JSON Conversion**: Convert between JSON and MetaDat formats
- **Array Size Handling**: Automatically reads array sizes from MetaDat format declarations
- **High Performance**: Efficient parsing and serialization

## Installation

```bash
go get github.com/apaichon/metadat-go
```

### Version Information

```bash
# Check CLI version
./metadat -version

# In Go code
fmt.Println("Version:", metadat.Version)

# Get detailed version info
versionInfo := metadat.GetVersionInfo()
fmt.Printf("Library: %s v%s\n", versionInfo["name"], versionInfo["version"])
```

## Quick Start

### Writing a Struct to MetaDat

```go
package main

import (
    "fmt"
    "log"
    "github.com/apaichon/metadat-go"
)

type User struct {
    Name   string `json:"name"`
    Age    int    `json:"age"`
    Email  string `json:"email"`
    Active bool   `json:"active"`
}

func main() {
    user := User{
        Name:   "Alice Johnson",
        Age:    28,
        Email:  "alice@example.com",
        Active: true,
    }

    writer := metadat.NewWriter()
    content, err := writer.WriteStruct(user)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(content)
    // Output:
    // meta
    //     active: bool
    //     age: int
    //     email: string
    //     name: string
    // data
    //     active:
    //         true
    //     age:
    //         28
    //     email:
    //         alice@example.com
    //     name:
    //         Alice Johnson
}
```

### Parsing MetaDat

```go
parser := metadat.NewParser()
data, err := parser.ParseMetaDat(metadatContent)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Name: %s\n", data["name"])
fmt.Printf("Age: %d\n", data["age"])
```

### Working with Arrays and Objects

```go
type Product struct {
    ID       int      `json:"id"`
    Name     string   `json:"name"`
    Price    float64  `json:"price"`
    Tags     []string `json:"tags"`
    Details  Details  `json:"details"`
}

type Details struct {
    Weight     float32 `json:"weight"`
    Dimensions string  `json:"dimensions"`
}

product := Product{
    ID:    1001,
    Name:  "Laptop Pro",
    Price: 1299.99,
    Tags:  []string{"electronics", "computers", "premium"},
    Details: Details{
        Weight:     2.5,
        Dimensions: "30x20x2cm",
    },
}

writer := metadat.NewWriter()
content, err := writer.WriteStruct(product)
```

### Separated Files Mode

```go
// Write to separate schema and data files
writer := metadat.NewWriter()
err := writer.WriteStructToFiles(user, "schema.metadat", "data.metadat")

// Read from separate files
parser := metadat.NewParser()
data, err := parser.ParseFromFiles("schema.metadat", "data.metadat")
```

### JSON Conversion

```go
// JSON to MetaDat
jsonStr := `{"name": "Bob", "age": 30, "active": true}`
metadatStr, err := metadat.ConvertJSONToMetaDat(jsonStr)

// MetaDat to JSON
jsonResult, err := metadat.ConvertMetaDatToJSON(metadatStr)
```

## API Reference

### Writer

#### `NewWriter() *Writer`
Creates a new MetaDat writer instance.

#### `WriteStruct(v interface{}) (string, error)`
Converts a Go struct to MetaDat format (single file).

#### `WriteMetaDat(data map[string]interface{}) (string, error)`
Writes a map to MetaDat format using the current schema.

#### `WriteSeparated(data map[string]interface{}) (schema, data string, err error)`
Returns schema and data as separate strings.

#### `WriteToFiles(data map[string]interface{}, schemaFile, dataFile string) error`
Writes schema and data to separate files.

#### `WriteStructToFile(v interface{}, filename string) error`
Writes a struct to a single MetaDat file.

#### `WriteStructToFiles(v interface{}, schemaFile, dataFile string) error`
Writes a struct to separate schema and data files.

### Parser

#### `NewParser() *Parser`
Creates a new MetaDat parser instance.

#### `ParseMetaDat(content string) (map[string]interface{}, error)`
Parses a complete MetaDat format string.

#### `ParseFromFiles(schemaFile, dataFile string) (map[string]interface{}, error)`
Parses MetaDat from separate schema and data files.

#### `ParseSchema(schemaContent string) error`
Parses only the schema definition.

#### `ParseData(dataContent string) (map[string]interface{}, error)`
Parses data using the current schema.

### Schema

#### `InferSchemaFromStruct(v interface{}) (Schema, error)`
Infers a MetaDat schema from a Go struct.

#### `InferSchemaFromJSON(data interface{}) Schema`
Infers a MetaDat schema from JSON data.

#### `ValidateData(data map[string]interface{}) error`
Validates data against the schema.

## Examples

### Complex Nested Structure

```go
type Company struct {
    Name      string     `json:"name"`
    Founded   int        `json:"founded"`
    Employees []Employee `json:"employees"`
}

type Employee struct {
    ID     int     `json:"id"`
    Name   string  `json:"name"`
    Role   string  `json:"role"`
    Salary float64 `json:"salary"`
}

company := Company{
    Name:    "TechCorp",
    Founded: 2010,
    Employees: []Employee{
        {ID: 1, Name: "Alice", Role: "CEO", Salary: 150000},
        {ID: 2, Name: "Bob", Role: "CTO", Salary: 140000},
        {ID: 3, Name: "Carol", Role: "Developer", Salary: 90000},
    },
}

writer := metadat.NewWriter()
content, err := writer.WriteStruct(company)
```

Output:
```
meta
    employees: {id:int|name:string|role:string|salary:float64}[]
    founded: int
    name: string
data
    employees[3]:
        1|Alice|CEO|150000
        2|Bob|CTO|140000
        3|Carol|Developer|90000
    founded:
        2010
    name:
        TechCorp
```

### Custom Schema Definition

```go
// Define schema manually
schema := metadat.Schema{
    Fields: map[string]metadat.FieldType{
        "userId": {Type: "string"},
        "settings": {
            Type: "object",
            ObjectFields: map[string]metadat.FieldType{
                "theme": {Type: "string"},
                "notifications": {Type: "bool"},
            },
        },
        "scores": {
            Type: "array",
            ElementType: &metadat.FieldType{Type: "int"},
        },
    },
}

writer := metadat.NewWriter()
writer.SetSchema(schema)

data := map[string]interface{}{
    "userId": "U12345",
    "settings": map[string]interface{}{
        "theme": "dark",
        "notifications": true,
    },
    "scores": []interface{}{95, 87, 92, 88},
}

content, err := writer.WriteMetaDat(data)
```

## Array Size Handling

The MetaDat format embeds array sizes directly in the data section. The library automatically reads and validates these sizes:

### Format Example
```
meta
    tags: string[]
    products: {id:int|name:string|price:float64}[]

data
    tags[3]: electronics|computers|premium
    products[1000000]:
        1|Widget A|19.99
        2|Widget B|29.99
        3|Widget C|39.99
        ...
```

### Key Features:
- **Automatic Size Detection**: Array sizes are read from `arrayName[size]:` declarations
- **Size Validation**: The parser validates that the declared size matches the actual number of elements
- **No Memory Limits**: Can handle arrays of any size declared in the format
- **Error Reporting**: Clear error messages for size mismatches

### Usage:
```go
// The parser automatically handles any array size
parser := metadat.NewParser()
data, err := parser.ParseMetaDat(content)
// No need to configure maximum array sizes
```

## Performance

The MetaDat Go library is designed for high performance:

- **Efficient Parsing**: Single-pass parsing with minimal allocations
- **Streaming Support**: Can process large files without loading everything into memory
- **Concurrent Safe**: Parser and Writer instances can be used concurrently

Benchmark results on a typical machine:
```
BenchmarkWriteStruct-8      300000      4521 ns/op     1856 B/op      42 allocs/op
BenchmarkParseMetaDat-8     200000      7832 ns/op     2144 B/op      58 allocs/op
```

## Testing

Run the test suite:

```bash
go test -v
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details.