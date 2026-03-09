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
- **v1.1 Extended Types**: `decimal` (arbitrary-precision), `binary` (base64-encoded), `byte` (0-255)

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

### v1.1 Types: decimal, byte, binary

```go
schema := metadat.Schema{
    Fields: map[string]metadat.FieldType{
        "product":  {Type: "string"},
        "price":    {Type: "decimal"},   // arbitrary-precision decimal
        "avatar":   {Type: "binary"},    // base64-encoded binary data
        "priority": {Type: "byte"},      // unsigned integer 0-255
        "payload":  {Type: "array", ElementType: &metadat.FieldType{Type: "byte"}},
    },
    FieldOrder: []string{"product", "price", "avatar", "priority", "payload"},
}

data := map[string]interface{}{
    "product":  "Gold Bar",
    "price":    "2345.6789012345",                  // full precision preserved
    "avatar":   "iVBORw0KGgoAAAANSUhEUgAAAAE=",     // base64 string
    "priority": 128,                                 // byte value
    "payload":  []interface{}{0, 127, 255},          // byte array
}

writer := metadat.NewWriter()
writer.SetSchema(schema)
content, err := writer.WriteMetaDat(data)
```

Output:
```
meta
    product: string
    price: decimal
    avatar: binary
    priority: byte
    payload: byte[]

data
    product:
        Gold Bar
    price:
        2345.6789012345
    avatar:
        iVBORw0KGgoAAAANSUhEUgAAAAE=
    priority:
        128
    payload[3]: 0|127|255
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

## Type System

### Basic Types (v1.0)
| Type | Go Representation | Description | Example |
|------|-------------------|-------------|---------|
| `string` | `string` | UTF-8 text | `Hello World` |
| `int` | `int` | 32-bit signed integer | `42` |
| `int32` | `int` | 32-bit signed integer | `42` |
| `int64` | `int` | 64-bit signed integer | `9223372036854775807` |
| `float32` | `float32` | 32-bit floating point | `3.14` |
| `float64` | `float64` | 64-bit floating point | `3.141592653589793` |
| `bool` | `bool` | Boolean value | `true` / `false` |

### Extended Types (v1.1)
| Type | Go Representation | Description | Example |
|------|-------------------|-------------|---------|
| `decimal` | `string` | Arbitrary-precision decimal (no float rounding) | `12345.6789012345` |
| `binary` | `string` | Base64-encoded binary data | `SGVsbG8gV29ybGQ=` |
| `byte` | `int` | Single unsigned byte (0-255) | `255` |

### Complex Types
| Type | Syntax | Example |
|------|--------|---------|
| Array | `type[]` | `string[]`, `byte[]`, `decimal[]` |
| Object | `{field1:type1\|field2:type2}` | `{name:string\|price:decimal}` |
| Array of Objects | `{field:type}[]` | `{name:string\|weight:byte}[]` |

### v1.1 Type Details

**`decimal`** - Use for financial data, scientific measurements, or any value where floating-point rounding is unacceptable. Values are stored and parsed as strings to preserve exact precision.

**`binary`** - Use for embedding files, images, cryptographic keys, or any raw binary content. Data is encoded as standard Base64 (RFC 4648). The parser validates the Base64 encoding on read.

**`byte`** - Use for status codes, flags, protocol fields, or pixel data. Accepts unsigned integers from 0 to 255. `byte[]` arrays use pipe-separated inline format for compact storage of raw byte sequences.

### Storing Large Files with byte[]

For binary files that need individual byte access (e.g., firmware, pixel data), use `byte[]`:

```
meta
    filename: string
    size: int
    content: byte[]

data
    filename:
        firmware_v2.bin
    size:
        256
    content[256]: 0|1|2|3|4|5|...|253|254|255
```

For opaque binary blobs (e.g., images, documents), prefer `binary` which stores the entire file as a single Base64 string:

```
meta
    filename: string
    content: binary

data
    filename:
        photo.png
    content:
        iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAY...
```

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

### Custom Schema with v1.1 Types

```go
schema := metadat.Schema{
    Fields: map[string]metadat.FieldType{
        "userId":  {Type: "string"},
        "balance": {Type: "decimal"},
        "avatar":  {Type: "binary"},
        "level":   {Type: "byte"},
        "items": {
            Type: "array",
            ElementType: &metadat.FieldType{
                Type: "object",
                ObjectFields: map[string]metadat.FieldType{
                    "name":   {Type: "string"},
                    "price":  {Type: "decimal"},
                    "weight": {Type: "byte"},
                },
                ObjectOrder: []string{"name", "price", "weight"},
            },
        },
    },
    FieldOrder: []string{"userId", "balance", "avatar", "level", "items"},
}

writer := metadat.NewWriter()
writer.SetSchema(schema)

data := map[string]interface{}{
    "userId":  "U12345",
    "balance": "50000.123456789",
    "avatar":  "SGVsbG8gV29ybGQ=",
    "level":   42,
    "items": []interface{}{
        map[string]interface{}{"name": "Sword", "price": "199.99", "weight": 150},
        map[string]interface{}{"name": "Shield", "price": "149.50", "weight": 200},
    },
}

content, err := writer.WriteMetaDat(data)
```

Output:
```
meta
    userId: string
    balance: decimal
    avatar: binary
    level: byte
    items: {name:string|price:decimal|weight:byte}[]

data
    userId:
        U12345
    balance:
        50000.123456789
    avatar:
        SGVsbG8gV29ybGQ=
    level:
        42
    items[2]:
        Sword|199.99|150
        Shield|149.50|200
```

## Array Size Handling

The MetaDat format embeds array sizes directly in the data section. The library automatically reads and validates these sizes:

### Format Example
```
meta
    tags: string[]
    prices: decimal[]
    flags: byte[]
    products: {id:int|name:string|price:decimal|weight:byte}[]

data
    tags[3]: electronics|computers|premium
    prices[3]: 19.99|29.999|39.12345
    flags[4]: 0|128|200|255
    products[3]:
        1|Widget A|19.99|50
        2|Widget B|29.99|100
        3|Widget C|39.99|200
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

Run the test suite (40 tests including v1.1 type coverage):

```bash
go test -v
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

### Test Coverage

- Basic types: string, int, int32, int64, float32, float64, bool
- v1.1 types: decimal, byte, binary (parsing, writing, round-trip, validation, edge cases)
- Arrays: inline and multi-line formats, byte[], decimal[]
- Objects: with all type combinations including v1.1 types
- Array of objects: with decimal and byte fields
- Schema validation: type checking for all types
- File I/O: single and separated file modes
- JSON conversion: round-trip integrity

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details.