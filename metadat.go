// Package metadat provides functionality for parsing and writing MetaDat format files.
// MetaDat is a schema-first data serialization format that separates metadata from data
// for enhanced type safety and compression.
package metadat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Parser handles parsing of MetaDat format files
type Parser struct {
	schema Schema
}

// Writer handles writing data to MetaDat format
type Writer struct {
	schema Schema
}

// NewParser creates a new MetaDat parser
func NewParser() *Parser {
	return &Parser{
		schema: Schema{Fields: make(map[string]FieldType)},
	}
}

// NewWriter creates a new MetaDat writer
func NewWriter() *Writer {
	return &Writer{
		schema: Schema{Fields: make(map[string]FieldType)},
	}
}

// ParseMetaDat parses a complete MetaDat format string with both meta and data sections
func (p *Parser) ParseMetaDat(content string) (map[string]interface{}, error) {
	sections := strings.Split(content, "\ndata\n")
	if len(sections) != 2 {
		return nil, fmt.Errorf("invalid MetaDat format: must have 'meta' and 'data' sections")
	}

	metaSection := strings.TrimPrefix(sections[0], "meta\n")
	dataSection := sections[1]

	// Parse schema
	schema, err := parseSchema(metaSection)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %v", err)
	}
	p.schema = schema

	// Parse data
	return p.ParseData(dataSection)
}

// ParseFromFiles parses MetaDat from separate schema and data files
func (p *Parser) ParseFromFiles(schemaFile, dataFile string) (map[string]interface{}, error) {
	// Read schema file
	schemaContent, err := os.ReadFile(schemaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %v", err)
	}

	// Parse schema
	schema, err := parseSchema(string(schemaContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %v", err)
	}
	p.schema = schema

	// Read data file
	dataContent, err := os.ReadFile(dataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %v", err)
	}

	// Parse data
	return p.ParseData(string(dataContent))
}

// ParseSchema parses only the schema definition
func (p *Parser) ParseSchema(schemaContent string) error {
	schema, err := parseSchema(schemaContent)
	if err != nil {
		return err
	}
	p.schema = schema
	return nil
}

// ParseData parses the data section using the current schema
func (p *Parser) ParseData(dataContent string) (map[string]interface{}, error) {
	if len(p.schema.Fields) == 0 {
		return nil, fmt.Errorf("no schema loaded")
	}

	result := make(map[string]interface{})
	lines := strings.Split(strings.TrimSpace(dataContent), "\n")
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}

		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			return nil, fmt.Errorf("invalid data format at line %d: %s", i+1, line)
		}

		fieldNameWithSize := strings.TrimSpace(line[:colonIndex])
		fieldValue := strings.TrimSpace(line[colonIndex+1:])

		// Handle array notation like "arrayName[3]:"
		fieldName := fieldNameWithSize
		arraySize := 0
		if strings.Contains(fieldNameWithSize, "[") {
			bracketIndex := strings.Index(fieldNameWithSize, "[")
			closeBracketIndex := strings.Index(fieldNameWithSize, "]")
			if closeBracketIndex > bracketIndex {
				fieldName = fieldNameWithSize[:bracketIndex]
				sizeStr := fieldNameWithSize[bracketIndex+1 : closeBracketIndex]
				if size, err := strconv.Atoi(sizeStr); err == nil {
					arraySize = size
				}
			}
		}

		fieldType, exists := p.schema.Fields[fieldName]
		if !exists {
			return nil, fmt.Errorf("unknown field: %s", fieldName)
		}

		value, newIndex, err := p.parseValueWithArraySize(fieldType, fieldValue, lines, i, arraySize)
		if err != nil {
			return nil, fmt.Errorf("error parsing field %s: %v", fieldName, err)
		}

		result[fieldName] = value
		i = newIndex
	}

	return result, nil
}

