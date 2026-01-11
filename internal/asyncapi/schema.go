package asyncapi

import (
	"reflect"
	"strconv"
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

	//nolint:exhaustive // Only handling common types; default case handles others
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

		// Apply struct field tags
		applyFieldTags(fieldSchema, field)

		properties[jsonName] = fieldSchema

		// Check for explicit required tag
		if requiredTag := field.Tag.Get("required"); requiredTag == "true" {
			isRequired = true
		}

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

// applyFieldTags applies struct field tags to the field schema.
//
//nolint:gocritic // Passing by value is acceptable for this use case
func applyFieldTags(schema map[string]interface{}, field reflect.StructField) {
	// Apply format tag
	if format := field.Tag.Get("format"); format != "" {
		schema["format"] = format
	}

	// Apply example tag
	if example := field.Tag.Get("example"); example != "" {
		schema["example"] = parseExampleValue(example, schema)
	}

	// Apply description tag
	if description := field.Tag.Get("description"); description != "" {
		schema["description"] = description
	}

	// Apply validate tag
	if validate := field.Tag.Get("validate"); validate != "" {
		applyValidationRules(schema, validate)
	}
}

// parseExampleValue converts the example string to the appropriate type.
func parseExampleValue(example string, schema map[string]interface{}) interface{} {
	schemaType, ok := schema["type"].(string)
	if !ok {
		return example
	}

	switch schemaType {
	case "integer":
		if val, err := strconv.ParseInt(example, 10, 64); err == nil {
			return val
		}
	case "number":
		if val, err := strconv.ParseFloat(example, 64); err == nil {
			return val
		}
	case "boolean":
		if val, err := strconv.ParseBool(example); err == nil {
			return val
		}
	}
	return example
}

// applyValidationRules parses validation rules and applies them to the schema.
// Supports both custom validation format and go-playground/validator tags.
//
//nolint:gocyclo // Complex validation logic is intentionally centralized
func applyValidationRules(schema map[string]interface{}, validate string) {
	rules := strings.Split(validate, ",")
	schemaType, ok := schema["type"].(string)
	if !ok {
		schemaType = ""
	}

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		// Parse rule format: "key=value" or "key"
		parts := strings.SplitN(rule, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := ""
		if len(parts) > 1 {
			value = strings.TrimSpace(parts[1])
		}

		switch key {
		// Numeric comparisons (go-playground/validator compatible)
		case "min":
			if schemaType == "string" || schemaType == "array" {
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					if schemaType == "string" {
						schema["minLength"] = val
					} else {
						schema["minItems"] = val
					}
				}
			} else {
				if val, err := strconv.ParseFloat(value, 64); err == nil {
					schema["minimum"] = val
				}
			}
		case "max":
			if schemaType == "string" || schemaType == "array" {
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					if schemaType == "string" {
						schema["maxLength"] = val
					} else {
						schema["maxItems"] = val
					}
				}
			} else {
				if val, err := strconv.ParseFloat(value, 64); err == nil {
					schema["maximum"] = val
				}
			}
		case "gt":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				schema["exclusiveMinimum"] = val
			}
		case "gte":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				schema["minimum"] = val
			}
		case "lt":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				schema["exclusiveMaximum"] = val
			}
		case "lte":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				schema["maximum"] = val
			}

		// Length validations
		case "minLength":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				schema["minLength"] = val
			}
		case "maxLength":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				schema["maxLength"] = val
			}
		case "len":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				if schemaType == "string" {
					schema["minLength"] = val
					schema["maxLength"] = val
				} else if schemaType == "array" {
					schema["minItems"] = val
					schema["maxItems"] = val
				}
			}

		// Enum validations
		case "oneof", "oneOf":
			if value != "" {
				enumValues := strings.Split(value, "|")
				var typedEnums []interface{}
				for _, v := range enumValues {
					v = strings.TrimSpace(v)
					typedEnums = append(typedEnums, convertToType(v, schemaType))
				}
				if len(typedEnums) > 0 {
					schema["enum"] = typedEnums
				}
			}
		case "eq":
			if value != "" {
				schema["const"] = convertToType(value, schemaType)
			}

		// String patterns
		case "alpha":
			schema["pattern"] = "^[a-zA-Z]+$"
		case "alphanum":
			schema["pattern"] = "^[a-zA-Z0-9]+$"
		case "alphaspace":
			schema["pattern"] = "^[a-zA-Z ]+$"
		case "alphanumunicode":
			schema["pattern"] = "^[\\p{L}\\p{N}]+$"
		case "lowercase":
			schema["pattern"] = "^[a-z]+$"
		case "uppercase":
			schema["pattern"] = "^[A-Z]+$"
		case "numeric":
			schema["pattern"] = "^[0-9]+$"
		case "hexadecimal":
			schema["pattern"] = "^[0-9a-fA-F]+$"
		case "hexcolor":
			schema["pattern"] = "^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"
		case "ascii":
			schema["pattern"] = "^[\x00-\x7F]+$"
		case "printascii":
			schema["pattern"] = "^[\x20-\x7E]+$"
		case "startswith":
			if value != "" {
				schema["pattern"] = "^" + escapeRegex(value)
			}
		case "endswith":
			if value != "" {
				schema["pattern"] = escapeRegex(value) + "$"
			}
		case "contains":
			if value != "" {
				schema["pattern"] = escapeRegex(value)
			}
		case "pattern":
			schema["pattern"] = value

		// Format validations (go-playground/validator compatible)
		case "email":
			schema["format"] = "email"
		case "url", "uri", "http_url":
			schema["format"] = "uri"
		case "uuid", "uuid4", "uuid_rfc4122":
			schema["format"] = "uuid"
		case "uuid3", "uuid3_rfc4122":
			schema["format"] = "uuid"
		case "uuid5", "uuid5_rfc4122":
			schema["format"] = "uuid"
		case "datetime":
			schema["format"] = "date-time"
		case "date":
			schema["format"] = "date"
		case "time":
			schema["format"] = "time"
		case "duration":
			schema["format"] = "duration"
		case "hostname", "fqdn", "hostname_rfc1123":
			schema["format"] = "hostname"
		case "ipv4", "ip4_addr":
			schema["format"] = "ipv4"
		case "ipv6", "ip6_addr":
			schema["format"] = "ipv6"
		case "ip", "ip_addr":
			schema["format"] = "ipv4"
		case "base64", "base64url":
			schema["format"] = "base64"
		case "datauri":
			schema["format"] = "data-uri"
		case "json":
			schema["contentMediaType"] = "application/json"
		case "jwt":
			schema["pattern"] = "^[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]*$"

		// Geographic
		case "latitude":
			schema["minimum"] = -90.0
			schema["maximum"] = 90.0
		case "longitude":
			schema["minimum"] = -180.0
			schema["maximum"] = 180.0

		// Network
		case "mac":
			schema["pattern"] = "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"
		case "cidr":
			schema["pattern"] = "^([0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2}$"
		case "port":
			schema["minimum"] = 1.0
			schema["maximum"] = 65535.0

		// ISBN/ISSN
		case "isbn":
			schema["pattern"] = "^(?:ISBN(?:-1[03])?:? )?(?=[0-9X]{10}$|(?=(?:[0-9]+[- ]){3})[- 0-9X]{13}$|97[89][0-9]{10}$|(?=(?:[0-9]+[- ]){4})[- 0-9]{17}$)(?:97[89][- ]?)?[0-9]{1,5}[- ]?[0-9]+[- ]?[0-9]+[- ]?[0-9X]$"
		case "isbn10":
			schema["pattern"] = "^(?:[0-9]{9}X|[0-9]{10})$"
		case "isbn13":
			schema["pattern"] = "^(?:97[89][0-9]{10})$"
		case "issn":
			schema["pattern"] = "^[0-9]{4}-[0-9]{3}[0-9X]$"

		// Credit card
		case "credit_card":
			schema["pattern"] = "^[0-9]{13,19}$"

		// Bitcoin
		case "btc_addr":
			schema["pattern"] = "^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$"

		// Ethereum
		case "eth_addr":
			schema["pattern"] = "^0x[0-9a-fA-F]{40}$"

		// SSN
		case "ssn":
			schema["pattern"] = "^[0-9]{3}-[0-9]{2}-[0-9]{4}$"

		// Semantic versioning
		case "semver":
			schema["pattern"] = "^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"

		// Phone number
		case "e164":
			schema["pattern"] = "^\\+[1-9]\\d{1,14}$"

		// Array specific
		case "unique":
			schema["uniqueItems"] = true
		case "dive":
			// dive is handled at the array level, not individual item level
			// This is a marker for nested validation
		}
	}
}

// convertToType converts a string value to the appropriate type based on schema type.
func convertToType(value, schemaType string) interface{} {
	switch schemaType {
	case "integer":
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	case "number":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	case "boolean":
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return value
}

// escapeRegex escapes special regex characters in a string.
func escapeRegex(s string) string {
	special := []string{".", "+", "*", "?", "^", "$", "(", ")", "[", "]", "{", "}", "|", "\\"}
	result := s
	for _, char := range special {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

func generateArraySchema(val reflect.Value) map[string]interface{} {
	var itemsSchema map[string]interface{}

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

func generateMapSchema(_ reflect.Value) map[string]interface{} {
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

	//nolint:exhaustive // Only handling common types; default case handles others
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
