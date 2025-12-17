package main

import (
	"fmt"
	"log"

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

	// Example 7: Performance comparison
	fmt.Println("\n7. Size Comparison")
	
	// Original JSON size
	jsonData, _ := metadat.ConvertMetaDatToJSON(userMetaDat)
	fmt.Printf("JSON size: %d bytes\n", len(jsonData))
	fmt.Printf("MetaDat (single file) size: %d bytes\n", len(userMetaDat))
	fmt.Printf("MetaDat (data only) size: %d bytes\n", len(data))
	
	reduction := float64(len(jsonData)-len(userMetaDat)) / float64(len(jsonData)) * 100
	dataReduction := float64(len(jsonData)-len(data)) / float64(len(jsonData)) * 100
	
	if reduction > 0 {
		fmt.Printf("Single file reduction: %.1f%%\n", reduction)
	} else {
		fmt.Printf("Single file overhead: %.1f%%\n", -reduction)
	}
	fmt.Printf("Data-only reduction: %.1f%%\n", dataReduction)

	fmt.Println("\n=== Examples completed successfully! ===")
}