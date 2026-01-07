package asyncapi

import (
	"reflect"
	"strings"
	"time"
)

// GenerateJSONSchema converts a struct instance to a JSON Schema definition.
// This creates a proper schema with type, properties, etc. instead of example values.
// It unwraps Msg and MsgResponse wrapper types to return only the inner payload schema.
func GenerateJSONSchema(v interface{}) map[string]interface{} {
	if v == nil {
		return map[string]interface{}{
			"type": "object",
		}
	}

	val := reflect.ValueOf(v)
	typ := val.Type()

	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		if val.IsNil() {
			return map[string]interface{}{
				"type": "object",
			}
		}
		val = val.Elem()
		typ = val.Type()
	}

	// Handle the Msg and MsgResponse wrapper types - unwrap and return inner schema
	if typ.Kind() == reflect.Struct && typ.NumField() > 0 {
		// Check if this is a Msg wrapper (has Data field as first field)
		firstField := typ.Field(0)
		if firstField.Name == "Data" {
			// Unwrap and process the inner data
			innerVal := val.Field(0)

			// Get the inner value's actual type to generate full schema
			innerType := innerVal.Type()

			// For interface{} types, we need to get the concrete value
			if innerType.Kind() == reflect.Interface && !innerVal.IsNil() {
				innerVal = innerVal.Elem()
			}

			// Return only the inner schema without the wrapper
			return generateSchemaForValue(innerVal)
		}

		// Check if this is a MsgResponse wrapper (has Response field)
		// MsgResponse has both Id and Response fields, we only want the Response content
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.Name == "Response" {
				// Unwrap and process the Response field
				innerVal := val.Field(i)

				// Get the inner value's actual type to generate full schema
				innerType := innerVal.Type()

				// For interface{} types, we need to get the concrete value
				if innerType.Kind() == reflect.Interface && !innerVal.IsNil() {
					innerVal = innerVal.Elem()
				}

				// Return only the inner schema without the wrapper
				return generateSchemaForValue(innerVal)
			}
		}
	}

	return generateSchemaForValue(val)
}

func generateSchemaForValue(val reflect.Value) map[string]interface{} {
	typ := val.Type()

	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		if val.IsNil() {
			return map[string]interface{}{
				"type": "null",
			}
		}
		val = val.Elem()
		typ = val.Type()
	}

	switch typ.Kind() {
	case reflect.Struct:
		return generateObjectSchema(val)
	case reflect.Slice, reflect.Array:
		return generateArraySchema(val)
	case reflect.Map:
		return generateMapSchema(val)
	case reflect.String:
		return map[string]interface{}{
			"type": "string",
		}
	case reflect.Bool:
		return map[string]interface{}{
			"type": "boolean",
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{
			"type": "integer",
		}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{
			"type": "number",
		}
	default:
		return map[string]interface{}{
			"type": "object",
		}
	}
}

func generateObjectSchema(val reflect.Value) map[string]interface{} {
	typ := val.Type()

	// Special handling for time.Time
	if typ == reflect.TypeOf(time.Time{}) {
		return map[string]interface{}{
			"type":   "string",
			"format": "date-time",
		}
	}

	properties := make(map[string]interface{})
	required := []string{}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse JSON tag (e.g., "fieldName,omitempty")
		jsonName := jsonTag
		isRequired := true
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonName = jsonTag[:idx]
			options := jsonTag[idx+1:]
			if strings.Contains(options, "omitempty") {
				isRequired = false
			}
		}

		// Generate schema for field
		fieldSchema := generateSchemaForValue(fieldVal)
		properties[jsonName] = fieldSchema

		if isRequired {
			required = append(required, jsonName)
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func generateArraySchema(val reflect.Value) map[string]interface{} {
	itemsSchema := map[string]interface{}{
		"type": "object",
	}

	// If array has elements, use the first element to generate schema
	if val.Len() > 0 {
		itemsSchema = generateSchemaForValue(val.Index(0))
	} else {
		// For empty arrays, try to infer from type
		elemType := val.Type().Elem()
		if elemType.Kind() == reflect.Struct {
			// Create a zero value to generate schema
			zeroVal := reflect.New(elemType).Elem()
			itemsSchema = generateSchemaForValue(zeroVal)
		} else {
			itemsSchema = generateSchemaForType(elemType)
		}
	}

	return map[string]interface{}{
		"type":  "array",
		"items": itemsSchema,
	}
}

func generateMapSchema(val reflect.Value) map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"additionalProperties": map[string]interface{}{
			"type": "object",
		},
	}
}

func generateSchemaForType(typ reflect.Type) map[string]interface{} {
	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	switch typ.Kind() {
	case reflect.String:
		return map[string]interface{}{
			"type": "string",
		}
	case reflect.Bool:
		return map[string]interface{}{
			"type": "boolean",
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{
			"type": "integer",
		}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{
			"type": "number",
		}
	case reflect.Struct:
		if typ == reflect.TypeOf(time.Time{}) {
			return map[string]interface{}{
				"type":   "string",
				"format": "date-time",
			}
		}
		// Create a zero value and generate schema
		zeroVal := reflect.New(typ).Elem()
		return generateObjectSchema(zeroVal)
	default:
		return map[string]interface{}{
			"type": "object",
		}
	}
}
