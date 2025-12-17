package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/apaichon/metadat-go"
)

// Use version from the library package

func main() {
	var (
		inputFile    = flag.String("input", "", "Input file (JSON or MetaDat)")
		outputFile   = flag.String("output", "", "Output file (leave empty for stdout)")
		schemaFile   = flag.String("schema", "", "Schema file for separated mode")
		dataFile     = flag.String("data", "", "Data file for separated mode")
		mode         = flag.String("mode", "auto", "Conversion mode: json-to-metadat, metadat-to-json, parse, validate, or auto")
		separated    = flag.Bool("separated", false, "Use separated files mode for output")
		showVersion  = flag.Bool("version", false, "Show version information")
		showHelp     = flag.Bool("help", false, "Show help information")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("metadat version %s\n", metadat.Version)
		return
	}

	if *showHelp || len(os.Args) == 1 {
		showUsage()
		return
	}

	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: input file is required")
		os.Exit(1)
	}

	// Read input file
	content, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var result string

	switch *mode {
	case "json-to-metadat":
		result, err = convertJSONToMetaDat(string(content), *separated, *schemaFile, *dataFile)
	case "metadat-to-json":
		result, err = convertMetaDatToJSON(string(content), *schemaFile, *dataFile)
	case "parse":
		result, err = parseMetaDat(string(content), *schemaFile, *dataFile)
	case "validate":
		result, err = validateMetaDat(string(content), *schemaFile, *dataFile)
	case "auto":
		result, err = autoConvert(string(content), *separated, *schemaFile, *dataFile)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown mode '%s'\n", *mode)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Output result
	if *outputFile != "" {
		err = os.WriteFile(*outputFile, []byte(result), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Output written to %s\n", *outputFile)
	} else {
		fmt.Print(result)
	}
}

func showUsage() {
	fmt.Printf(`metadat - MetaDat format conversion tool v%s

USAGE:
    metadat [OPTIONS] -input <file>

MODES:
    json-to-metadat    Convert JSON to MetaDat format
    metadat-to-json    Convert MetaDat to JSON format  
    parse             Parse MetaDat and display structure
    validate          Validate MetaDat format
    auto              Auto-detect input format and convert

OPTIONS:
    -input <file>      Input file (required)
    -output <file>     Output file (stdout if not specified)
    -schema <file>     Schema file for separated mode
    -data <file>       Data file for separated mode
    -mode <mode>       Conversion mode (default: auto)
    -separated         Use separated files mode for output
    -version           Show version information
    -help              Show this help message

EXAMPLES:
    # Convert JSON to MetaDat
    metadat -mode json-to-metadat -input data.json -output data.metadat

    # Convert JSON to separated MetaDat files
    metadat -mode json-to-metadat -input data.json -separated -schema schema.metadat -data data.metadat

    # Convert MetaDat to JSON
    metadat -mode metadat-to-json -input data.metadat -output data.json

    # Parse separated MetaDat files
    metadat -mode parse -schema schema.metadat -data data.metadat

    # Auto-detect and convert
    metadat -input data.json -output data.metadat

    # Validate MetaDat file
    metadat -mode validate -input data.metadat
`, metadat.Version)
}

func convertJSONToMetaDat(jsonContent string, separated bool, schemaFile, dataFile string) (string, error) {
	if separated {
		if schemaFile == "" || dataFile == "" {
			return "", fmt.Errorf("schema and data files must be specified for separated mode")
		}

		// Parse JSON
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &data); err != nil {
			return "", fmt.Errorf("invalid JSON: %v", err)
		}

		// Infer schema and write separated files
		schema := metadat.InferSchemaFromJSON(data)
		writer := metadat.NewWriter()
		writer.SetSchema(schema)

		err := writer.WriteToFiles(data, schemaFile, dataFile)
		if err != nil {
			return "", fmt.Errorf("failed to write separated files: %v", err)
		}

		return fmt.Sprintf("Schema written to: %s\nData written to: %s\n", schemaFile, dataFile), nil
	}

	// Convert to single MetaDat file
	return metadat.ConvertJSONToMetaDat(jsonContent)
}

func convertMetaDatToJSON(metadatContent, schemaFile, dataFile string) (string, error) {
	// Create parser
	parser := metadat.NewParser()
	var data map[string]interface{}
	var err error

	if schemaFile != "" && dataFile != "" {
		// Parse from separated files
		data, err = parser.ParseFromFiles(schemaFile, dataFile)
	} else {
		// Parse from single file
		data, err = parser.ParseMetaDat(metadatContent)
	}

	if err != nil {
		return "", fmt.Errorf("failed to parse MetaDat: %v", err)
	}

	// Convert to JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to convert to JSON: %v", err)
	}

	return string(jsonBytes), nil
}

func parseMetaDat(metadatContent, schemaFile, dataFile string) (string, error) {
	// Create parser
	parser := metadat.NewParser()
	var data map[string]interface{}
	var err error

	if schemaFile != "" && dataFile != "" {
		// Parse from separated files
		data, err = parser.ParseFromFiles(schemaFile, dataFile)
	} else {
		// Parse from single file
		data, err = parser.ParseMetaDat(metadatContent)
	}

	if err != nil {
		return "", fmt.Errorf("failed to parse MetaDat: %v", err)
	}

	// Display structure
	result := fmt.Sprintf("Successfully parsed MetaDat file\n")
	result += fmt.Sprintf("Fields found: %d\n\n", len(data))

	for key, value := range data {
		result += fmt.Sprintf("Field: %s\n", key)
		result += fmt.Sprintf("Type: %T\n", value)
		
		// Show sample value
		switch v := value.(type) {
		case []interface{}:
			result += fmt.Sprintf("Length: %d\n", len(v))
			if len(v) > 0 {
				result += fmt.Sprintf("First element: %v (%T)\n", v[0], v[0])
			}
		case map[string]interface{}:
			result += fmt.Sprintf("Object fields: %d\n", len(v))
			for objKey := range v {
				result += fmt.Sprintf("  - %s\n", objKey)
			}
		case string:
			if len(v) > 50 {
				result += fmt.Sprintf("Value: %s... (truncated)\n", v[:50])
			} else {
				result += fmt.Sprintf("Value: %s\n", v)
			}
		default:
			result += fmt.Sprintf("Value: %v\n", v)
		}
		result += "\n"
	}

	return result, nil
}

func validateMetaDat(metadatContent, schemaFile, dataFile string) (string, error) {
	// Create parser
	parser := metadat.NewParser()
	var err error

	if schemaFile != "" && dataFile != "" {
		// Validate separated files
		_, err = parser.ParseFromFiles(schemaFile, dataFile)
	} else {
		// Validate single file
		_, err = parser.ParseMetaDat(metadatContent)
	}

	if err != nil {
		return "", fmt.Errorf("validation failed: %v", err)
	}

	return "✓ MetaDat file is valid\n", nil
}

func autoConvert(content string, separated bool, schemaFile, dataFile string) (string, error) {
	// Try to detect format by parsing as JSON first
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonData); err == nil {
		// It's valid JSON, convert to MetaDat
		return convertJSONToMetaDat(content, separated, schemaFile, dataFile)
	}

	// Try to parse as MetaDat
	parser := metadat.NewParser()
	if _, err := parser.ParseMetaDat(content); err == nil {
		// It's valid MetaDat, convert to JSON
		return convertMetaDatToJSON(content, schemaFile, dataFile)
	}

	return "", fmt.Errorf("unable to detect input format (not valid JSON or MetaDat)")
}