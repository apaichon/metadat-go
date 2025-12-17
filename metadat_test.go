package metadat

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structures for writing
type User struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	InStock  bool    `json:"inStock"`
	Tags     []string `json:"tags"`
}

type Company struct {
	Name       string     `json:"name"`
	Founded    int        `json:"founded"`
	Employees  []Employee `json:"employees"`
}

type Employee struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Salary float64 `json:"salary"`
}

func TestParseSimpleMetaDat(t *testing.T) {
	content := `meta
    name: string
    age: int
    email: string
    active: bool
data
    name:
        John Doe
    age:
        30
    email:
        john.doe@example.com
    active:
        true`

	parser := NewParser()
	result, err := parser.ParseMetaDat(content)

	require.NoError(t, err)
	assert.Equal(t, "John Doe", result["name"])
	assert.Equal(t, 30, result["age"])
	assert.Equal(t, "john.doe@example.com", result["email"])
	assert.Equal(t, true, result["active"])
}

func TestWriteSimpleStruct(t *testing.T) {
	user := User{
		Name:   "Alice Johnson",
		Age:    28,
		Email:  "alice@example.com",
		Active: true,
	}

	writer := NewWriter()
	content, err := writer.WriteStruct(user)
	
	require.NoError(t, err)
	assert.Contains(t, content, "meta")
	assert.Contains(t, content, "name: string")
	assert.Contains(t, content, "age: int")
	assert.Contains(t, content, "email: string")
	assert.Contains(t, content, "active: bool")
	assert.Contains(t, content, "data")
	assert.Contains(t, content, "Alice Johnson")
	assert.Contains(t, content, "28")
	assert.Contains(t, content, "alice@example.com")
	assert.Contains(t, content, "true")
}

func TestWriteStructWithArray(t *testing.T) {
	product := Product{
		ID:      1,
		Name:    "Laptop Pro",
		Price:   1299.99,
		InStock: true,
		Tags:    []string{"electronics", "computers", "premium"},
	}

	writer := NewWriter()
	content, err := writer.WriteStruct(product)
	
	require.NoError(t, err)
	assert.Contains(t, content, "tags: string[]")
	assert.Contains(t, content, "tags[3]: electronics|computers|premium")
}

func TestWriteStructWithNestedObjects(t *testing.T) {
	company := Company{
		Name:    "TechCorp",
		Founded: 2010,
		Employees: []Employee{
			{ID: 1, Name: "Alice", Role: "CEO", Salary: 150000},
			{ID: 2, Name: "Bob", Role: "CTO", Salary: 140000},
		},
	}

	writer := NewWriter()
	content, err := writer.WriteStruct(company)
	
	require.NoError(t, err)
	assert.Contains(t, content, "employees: {")
	assert.Contains(t, content, "id:int")
	assert.Contains(t, content, "name:string")
	assert.Contains(t, content, "role:string")
	assert.Contains(t, content, "salary:")
	assert.Contains(t, content, "employees[2]:")
	assert.Contains(t, content, "Alice")
	assert.Contains(t, content, "Bob")
	assert.Contains(t, content, "CEO")
	assert.Contains(t, content, "CTO")
	assert.Contains(t, content, "150000")
	assert.Contains(t, content, "140000")
}

func TestWriteSeparatedFiles(t *testing.T) {
	user := User{
		Name:   "Bob Smith",
		Age:    35,
		Email:  "bob@example.com",
		Active: false,
	}

	writer := NewWriter()
	schema, data, err := writer.WriteSeparated(user)
	
	require.NoError(t, err)
	
	// Check schema
	assert.Contains(t, schema, "name: string")
	assert.Contains(t, schema, "age: int")
	assert.Contains(t, schema, "email: string")
	assert.Contains(t, schema, "active: bool")
	assert.NotContains(t, schema, "Bob Smith")
	
	// Check data
	assert.Contains(t, data, "Bob Smith")
	assert.Contains(t, data, "35")
	assert.Contains(t, data, "bob@example.com")
	assert.Contains(t, data, "false")
	assert.NotContains(t, data, "meta")
}

func TestRoundTripConversion(t *testing.T) {
	// Original struct
	original := Company{
		Name:    "StartupInc",
		Founded: 2020,
		Employees: []Employee{
			{ID: 1, Name: "Charlie", Role: "Developer", Salary: 80000},
			{ID: 2, Name: "Dana", Role: "Designer", Salary: 75000},
			{ID: 3, Name: "Eve", Role: "Manager", Salary: 90000},
		},
	}

	// Write to MetaDat
	writer := NewWriter()
	metadatContent, err := writer.WriteStruct(original)
	require.NoError(t, err)

	// Parse back
	parser := NewParser()
	parsed, err := parser.ParseMetaDat(metadatContent)
	require.NoError(t, err)

	// Verify data
	assert.Equal(t, "StartupInc", parsed["name"])
	assert.Equal(t, 2020, parsed["founded"])
	
	employees := parsed["employees"].([]interface{})
	assert.Len(t, employees, 3)
	
	emp1 := employees[0].(map[string]interface{})
	assert.Equal(t, 1, emp1["id"])
	assert.Equal(t, "Charlie", emp1["name"])
	assert.Equal(t, "Developer", emp1["role"])
}

