package metadat

import (
	"fmt"
	"strconv"
	"strings"
)

// parseValue parses a value according to its type from the data section
func parseValue(fieldType FieldType, valueStr string, lines []string, currentIndex int) (interface{}, int, error) {
	switch fieldType.Type {
	case "string":
		// Multi-line string values continue on next lines with indentation
		if valueStr == "" && currentIndex+1 < len(lines) {
			// Check next line for indented content
			nextLine := lines[currentIndex+1]
			if strings.HasPrefix(nextLine, "    ") || strings.HasPrefix(nextLine, "\t") {
				return strings.TrimSpace(nextLine), currentIndex + 2, nil
			}
		}
		return valueStr, currentIndex + 1, nil

	case "int", "int32", "int64":
		// Check for multi-line value
		if valueStr == "" && currentIndex+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[currentIndex+1])
			if nextLine != "" && !strings.Contains(nextLine, ":") {
				valueStr = nextLine
				currentIndex++
			}
		}
		
		val, err := strconv.ParseInt(strings.TrimSpace(valueStr), 10, 64)
		if err != nil {
			return nil, currentIndex, fmt.Errorf("invalid integer value: %s", valueStr)
		}
		return int(val), currentIndex + 1, nil

	case "float32":
		// Check for multi-line value
		if valueStr == "" && currentIndex+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[currentIndex+1])
			if nextLine != "" && !strings.Contains(nextLine, ":") {
				valueStr = nextLine
				currentIndex++
			}
		}
		
		val, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 32)
		if err != nil {
			return nil, currentIndex, fmt.Errorf("invalid float32 value: %s", valueStr)
		}
		return float32(val), currentIndex + 1, nil

	case "float64":
		// Check for multi-line value
		if valueStr == "" && currentIndex+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[currentIndex+1])
			if nextLine != "" && !strings.Contains(nextLine, ":") {
				valueStr = nextLine
				currentIndex++
			}
		}
		
		val, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
		if err != nil {
			return nil, currentIndex, fmt.Errorf("invalid float64 value: %s", valueStr)
		}
		return val, currentIndex + 1, nil

	case "bool":
		// Check for multi-line value
		if valueStr == "" && currentIndex+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[currentIndex+1])
			if nextLine != "" && !strings.Contains(nextLine, ":") {
				valueStr = nextLine
				currentIndex++
			}
		}
		
		val, err := strconv.ParseBool(strings.TrimSpace(valueStr))
		if err != nil {
			return nil, currentIndex, fmt.Errorf("invalid boolean value: %s", valueStr)
		}
		return val, currentIndex + 1, nil

	case "array":
		return parseArray(fieldType, valueStr, lines, currentIndex)

	case "object":
		return parseObject(fieldType, valueStr, lines, currentIndex)

	default:
		return nil, currentIndex, fmt.Errorf("unknown type: %s", fieldType.Type)
	}
}

