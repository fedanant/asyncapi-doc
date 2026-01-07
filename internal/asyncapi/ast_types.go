package asyncapi

import (
	"time"
)

// TypeInfo holds information extracted from type checking.
type TypeInfo struct {
	Name   string
	Fields []FieldInfo
}

// FieldInfo holds information about a struct field.
type FieldInfo struct {
	Name     string
	Type     string
	JSONTag  string
	IsArray  bool
	IsPtr    bool
	ElemType string
}

// CreateStructFromTypeInfo creates a struct instance based on TypeInfo.
func CreateStructFromTypeInfo(typeInfo *TypeInfo) interface{} {
	if typeInfo == nil {
		return struct{}{}
	}

	// Create a map to represent the struct
	result := make(map[string]interface{})

	for _, field := range typeInfo.Fields {
		jsonName := field.JSONTag
		if jsonName == "" || jsonName == "-" {
			continue
		}

		// Remove omitempty and other flags
		for i := 0; i < len(jsonName); i++ {
			if jsonName[i] == ',' {
				jsonName = jsonName[:i]
				break
			}
		}

		result[jsonName] = getDefaultValueForType(field.Type, field.IsArray)
	}

	return result
}

func getDefaultValueForType(typeName string, isArray bool) interface{} {
	if isArray {
		return []interface{}{}
	}

	// Handle pointer types
	if typeName != "" && typeName[0] == '*' {
		typeName = typeName[1:]
	}

	switch typeName {
	case "string":
		return ""
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return 0
	case "float32", "float64":
		return 0.0
	case "bool":
		return false
	case "time.Time":
		return time.Now()
	case "[]byte":
		return []byte{}
	default:
		// For complex types, return an empty map
		return map[string]interface{}{}
	}
}
