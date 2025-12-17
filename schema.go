package metadat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Schema represents the metadata structure
type Schema struct {
	Fields     map[string]FieldType
	FieldOrder []string // preserve original field order
}

// FieldType represents a field's type information
type FieldType struct {
	Type         string                 // basic type: string, int, float32, float64, bool, array, object
	ElementType  *FieldType             // for arrays
	ObjectFields map[string]FieldType   // for objects
	ObjectOrder  []string              // preserve object field order
	Name         string                 // field name (used in arrays/objects)
}

// parseSchema parses the meta section into a Schema
func parseSchema(metaContent string) (Schema, error) {
	schema := Schema{
		Fields:     make(map[string]FieldType),
		FieldOrder: make([]string, 0),
	}
	lines := strings.Split(strings.TrimSpace(metaContent), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue
		}

		fieldName := strings.TrimSpace(line[:colonIndex])
		typeStr := strings.TrimSpace(line[colonIndex+1:])

		fieldType, err := parseType(typeStr)
		if err != nil {
			return schema, fmt.Errorf("error parsing type for field %s: %v", fieldName, err)
		}

		schema.Fields[fieldName] = fieldType
		schema.FieldOrder = append(schema.FieldOrder, fieldName)
	}

	return schema, nil
}

// parseType parses a type string into a FieldType
func parseType(typeStr string) (FieldType, error) {
	typeStr = strings.TrimSpace(typeStr)

	// Check for array type
	if strings.HasSuffix(typeStr, "[]") {
		elementTypeStr := strings.TrimSuffix(typeStr, "[]")
		elementType, err := parseType(elementTypeStr)
		if err != nil {
			return FieldType{}, err
		}
		return FieldType{
			Type:        "array",
			ElementType: &elementType,
		}, nil
	}

	// Check for object type
	if strings.HasPrefix(typeStr, "{") && strings.HasSuffix(typeStr, "}") {
		objectStr := typeStr[1 : len(typeStr)-1]
		fields := make(map[string]FieldType)
		fieldOrder := make([]string, 0)
		
		// Parse object fields
		fieldPairs := splitObjectFields(objectStr)
		for _, pair := range fieldPairs {
			colonIndex := strings.Index(pair, ":")
			if colonIndex == -1 {
				return FieldType{}, fmt.Errorf("invalid object field format: %s", pair)
			}
			
			fieldName := strings.TrimSpace(pair[:colonIndex])
			fieldTypeStr := strings.TrimSpace(pair[colonIndex+1:])
			
			fieldType, err := parseType(fieldTypeStr)
			if err != nil {
				return FieldType{}, err
			}
			fieldType.Name = fieldName
			fields[fieldName] = fieldType
			fieldOrder = append(fieldOrder, fieldName)
		}
		
		return FieldType{
			Type:         "object",
			ObjectFields: fields,
			ObjectOrder:  fieldOrder,
		}, nil
	}

	// Basic type
	switch typeStr {
	case "string", "int", "int32", "int64", "float32", "float64", "bool":
		return FieldType{Type: typeStr}, nil
	default:
		return FieldType{}, fmt.Errorf("unknown type: %s", typeStr)
	}
}