// parseValueWithArraySize parses a value with the array size specified in the format
func (p *Parser) parseValueWithArraySize(fieldType FieldType, valueStr string, lines []string, currentIndex int, arraySize int) (interface{}, int, error) {
	switch fieldType.Type {
	case "array":
		return p.parseArrayWithDeclaredSize(fieldType, valueStr, lines, currentIndex, arraySize)
	default:
		return parseValue(fieldType, valueStr, lines, currentIndex)
	}
}

// parseArrayWithDeclaredSize parses an array value using the size declared in the format
func (p *Parser) parseArrayWithDeclaredSize(fieldType FieldType, valueStr string, lines []string, currentIndex int, declaredSize int) ([]interface{}, int, error) {
	// Check if values are on the same line (pipe-separated)
	if valueStr != "" && strings.Contains(valueStr, "|") {
		values := strings.Split(valueStr, "|")
		// Validate that the number of values matches the declared size
		if declaredSize > 0 && len(values) != declaredSize {
			return nil, currentIndex, fmt.Errorf("array size mismatch: declared %d, found %d elements", declaredSize, len(values))
		}
		result := make([]interface{}, len(values))
		for i, v := range values {
			result[i] = strings.TrimSpace(v)
		}
		return result, currentIndex + 1, nil
	}

	// For multi-line arrays, parse exactly the declared number of elements
	expectedSize := declaredSize
	if expectedSize <= 0 {
		// If no size declared, parse until end of indented block
		expectedSize = 1000000 // Large fallback for undeclared arrays
	}

	// Parse multi-line array elements
	result := make([]interface{}, 0, expectedSize)
	i := currentIndex + 1

	for i < len(lines) && (declaredSize <= 0 || len(result) < declaredSize) {
		line := lines[i]
		
		// Check if line is indented (part of array)
		if !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t") {
			break
		}

		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			i++
			continue
		}

		// Parse array element based on element type
		if fieldType.ElementType != nil && fieldType.ElementType.Type == "object" {
			// Parse object from pipe-separated values
			obj, _, err := parseObjectFromLine(trimmedLine, fieldType.ElementType)
			if err != nil {
				return nil, i, err
			}
			result = append(result, obj)
		} else {
			// Simple value
			result = append(result, trimmedLine)
		}
		
		i++
	}

	// Validate that we got the expected number of elements
	if declaredSize > 0 && len(result) != declaredSize {
		return nil, i, fmt.Errorf("array size mismatch: declared %d, found %d elements", declaredSize, len(result))
	}

	return result, i, nil
}

// WriteStruct writes a Go struct to MetaDat format
func (w *Writer) WriteStruct(v interface{}) (string, error) {
	// Infer schema from struct
	schema, err := InferSchemaFromStruct(v)
	if err != nil {
		return "", fmt.Errorf("failed to infer schema: %v", err)
	}
	w.schema = schema

	// Convert struct to map
	data, err := structToMap(v)
	if err != nil {
		return "", fmt.Errorf("failed to convert struct to map: %v", err)
	}

	// Write as MetaDat
	return w.WriteMetaDat(data)
}

// WriteMetaDat writes data to MetaDat format (single file)
func (w *Writer) WriteMetaDat(data map[string]interface{}) (string, error) {
	if len(w.schema.Fields) == 0 {
		return "", fmt.Errorf("no schema defined")
	}

	var buffer bytes.Buffer
	
	// Write meta section
	buffer.WriteString("meta\n")
	schemaStr := w.schema.ToString()
	buffer.WriteString(schemaStr)
	
	// Write data section
	buffer.WriteString("\ndata\n")
	dataStr, err := w.writeData(data)
	if err != nil {
		return "", err
	}
	buffer.WriteString(dataStr)

	return buffer.String(), nil
}

