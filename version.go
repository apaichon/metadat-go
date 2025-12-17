package metadat

// Version information for the MetaDat Go library
const (
	// Version is the current version of the MetaDat Go library
	Version = "1.1.0"
	
	// Name is the library name
	Name = "metadat-go"
	
	// Description is a brief description of the library
	Description = "MetaDat format parser and writer for Go"
)

// GetVersion returns version information
func GetVersion() string {
	return Version
}

// GetVersionInfo returns detailed version information
func GetVersionInfo() map[string]string {
	return map[string]string{
		"version":     Version,
		"name":        Name,
		"description": Description,
	}
}