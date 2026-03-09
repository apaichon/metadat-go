package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/apaichon/metadat-go"
)

// Example structs
type User struct {
	Name     string   `json:"name"`
	Age      int      `json:"age"`
	Email    string   `json:"email"`
	Active   bool     `json:"active"`
	Tags     []string `json:"tags"`
	Settings Settings `json:"settings"`
}

type Settings struct {
	Theme         string  `json:"theme"`
	Notifications bool    `json:"notifications"`
	Version       float64 `json:"version"`
}

type Company struct {
	Name       string     `json:"name"`
	Founded    int        `json:"founded"`
	Revenue    float64    `json:"revenue"`
	Public     bool       `json:"public"`
	Employees  []Employee `json:"employees"`
	HQ         Address    `json:"hq"`
}

type Employee struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Position string  `json:"position"`
	Salary   float64 `json:"salary"`
	Remote   bool    `json:"remote"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
	Zipcode string `json:"zipcode"`
}

func main() {
	fmt.Printf("=== MetaDat Go Library v%s Examples ===\n\n", metadat.Version)

	// Example 1: Simple struct serialization
	fmt.Println("1. Basic Struct Serialization")
	user := User{
		Name:   "Alice Johnson",
		Age:    28,
		Email:  "alice@example.com",
		Active: true,
		Tags:   []string{"developer", "golang", "tech"},
		Settings: Settings{
			Theme:         "dark",
			Notifications: true,
			Version:       2.1,
		},
	}

	writer := metadat.NewWriter()
	userMetaDat, err := writer.WriteStruct(user)
	if err != nil {
		log.Fatalf("Error writing struct: %v", err)
	}

	fmt.Println("Generated MetaDat:")
	fmt.Println(userMetaDat)

	// Parse it back
	parser := metadat.NewParser()
	parsedUser, err := parser.ParseMetaDat(userMetaDat)
	if err != nil {
		log.Fatalf("Error parsing MetaDat: %v", err)
	}

	fmt.Printf("Parsed Name: %s\n", parsedUser["name"])
	fmt.Printf("Parsed Age: %d\n", parsedUser["age"])
	fmt.Printf("Parsed Active: %t\n\n", parsedUser["active"])

	// Example 2: Complex nested structure
	fmt.Println("2. Complex Nested Structure")
	company := Company{
		Name:    "TechCorp Inc",
		Founded: 2015,
		Revenue: 50000000.75,
		Public:  true,
		Employees: []Employee{
			{ID: 1, Name: "John Doe", Position: "CEO", Salary: 200000, Remote: false},
			{ID: 2, Name: "Jane Smith", Position: "CTO", Salary: 180000, Remote: true},
			{ID: 3, Name: "Bob Wilson", Position: "Lead Developer", Salary: 120000, Remote: true},
			{ID: 4, Name: "Carol Davis", Position: "Designer", Salary: 95000, Remote: false},
		},
		HQ: Address{
			Street:  "123 Tech Street",
			City:    "San Francisco",
			Country: "USA",
			Zipcode: "94105",
		},
	}

	companyWriter := metadat.NewWriter()
	companyMetaDat, err := companyWriter.WriteStruct(company)
	if err != nil {
		log.Fatalf("Error writing company struct: %v", err)
	}

	fmt.Println("Company MetaDat (first 500 chars):")
	if len(companyMetaDat) > 500 {
		fmt.Printf("%s...\n\n", companyMetaDat[:500])
	} else {
		fmt.Println(companyMetaDat)
	}

	// Example 3: Separated files mode
	fmt.Println("3. Separated Files Mode")
	separateWriter := metadat.NewWriter()
	schema, data, err := separateWriter.WriteSeparated(user)
	if err != nil {
		log.Fatalf("Error creating separated content: %v", err)
	}

	fmt.Println("Schema content:")
	fmt.Println(schema)
	fmt.Println("\nData content:")
	fmt.Println(data)

	// Example 4: JSON conversion
	fmt.Println("4. JSON to MetaDat Conversion")
	jsonStr := `{
		"product": {
			"id": 1001,
			"name": "Premium Laptop",
			"price": 1299.99,
			"available": true,
			"specs": {
				"cpu": "Intel i7",
				"ram": "16GB",
				"storage": "512GB SSD"
			}
		},
		"quantity": 25,
		"categories": ["electronics", "computers", "laptops"]
	}`

	jsonMetaDat, err := metadat.ConvertJSONToMetaDat(jsonStr)
	if err != nil {
		log.Fatalf("Error converting JSON: %v", err)
	}

	fmt.Println("JSON converted to MetaDat:")
	fmt.Println(jsonMetaDat)

	// Convert back to JSON
	convertedJSON, err := metadat.ConvertMetaDatToJSON(jsonMetaDat)
	if err != nil {
		log.Fatalf("Error converting back to JSON: %v", err)
	}

	fmt.Println("\nConverted back to JSON:")
	fmt.Println(convertedJSON)

	// Example 5: Schema validation
	fmt.Println("\n5. Schema Validation")
	
	// Create a schema manually
	schema_def := metadat.Schema{
		Fields: map[string]metadat.FieldType{
			"username": {Type: "string"},
			"age":      {Type: "int"},
			"email":    {Type: "string"},
			"premium":  {Type: "bool"},
		},
	}

	// Valid data
	validData := map[string]interface{}{
		"username": "testuser",
		"age":      25,
		"email":    "test@example.com",
		"premium":  true,
	}

	err = schema_def.ValidateData(validData)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("✓ Valid data passed validation")
	}

	// Invalid data
	invalidData := map[string]interface{}{
		"username": "testuser2",
		"age":      "twenty-five", // Should be int
		"email":    "test2@example.com",
		"premium":  true,
	}

	err = schema_def.ValidateData(invalidData)
	if err != nil {
		fmt.Printf("✓ Invalid data correctly rejected: %v\n", err)
	} else {
		fmt.Println("✗ Invalid data passed validation (shouldn't happen)")
	}

	// Example 6: File operations
	fmt.Println("\n6. File Operations")
	
	// Write to single file
	singleFileName := "/tmp/user_single.metadat"
	err = writer.WriteStructToFile(user, singleFileName)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
	} else {
		fmt.Printf("✓ Successfully wrote to %s\n", singleFileName)
	}

	// Write to separated files
	schemaFileName := "/tmp/user_schema.metadat"
	dataFileName := "/tmp/user_data.metadat"
	err = writer.WriteStructToFiles(user, schemaFileName, dataFileName)
	if err != nil {
		fmt.Printf("Error writing to separated files: %v\n", err)
	} else {
		fmt.Printf("✓ Successfully wrote to %s and %s\n", schemaFileName, dataFileName)
	}

	// Read from separated files
	fileParser := metadat.NewParser()
	fileData, err := fileParser.ParseFromFiles(schemaFileName, dataFileName)
	if err != nil {
		fmt.Printf("Error reading from files: %v\n", err)
	} else {
		fmt.Printf("✓ Successfully read from files: user %s\n", fileData["name"])
	}

	// Example 7: v1.1 New Types - decimal, byte, binary
	fmt.Println("\n7. v1.1 New Types: decimal, byte, binary")

	// 7a: Decimal type - arbitrary-precision decimal values
	fmt.Println("\n7a. Decimal Type (arbitrary-precision)")
	decimalSchema := metadat.Schema{
		Fields: map[string]metadat.FieldType{
			"product":    {Type: "string"},
			"unitPrice":  {Type: "decimal"},
			"totalPrice": {Type: "decimal"},
			"taxRate":    {Type: "decimal"},
		},
		FieldOrder: []string{"product", "unitPrice", "totalPrice", "taxRate"},
	}

	decimalData := map[string]interface{}{
		"product":    "Gold Bar 1oz",
		"unitPrice":  "2345.6789012345",
		"totalPrice": "23456.789012345",
		"taxRate":    "0.0725",
	}

	decimalWriter := metadat.NewWriter()
	decimalWriter.SetSchema(decimalSchema)
	decimalContent, err := decimalWriter.WriteMetaDat(decimalData)
	if err != nil {
		log.Fatalf("Error writing decimal data: %v", err)
	}
	fmt.Println(decimalContent)

	// Parse back and verify precision is preserved
	decimalParser := metadat.NewParser()
	parsedDecimal, err := decimalParser.ParseMetaDat(decimalContent)
	if err != nil {
		log.Fatalf("Error parsing decimal data: %v", err)
	}
	fmt.Printf("Parsed unitPrice (full precision): %s\n", parsedDecimal["unitPrice"])
	fmt.Printf("Parsed taxRate: %s\n\n", parsedDecimal["taxRate"])

	// 7b: Byte type - single unsigned byte (0-255)
	fmt.Println("7b. Byte Type (0-255)")
	byteSchema := metadat.Schema{
		Fields: map[string]metadat.FieldType{
			"protocol":   {Type: "string"},
			"version":    {Type: "byte"},
			"statusCode": {Type: "byte"},
			"priority":   {Type: "byte"},
		},
		FieldOrder: []string{"protocol", "version", "statusCode", "priority"},
	}

	byteData := map[string]interface{}{
		"protocol":   "HTTP",
		"version":    2,
		"statusCode": 200,
		"priority":   128,
	}

	byteWriter := metadat.NewWriter()
	byteWriter.SetSchema(byteSchema)
	byteContent, err := byteWriter.WriteMetaDat(byteData)
	if err != nil {
		log.Fatalf("Error writing byte data: %v", err)
	}
	fmt.Println(byteContent)

	// 7c: Binary type - base64-encoded binary data
	fmt.Println("7c. Binary Type (base64-encoded)")
	binaryPayload := []byte("Hello, MetaDat v1.1! This is binary data stored as base64.")
	binaryEncoded := base64.StdEncoding.EncodeToString(binaryPayload)

	binarySchema := metadat.Schema{
		Fields: map[string]metadat.FieldType{
			"filename":    {Type: "string"},
			"contentType": {Type: "string"},
			"payload":     {Type: "binary"},
		},
		FieldOrder: []string{"filename", "contentType", "payload"},
	}

	binaryData := map[string]interface{}{
		"filename":    "message.txt",
		"contentType": "text/plain",
		"payload":     binaryEncoded,
	}

	binaryWriter := metadat.NewWriter()
	binaryWriter.SetSchema(binarySchema)
	binaryContent, err := binaryWriter.WriteMetaDat(binaryData)
	if err != nil {
		log.Fatalf("Error writing binary data: %v", err)
	}
	fmt.Println(binaryContent)

	// Decode and verify
	binaryParser := metadat.NewParser()
	parsedBinary, err := binaryParser.ParseMetaDat(binaryContent)
	if err != nil {
		log.Fatalf("Error parsing binary data: %v", err)
	}
	decoded, _ := base64.StdEncoding.DecodeString(parsedBinary["payload"].(string))
	fmt.Printf("Decoded payload: %s\n\n", string(decoded))

	// 7d: byte[] - Storing a large file as byte array
	fmt.Println("7d. Large File with byte[] (simulated 1KB file)")

	// Simulate a large binary file as a byte array (1024 bytes)
	largeFileBytes := make([]interface{}, 256)
	for i := 0; i < 256; i++ {
		largeFileBytes[i] = i // Values 0-255 repeating
	}

	largeFileSchema := metadat.Schema{
		Fields: map[string]metadat.FieldType{
			"filename": {Type: "string"},
			"size":     {Type: "int"},
			"checksum": {Type: "decimal"},
			"content":  {Type: "array", ElementType: &metadat.FieldType{Type: "byte"}},
		},
		FieldOrder: []string{"filename", "size", "checksum", "content"},
	}

	largeFileData := map[string]interface{}{
		"filename": "firmware_v2.bin",
		"size":     256,
		"checksum": "32640.0",
		"content":  largeFileBytes,
	}

	largeFileWriter := metadat.NewWriter()
	largeFileWriter.SetSchema(largeFileSchema)
	largeFileContent, err := largeFileWriter.WriteMetaDat(largeFileData)
	if err != nil {
		log.Fatalf("Error writing large file data: %v", err)
	}

	// Show first 600 chars to demonstrate the format
	fmt.Println("Large byte[] file (first 600 chars):")
	if len(largeFileContent) > 600 {
		fmt.Printf("%s...\n", largeFileContent[:600])
	} else {
		fmt.Println(largeFileContent)
	}
	fmt.Printf("\nTotal MetaDat size: %d bytes\n", len(largeFileContent))

	// Parse back and verify
	largeFileParser := metadat.NewParser()
	parsedLargeFile, err := largeFileParser.ParseMetaDat(largeFileContent)
	if err != nil {
		log.Fatalf("Error parsing large file data: %v", err)
	}
	parsedContent := parsedLargeFile["content"].([]interface{})
	fmt.Printf("Parsed byte count: %d\n", len(parsedContent))
	fmt.Printf("First 10 bytes: %v\n", parsedContent[:10])
	fmt.Printf("Last 5 bytes: %v\n\n", parsedContent[len(parsedContent)-5:])

	// 7e: Combined - all new types in objects and arrays
	fmt.Println("7e. Objects with All New Types")

	combinedContent := `meta
    name: string
    balance: decimal
    avatar: binary
    flags: byte[]
    items: {name:string|price:decimal|weight:byte}[]