// WriteSeparated writes schema and data to separate strings
func (w *Writer) WriteSeparated(v interface{}) (schema string, dataContent string, err error) {
	// Handle both struct and map inputs
	var data map[string]interface{}
	if m, ok := v.(map[string]interface{}); ok {
		data = m
	} else {
		// Convert struct to map
		data, err = structToMap(v)
		if err != nil {
			return "", "", fmt.Errorf("failed to convert input to map: %v", err)
		}
		
		// Infer schema if not already set
		if len(w.schema.Fields) == 0 {
			w.schema, err = InferSchemaFromStruct(v)
			if err != nil {
				return "", "", fmt.Errorf("failed to infer schema: %v", err)
			}
		}
	}

	if len(w.schema.Fields) == 0 {
		return "", "", fmt.Errorf("no schema defined")
	}

	// Get schema string
	schema = w.schema.ToString()

	// Get data string
	dataContent, err = w.writeData(data)
	if err != nil {
		return "", "", err
	}

	return schema, dataContent, nil
}

// WriteToFiles writes schema and data to separate files
func (w *Writer) WriteToFiles(data map[string]interface{}, schemaFile, dataFile string) error {
	schema, dataContent, err := w.WriteSeparated(data)
	if err != nil {
		return err
	}

	// Write schema file
	if err := os.WriteFile(schemaFile, []byte(schema), 0644); err != nil {
		return fmt.Errorf("failed to write schema file: %v", err)
	}

	// Write data file
	if err := os.WriteFile(dataFile, []byte(dataContent), 0644); err != nil {
		return fmt.Errorf("failed to write data file: %v", err)
	}

	return nil
}

// WriteStructToFile writes a struct to a single MetaDat file
func (w *Writer) WriteStructToFile(v interface{}, filename string) error {
	content, err := w.WriteStruct(v)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

// WriteStructToFiles writes a struct to separate schema and data files
func (w *Writer) WriteStructToFiles(v interface{}, schemaFile, dataFile string) error {
	// Infer schema from struct
	schema, err := InferSchemaFromStruct(v)
	if err != nil {
		return fmt.Errorf("failed to infer schema: %v", err)
	}
	w.schema = schema

	// Convert struct to map
	data, err := structToMap(v)
	if err != nil {
		return fmt.Errorf("failed to convert struct to map: %v", err)
	}

	// Write to files
	return w.WriteToFiles(data, schemaFile, dataFile)
}


// SetSchema sets the schema for the writer
func (w *Writer) SetSchema(schema Schema) {
	w.schema = schema
}

// writeData writes the data portion of MetaDat format
func (w *Writer) writeData(data map[string]interface{}) (string, error) {
	var buffer bytes.Buffer

	// Get ordered field names from schema
	fieldOrder := w.schema.GetFieldOrder()

	for _, fieldName := range fieldOrder {
		fieldType, exists := w.schema.Fields[fieldName]
		if !exists {
			continue
		}

		value, exists := data[fieldName]
		if !exists {
			// Skip missing fields
			continue
		}

		fieldStr, err := w.writeField(fieldName, value, fieldType, 0)
		if err != nil {
			return "", fmt.Errorf("error writing field %s: %v", fieldName, err)
		}

		buffer.WriteString(fieldStr)
		if !strings.HasSuffix(fieldStr, "\n") {
			buffer.WriteString("\n")
		}
	}

	return buffer.String(), nil
}

// writeField writes a single field in MetaDat format
func (w *Writer) writeField(name string, value interface{}, fieldType FieldType, indent int) (string, error) {
	indentStr := strings.Repeat("    ", indent)

	switch fieldType.Type {
	case "string":
		return fmt.Sprintf("%s%s:\n%s    %v", indentStr, name, indentStr, value), nil

	case "int", "int32", "int64":
		return fmt.Sprintf("%s%s:\n%s    %v", indentStr, name, indentStr, value), nil

	case "float32", "float64":
		return fmt.Sprintf("%s%s:\n%s    %v", indentStr, name, indentStr, value), nil

	case "bool":
		return fmt.Sprintf("%s%s:\n%s    %v", indentStr, name, indentStr, value), nil

	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			// Try to convert from typed slices
			arr = convertToInterfaceSlice(value)
			if arr == nil {
				return "", fmt.Errorf("expected array for field %s", name)
			}
		}

		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("%s%s[%d]:", indentStr, name, len(arr)))

		if len(arr) == 0 {
			return buffer.String(), nil
		}

		// Check if it's a simple type array
		if fieldType.ElementType != nil && isSimpleType(fieldType.ElementType.Type) {
			// Write as pipe-separated values on same line
			buffer.WriteString(" ")
			values := make([]string, len(arr))
			for i, item := range arr {
				values[i] = fmt.Sprintf("%v", item)
			}
			buffer.WriteString(strings.Join(values, "|"))
		} else {
			// Write as multi-line for complex types
			buffer.WriteString("\n")
			for _, item := range arr {
				itemStr, err := w.writeArrayItem(item, fieldType.ElementType, indent+1)
				if err != nil {
					return "", err
				}
				buffer.WriteString(itemStr)
				buffer.WriteString("\n")
			}
		}

		return strings.TrimRight(buffer.String(), "\n"), nil

	case "object":
		obj, ok := value.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("expected object for field %s", name)
		}

		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("%s%s:\n", indentStr, name))

		// Write object fields in pipe-separated format
		values := make([]string, 0)
		fieldOrder := getObjectFieldOrder(&fieldType)
		for _, fieldName := range fieldOrder {
			if val, exists := obj[fieldName]; exists {
				values = append(values, fmt.Sprintf("%v", val))
			}
		}
		buffer.WriteString(fmt.Sprintf("%s    %s", indentStr, strings.Join(values, "|")))

		return buffer.String(), nil

	default:
		return "", fmt.Errorf("unknown field type: %s", fieldType.Type)
	}
}