// splitObjectFields splits object field definitions considering nested structures
func splitObjectFields(objectStr string) []string {
	var fields []string
	var current strings.Builder
	depth := 0
	
	for _, ch := range objectStr {
		switch ch {
		case '{':
			depth++
			current.WriteRune(ch)
		case '}':
			depth--
			current.WriteRune(ch)
		case '|':
			if depth == 0 {
				fields = append(fields, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}
	
	if current.Len() > 0 {
		fields = append(fields, current.String())
	}
	
	return fields
}

// preserveFieldOrder stores the original field order from schema
type FieldTypeWithOrder struct {
	FieldType
	Order int
}

// ToString converts schema to string representation
func (s Schema) ToString() string {
	var buffer bytes.Buffer
	
	// Get ordered field names
	fieldNames := s.GetFieldOrder()
	
	for _, name := range fieldNames {
		fieldType := s.Fields[name]
		buffer.WriteString(fmt.Sprintf("    %s: %s\n", name, fieldTypeToString(fieldType)))
	}
	
	return buffer.String()
}

// GetFieldOrder returns field names in their original schema order
func (s Schema) GetFieldOrder() []string {
	if len(s.FieldOrder) > 0 {
		return s.FieldOrder
	}
	
	// Fallback to alphabetical order if no order preserved
	names := make([]string, 0, len(s.Fields))
	for name := range s.Fields {
		names = append(names, name)
	}
	
	// Simple alphabetical ordering for consistency
	for i := 0; i < len(names)-1; i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	
	return names
}

// fieldTypeToString converts a FieldType to its string representation
func fieldTypeToString(ft FieldType) string {
	switch ft.Type {
	case "array":
		if ft.ElementType != nil {
			return fieldTypeToString(*ft.ElementType) + "[]"
		}
		return "[]"
		
	case "object":
		var fields []string
		// Use preserved order if available
		fieldNames := ft.ObjectOrder
		if len(fieldNames) == 0 {
			// Fallback to map iteration
			fieldNames = make([]string, 0, len(ft.ObjectFields))
			for name := range ft.ObjectFields {
				fieldNames = append(fieldNames, name)
			}
			// Sort for consistency
			for i := 0; i < len(fieldNames)-1; i++ {
				for j := i + 1; j < len(fieldNames); j++ {
					if fieldNames[i] > fieldNames[j] {
						fieldNames[i], fieldNames[j] = fieldNames[j], fieldNames[i]
					}
				}
			}
		}
		
		for _, name := range fieldNames {
			fieldType := ft.ObjectFields[name]
			fields = append(fields, fmt.Sprintf("%s:%s", name, fieldTypeToString(fieldType)))
		}
		return "{" + strings.Join(fields, "|") + "}"
		
	default:
		return ft.Type
	}
}

// InferSchemaFromJSON infers a Schema from JSON data
func InferSchemaFromJSON(data interface{}) Schema {
	schema := Schema{
		Fields:     make(map[string]FieldType),
		FieldOrder: make([]string, 0),
	}
	
	if obj, ok := data.(map[string]interface{}); ok {
		for key, value := range obj {
			schema.Fields[key] = inferFieldType(value)
			schema.FieldOrder = append(schema.FieldOrder, key)
		}
	}
	
	return schema
}

// inferFieldType infers the FieldType from a value
func inferFieldType(value interface{}) FieldType {
	if value == nil {
		return FieldType{Type: "string"}
	}
	
	switch v := value.(type) {
	case string:
		return FieldType{Type: "string"}
		
	case float64:
		// JSON numbers are always float64
		if v == float64(int(v)) {
			return FieldType{Type: "int"}
		}
		return FieldType{Type: "float64"}
		
	case int, int32, int64:
		return FieldType{Type: "int"}
		
	case float32:
		return FieldType{Type: "float32"}
		
	case bool:
		return FieldType{Type: "bool"}
		
	case []interface{}:
		// Infer array element type from first element
		var elementType *FieldType
		if len(v) > 0 {
			inferred := inferFieldType(v[0])
			elementType = &inferred
		}
		return FieldType{
			Type:        "array",
			ElementType: elementType,
		}
		
	case map[string]interface{}:
		fields := make(map[string]FieldType)
		objectOrder := make([]string, 0, len(v))
		for key, val := range v {
			fields[key] = inferFieldType(val)
			objectOrder = append(objectOrder, key)
		}
		return FieldType{
			Type:         "object",
			ObjectFields: fields,
			ObjectOrder:  objectOrder,
		}
		
	default:
		// Check if it's a slice of a specific type
		valType := reflect.TypeOf(value)
		if valType.Kind() == reflect.Slice {
			// Generic slice handling
			return FieldType{
				Type: "array",
				ElementType: &FieldType{Type: "string"}, // Default to string
			}
		}
		return FieldType{Type: "string"}
	}
}

// InferSchemaFromStruct infers a Schema from a Go struct
func InferSchemaFromStruct(v interface{}) (Schema, error) {
	schema := Schema{Fields: make(map[string]FieldType)}
	
	// Convert to JSON and back to handle nested structs
	data, err := json.Marshal(v)
	if err != nil {
		return schema, fmt.Errorf("failed to marshal struct: %v", err)
	}
	
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return schema, fmt.Errorf("failed to unmarshal to map: %v", err)
	}
	
	return InferSchemaFromJSON(jsonData), nil
}

// ValidateData validates data against the schema
func (s Schema) ValidateData(data map[string]interface{}) error {
	// Check for required fields
	for fieldName, fieldType := range s.Fields {
		value, exists := data[fieldName]
		if !exists {
			continue // Field is optional
		}
		
		if err := validateValue(value, fieldType); err != nil {
			return fmt.Errorf("validation error for field %s: %v", fieldName, err)
		}
	}
	
	// Check for unknown fields
	for fieldName := range data {
		if _, exists := s.Fields[fieldName]; !exists {
			return fmt.Errorf("unknown field: %s", fieldName)
		}
	}
	
	return nil
}

// validateValue validates a value against its expected type
func validateValue(value interface{}, fieldType FieldType) error {
	switch fieldType.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		
	case "int", "int32", "int64":
		switch value.(type) {
		case int, int32, int64, float64:
			// Accept numeric types
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
		
	case "float32", "float64":
		switch value.(type) {
		case float32, float64, int:
			// Accept numeric types
		default:
			return fmt.Errorf("expected float, got %T", value)
		}
		
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
		
	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
		
		// Validate each element if element type is defined
		if fieldType.ElementType != nil {
			for i, elem := range arr {
				if err := validateValue(elem, *fieldType.ElementType); err != nil {
					return fmt.Errorf("array element %d: %v", i, err)
				}
			}
		}
		
	case "object":
		obj, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
		
		// Validate object fields
		for fieldName, fieldDef := range fieldType.ObjectFields {
			if val, exists := obj[fieldName]; exists {
				if err := validateValue(val, fieldDef); err != nil {
					return fmt.Errorf("object field %s: %v", fieldName, err)
				}
			}
		}
		
	default:
		return fmt.Errorf("unknown type: %s", fieldType.Type)
	}
	
	return nil
}