data
    name:
        MetaDat Shop
    balance:
        999999.999999999999
    avatar:
        iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==
    flags[4]: 1|2|4|8
    items[3]:
        Widget A|19.99|50
        Widget B|29.999|100
        Widget C|99.12345|200`

	combinedParser := metadat.NewParser()
	parsedCombined, err := combinedParser.ParseMetaDat(combinedContent)
	if err != nil {
		log.Fatalf("Error parsing combined data: %v", err)
	}

	fmt.Printf("Name: %s\n", parsedCombined["name"])
	fmt.Printf("Balance (full precision): %s\n", parsedCombined["balance"])
	fmt.Printf("Avatar (base64, first 40 chars): %.40s...\n", parsedCombined["avatar"])
	fmt.Printf("Flags: %v\n", parsedCombined["flags"])
	items := parsedCombined["items"].([]interface{})
	for i, item := range items {
		obj := item.(map[string]interface{})
		fmt.Printf("  Item %d: %s, price=%s, weight=%v\n", i+1, obj["name"], obj["price"], obj["weight"])
	}

	// Example 8: Size Comparison
	fmt.Println("\n8. Size Comparison")

	// Build a comparison using the combined example
	// Equivalent JSON representation
	equivalentJSON := `{"name":"MetaDat Shop","balance":"999999.999999999999","avatar":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==","flags":[1,2,4,8],"items":[{"name":"Widget A","price":"19.99","weight":50},{"name":"Widget B","price":"29.999","weight":100},{"name":"Widget C","price":"99.12345","weight":200}]}`
	fmt.Printf("JSON size:    %d bytes\n", len(equivalentJSON))
	fmt.Printf("MetaDat size: %d bytes\n", len(strings.TrimSpace(combinedContent)))

	metadatSize := float64(len(strings.TrimSpace(combinedContent)))
	jsonSize := float64(len(equivalentJSON))
	if metadatSize < jsonSize {
		fmt.Printf("Reduction: %.1f%%\n", (jsonSize-metadatSize)/jsonSize*100)
	} else {
		fmt.Printf("Overhead: %.1f%%\n", (metadatSize-jsonSize)/jsonSize*100)
	}

	fmt.Println("\n=== Examples completed successfully! ===")
}