// writeArrayItem writes a single array item
func (w *Writer) writeArrayItem(item interface{}, itemType *FieldType, indent int) (string, error) {
	indentStr := strings.Repeat("    ", indent)

	if itemType == nil {
		return fmt.Sprintf("%s%v", indentStr, item), nil
	}

	switch itemType.Type {
	case "object":
		// Write object fields in pipe-separated format
		obj, ok := item.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("expected object in array")
		}

		values := make([]string, 0)
		fieldOrder := getObjectFieldOrder(itemType)
		for _, fieldName := range fieldOrder {
			if val, exists := obj[fieldName]; exists {
				values = append(values, fmt.Sprintf("%v", val))
			}
		}
		return fmt.Sprintf("%s%s", indentStr, strings.Join(values, "|")), nil

	default:
		return fmt.Sprintf("%s%v", indentStr, item), nil
	}
}

// Helper functions

func isSimpleType(t string) bool {
	return t == "string" || t == "int" || t == "int32" || t == "int64" ||
		t == "float32" || t == "float64" || t == "bool"
}

func convertToInterfaceSlice(v interface{}) []interface{} {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Slice {
		return nil
	}

	result := make([]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		result[i] = val.Index(i).Interface()
	}
	return result
}

func structToMap(v interface{}) (map[string]interface{}, error) {
	// Handle both struct and map inputs
	if m, ok := v.(map[string]interface{}); ok {
		return m, nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}


// ConvertJSONToMetaDat converts JSON string to MetaDat format
func ConvertJSONToMetaDat(jsonStr string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %v", err)
	}

	schema := InferSchemaFromJSON(data)
	
	// Convert data to map if it's not already
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("JSON must be an object at root level")
	}

	writer := NewWriter()
	writer.SetSchema(schema)
	return writer.WriteMetaDat(dataMap)
}

// ConvertMetaDatToJSON converts MetaDat format to JSON
func ConvertMetaDatToJSON(metadatContent string) (string, error) {
	parser := NewParser()
	data, err := parser.ParseMetaDat(metadatContent)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}