func TestWriteToFiles(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "metadat-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	schemaFile := tmpDir + "/schema.metadat"
	dataFile := tmpDir + "/data.metadat"

	user := User{
		Name:   "Test User",
		Age:    25,
		Email:  "test@example.com",
		Active: true,
	}

	writer := NewWriter()
	err = writer.WriteStructToFiles(user, schemaFile, dataFile)
	require.NoError(t, err)

	// Verify files exist
	_, err = os.Stat(schemaFile)
	assert.NoError(t, err)
	_, err = os.Stat(dataFile)
	assert.NoError(t, err)

	// Parse from files
	parser := NewParser()
	result, err := parser.ParseFromFiles(schemaFile, dataFile)
	require.NoError(t, err)

	assert.Equal(t, "Test User", result["name"])
	assert.Equal(t, 25, result["age"])
	assert.Equal(t, "test@example.com", result["email"])
	assert.Equal(t, true, result["active"])
}

func TestConvertJSONToMetaDat(t *testing.T) {
	jsonStr := `{
		"product": {
			"id": 123,
			"name": "Widget",
			"price": 29.99,
			"available": true
		},
		"quantity": 50,
		"tags": ["new", "featured", "sale"]
	}`

	metadat, err := ConvertJSONToMetaDat(jsonStr)
	require.NoError(t, err)

	assert.Contains(t, metadat, "meta")
	assert.Contains(t, metadat, "product: {")
	assert.Contains(t, metadat, "id:int")
	assert.Contains(t, metadat, "name:string")
	assert.Contains(t, metadat, "price:float64")
	assert.Contains(t, metadat, "available:bool")
	assert.Contains(t, metadat, "quantity: int")
	assert.Contains(t, metadat, "tags: string[]")
	assert.Contains(t, metadat, "data")
	assert.Contains(t, metadat, "Widget")
	assert.Contains(t, metadat, "29.99")
	assert.Contains(t, metadat, "tags[3]: new|featured|sale")
}

func TestParseArrayFormats(t *testing.T) {
	// Test inline array format
	content := `meta
    tags: string[]
data
    tags[3]: one|two|three`

	parser := NewParser()
	result, err := parser.ParseMetaDat(content)
	require.NoError(t, err)

	tags := result["tags"].([]interface{})
	assert.Len(t, tags, 3)
	assert.Equal(t, "one", tags[0])
	assert.Equal(t, "two", tags[1])
	assert.Equal(t, "three", tags[2])
}

func TestWriteComplexData(t *testing.T) {
	data := map[string]interface{}{
		"settings": map[string]interface{}{
			"notifications": true,
			"theme":         "dark",
		},
		"scores": []interface{}{95, 87, 92},
	}

	// Infer schema
	schema := InferSchemaFromJSON(data)
	
	writer := NewWriter()
	writer.SetSchema(schema)
	
	content, err := writer.WriteMetaDat(data)
	require.NoError(t, err)
	
	// Parse back and verify
	parser := NewParser()
	parsed, err := parser.ParseMetaDat(content)
	require.NoError(t, err)
	
	// Convert to JSON for easy comparison
	originalJSON, _ := json.Marshal(data)
	parsedJSON, _ := json.Marshal(parsed)
	
	var originalMap, parsedMap map[string]interface{}
	json.Unmarshal(originalJSON, &originalMap)
	json.Unmarshal(parsedJSON, &parsedMap)
	
	// Basic comparison
	assert.NotNil(t, parsedMap["settings"])
	assert.NotNil(t, parsedMap["scores"])
}

func TestSchemaValidation(t *testing.T) {
	schema := Schema{
		Fields: map[string]FieldType{
			"name":   {Type: "string"},
			"age":    {Type: "int"},
			"active": {Type: "bool"},
		},
	}

	// Valid data
	validData := map[string]interface{}{
		"name":   "Alice",
		"age":    30,
		"active": true,
	}
	err := schema.ValidateData(validData)
	assert.NoError(t, err)

	// Invalid data - wrong type
	invalidData := map[string]interface{}{
		"name":   "Bob",
		"age":    "thirty", // Should be int
		"active": true,
	}
	err = schema.ValidateData(invalidData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected integer")

	// Invalid data - unknown field
	unknownFieldData := map[string]interface{}{
		"name":    "Charlie",
		"age":     25,
		"active":  false,
		"unknown": "field",
	}
	err = schema.ValidateData(unknownFieldData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field")
}

// Benchmark tests
func BenchmarkWriteStruct(b *testing.B) {
	user := User{
		Name:   "Benchmark User",
		Age:    30,
		Email:  "bench@example.com",
		Active: true,
	}

	writer := NewWriter()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = writer.WriteStruct(user)
	}
}

func BenchmarkParseMetaDat(b *testing.B) {
	content := `meta
    name: string
    age: int
    email: string
    active: bool
data
    name:
        Benchmark User
    age:
        30
    email:
        bench@example.com
    active:
        true`

	parser := NewParser()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseMetaDat(content)
	}
}