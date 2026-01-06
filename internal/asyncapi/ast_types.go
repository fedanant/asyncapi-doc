package asyncapi

import (
	"go/ast"
	"go/token"
	"log"
	"reflect"
	"time"
)

// TypeInfo holds information extracted from AST
type TypeInfo struct {
	Name   string
	Fields []FieldInfo
}

type FieldInfo struct {
	Name     string
	Type     string
	JSONTag  string
	IsArray  bool
	IsPtr    bool
	ElemType string
}

// ExtractTypeFromAST extracts type information from the AST
func ExtractTypeFromAST(typeName string, pkg *ast.Package) *TypeInfo {
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != typeName {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				typeInfo := &TypeInfo{
					Name:   typeName,
					Fields: []FieldInfo{},
				}

				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						continue
					}

					fieldInfo := FieldInfo{
						Name: field.Names[0].Name,
					}

					// Extract JSON tag
					if field.Tag != nil {
						tag := field.Tag.Value
						// Simple JSON tag extraction
						fieldInfo.JSONTag = extractJSONTag(tag)
					}

					// Extract type information
					fieldInfo.Type, fieldInfo.IsArray, fieldInfo.IsPtr, fieldInfo.ElemType = extractFieldType(field.Type)

					typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
				}

				return typeInfo
			}
		}
	}

	return nil
}

func extractJSONTag(tag string) string {
	// Remove backticks
	if len(tag) > 0 && tag[0] == '`' {
		tag = tag[1 : len(tag)-1]
	}

	// Find json:"..." part
	jsonPrefix := `json:"`
	start := 0
	for i := 0; i < len(tag); i++ {
		if i+len(jsonPrefix) <= len(tag) && tag[i:i+len(jsonPrefix)] == jsonPrefix {
			start = i + len(jsonPrefix)
			break
		}
	}

	if start == 0 {
		return ""
	}

	// Find the closing quote
	end := start
	for end < len(tag) && tag[end] != '"' {
		end++
	}

	jsonTag := tag[start:end]

	// Extract just the field name (before comma)
	for i := 0; i < len(jsonTag); i++ {
		if jsonTag[i] == ',' {
			return jsonTag[:i]
		}
	}

	return jsonTag
}

func extractFieldType(expr ast.Expr) (typeName string, isArray bool, isPtr bool, elemType string) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, false, false, ""
	case *ast.ArrayType:
		elemTypeName, _, _, _ := extractFieldType(t.Elt)
		return "[]" + elemTypeName, true, false, elemTypeName
	case *ast.StarExpr:
		elemTypeName, isArr, _, elem := extractFieldType(t.X)
		return "*" + elemTypeName, isArr, true, elem
	case *ast.SelectorExpr:
		// e.g., time.Time
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name, false, false, ""
		}
		return t.Sel.Name, false, false, ""
	}
	return "interface{}", false, false, ""
}

// CreateStructFromTypeInfo creates a struct instance based on TypeInfo
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
	if len(typeName) > 0 && typeName[0] == '*' {
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

// CreateReflectValue creates a reflect.Value from TypeInfo
func CreateReflectValue(typeInfo *TypeInfo) interface{} {
	if typeInfo == nil {
		return struct{}{}
	}

	// Build struct fields
	var fields []reflect.StructField

	for _, field := range typeInfo.Fields {
		jsonTag := field.JSONTag
		if jsonTag == "-" {
			continue
		}

		if jsonTag == "" {
			jsonTag = field.Name
		}

		fieldType := getReflectType(field.Type, field.IsArray)

		structField := reflect.StructField{
			Name: field.Name,
			Type: fieldType,
			Tag:  reflect.StructTag(`json:"` + jsonTag + `"`),
		}

		fields = append(fields, structField)
	}

	if len(fields) == 0 {
		return struct{}{}
	}

	// Create the struct type
	structType := reflect.StructOf(fields)

	// Create an instance of the struct
	instance := reflect.New(structType).Elem()

	return instance.Interface()
}

func getReflectType(typeName string, isArray bool) reflect.Type {
	// Handle pointer types
	if len(typeName) > 0 && typeName[0] == '*' {
		typeName = typeName[1:]
	}

	var baseType reflect.Type

	switch typeName {
	case "string":
		baseType = reflect.TypeOf("")
	case "int":
		baseType = reflect.TypeOf(int(0))
	case "int8":
		baseType = reflect.TypeOf(int8(0))
	case "int16":
		baseType = reflect.TypeOf(int16(0))
	case "int32":
		baseType = reflect.TypeOf(int32(0))
	case "int64":
		baseType = reflect.TypeOf(int64(0))
	case "uint":
		baseType = reflect.TypeOf(uint(0))
	case "uint8":
		baseType = reflect.TypeOf(uint8(0))
	case "uint16":
		baseType = reflect.TypeOf(uint16(0))
	case "uint32":
		baseType = reflect.TypeOf(uint32(0))
	case "uint64":
		baseType = reflect.TypeOf(uint64(0))
	case "float32":
		baseType = reflect.TypeOf(float32(0))
	case "float64":
		baseType = reflect.TypeOf(float64(0))
	case "bool":
		baseType = reflect.TypeOf(false)
	case "time.Time":
		baseType = reflect.TypeOf(time.Time{})
	default:
		// For unknown types, use interface{}
		log.Printf("Unknown type '%s', using interface{}", typeName)
		baseType = reflect.TypeOf((*interface{})(nil)).Elem()
	}

	if isArray {
		return reflect.SliceOf(baseType)
	}

	return baseType
}