// parseArray parses an array value
func parseArray(fieldType FieldType, valueStr string, lines []string, currentIndex int) ([]interface{}, int, error) {
	// Check if values are on the same line (pipe-separated)
	if valueStr != "" && strings.Contains(valueStr, "|") {
		values := strings.Split(valueStr, "|")
		result := make([]interface{}, len(values))
		for i, v := range values {
			result[i] = strings.TrimSpace(v)
		}
		return result, currentIndex + 1, nil
	}

	// For multi-line arrays, use a reasonable default size
	// This function is used for legacy parsing; new code should use parseArrayWithSize
	var arraySize = 50000 // Default size for compatibility

	// Parse multi-line array elements
	result := make([]interface{}, 0, arraySize)
	i := currentIndex + 1

	for i < len(lines) {
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

		// Check if we've reached the default array size limit
		if len(result) >= arraySize {
			break
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

	return result, i, nil
}

// parseObject parses an object value
func parseObject(fieldType FieldType, valueStr string, lines []string, currentIndex int) (map[string]interface{}, int, error) {
	result := make(map[string]interface{})
	
	// Check if object is on same line (pipe-separated)
	if valueStr != "" && strings.Contains(valueStr, "|") {
		obj, _, err := parseObjectFromLine(valueStr, &fieldType)
		return obj, currentIndex + 1, err
	}
	
	// Multi-line object parsing
	i := currentIndex + 1
	
	// Check if next line contains pipe-separated values
	if i < len(lines) {
		nextLine := strings.TrimSpace(lines[i])
		if nextLine != "" && strings.Contains(nextLine, "|") {
			obj, _, err := parseObjectFromLine(nextLine, &fieldType)
			return obj, i + 1, err
		}
	}
	
	// Parse field-by-field
	fieldOrder := getFieldOrder(fieldType.ObjectFields)
	for _, fieldName := range fieldOrder {
		if i >= len(lines) {
			break
		}
		
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}
		
		// Parse "fieldName: value" format
		if strings.HasPrefix(line, fieldName+":") {
			valueStr := strings.TrimSpace(strings.TrimPrefix(line, fieldName+":"))
			fieldDef := fieldType.ObjectFields[fieldName]
			
			value, newIndex, err := parseValue(fieldDef, valueStr, lines, i)
			if err != nil {
				return nil, i, err
			}
			
			result[fieldName] = value
			i = newIndex
		}
	}
	
	return result, i, nil
}

// parseObjectFromLine parses an object from a pipe-separated line
func parseObjectFromLine(line string, fieldType *FieldType) (map[string]interface{}, int, error) {
	values := strings.Split(line, "|")
	result := make(map[string]interface{})
	
	fieldOrder := getObjectFieldOrder(fieldType)
	
	for i, fieldName := range fieldOrder {
		if i >= len(values) {
			break
		}
		
		fieldDef := fieldType.ObjectFields[fieldName]
		valueStr := strings.TrimSpace(values[i])
		
		// Convert value based on field type
		switch fieldDef.Type {
		case "int", "int32", "int64":
			val, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid integer for field %s: %s", fieldName, valueStr)
			}
			result[fieldName] = int(val)
			
		case "float32":
			val, err := strconv.ParseFloat(valueStr, 32)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid float32 for field %s: %s", fieldName, valueStr)
			}
			result[fieldName] = float32(val)
			
		case "float64":
			val, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid float64 for field %s: %s", fieldName, valueStr)
			}
			result[fieldName] = val
			
		case "bool":
			val, err := strconv.ParseBool(valueStr)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid boolean for field %s: %s", fieldName, valueStr)
			}
			result[fieldName] = val
			
		default:
			result[fieldName] = valueStr
		}
	}
	
	return result, 0, nil
}

// getFieldOrder returns field names in their original order
func getFieldOrder(fields map[string]FieldType) []string {
	// Try to get the order from ObjectOrder if available
	for _, field := range fields {
		if field.Type == "object" && len(field.ObjectOrder) > 0 {
			return field.ObjectOrder
		}
	}
	
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	
	// Sort alphabetically for consistency
	for i := 0; i < len(names)-1; i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	
	return names
}

// getObjectFieldOrder returns field names in their original object definition order
func getObjectFieldOrder(fieldType *FieldType) []string {
	if fieldType != nil && len(fieldType.ObjectOrder) > 0 {
		return fieldType.ObjectOrder
	}
	
	if fieldType != nil {
		names := make([]string, 0, len(fieldType.ObjectFields))
		for name := range fieldType.ObjectFields {
			names = append(names, name)
		}
		
		// Sort alphabetically for consistency
		for i := 0; i < len(names)-1; i++ {
			for j := i + 1; j < len(names); j++ {
				if names[i] > names[j] {
					names[i], names[j] = names[j], names[i]
				}
			}
		}
		
		return names
	}
	
	return []string{}
}