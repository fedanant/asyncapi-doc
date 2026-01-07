package asyncapi

import (
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"reflect"
	"time"
)

// TypeChecker wraps go/types functionality for extracting type information.
type TypeChecker struct {
	fset *token.FileSet
	pkg  *types.Package
	info *types.Info
}

// NewTypeChecker creates a new TypeChecker from parsed files.
func NewTypeChecker(fset *token.FileSet, files []*ast.File, pkgPath string) (*TypeChecker, error) {
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	config := &types.Config{
		Importer: importer.Default(),
		Error: func(_ error) {
			// Ignore errors for now - we want to be lenient
		},
	}

	pkg, err := config.Check(pkgPath, fset, files, info)
	_ = err // Intentionally ignored - we create a package even if type checking fails
	if pkg == nil {
		// If type checking fails, create an empty package
		pkg = types.NewPackage(pkgPath, pkgPath)
	}

	return &TypeChecker{
		fset: fset,
		pkg:  pkg,
		info: info,
	}, nil
}

// ExtractTypeInfo extracts type information from a named type.
func (tc *TypeChecker) ExtractTypeInfo(typeName string) *TypeInfo {
	obj := tc.pkg.Scope().Lookup(typeName)
	if obj == nil {
		return nil
	}

	named, ok := obj.Type().(*types.Named)
	if !ok {
		return nil
	}

	underlying := named.Underlying()
	structType, ok := underlying.(*types.Struct)
	if !ok {
		return nil
	}

	typeInfo := &TypeInfo{
		Name:   typeName,
		Fields: []FieldInfo{},
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if !field.Exported() {
			continue
		}

		fieldInfo := FieldInfo{
			Name: field.Name(),
		}

		// Extract JSON tag
		tag := structType.Tag(i)
		fieldInfo.JSONTag = extractJSONTagFromReflect(tag)

		// Extract type information
		fieldInfo.Type, fieldInfo.IsArray, fieldInfo.IsPtr, fieldInfo.ElemType = tc.extractFieldTypeInfo(field.Type())

		typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
	}

	return typeInfo
}

// extractFieldTypeInfo extracts type information from a types.Type.
func (tc *TypeChecker) extractFieldTypeInfo(typ types.Type) (typeName string, isArray, isPtr bool, elemType string) {
	switch t := typ.(type) {
	case *types.Basic:
		return t.Name(), false, false, ""
	case *types.Slice:
		elemTypeName, _, _, _ := tc.extractFieldTypeInfo(t.Elem())
		return "[]" + elemTypeName, true, false, elemTypeName
	case *types.Pointer:
		elemTypeName, isArr, _, elem := tc.extractFieldTypeInfo(t.Elem())
		return "*" + elemTypeName, isArr, true, elem
	case *types.Named:
		// Handle named types like time.Time
		obj := t.Obj()
		if obj.Pkg() != nil && obj.Pkg().Name() != tc.pkg.Name() {
			return obj.Pkg().Name() + "." + obj.Name(), false, false, ""
		}
		return obj.Name(), false, false, ""
	case *types.Array:
		elemTypeName, _, _, _ := tc.extractFieldTypeInfo(t.Elem())
		return "[]" + elemTypeName, true, false, elemTypeName
	}
	return "interface{}", false, false, ""
}

// GetReflectType converts a TypeInfo to a reflect.Type.
func (tc *TypeChecker) GetReflectType(typeInfo *TypeInfo) reflect.Type {
	if typeInfo == nil {
		return reflect.TypeOf(struct{}{})
	}

	var fields []reflect.StructField

	for _, field := range typeInfo.Fields {
		jsonTag := field.JSONTag
		if jsonTag == "-" {
			continue
		}

		if jsonTag == "" {
			jsonTag = field.Name
		}

		fieldType := tc.getReflectTypeFromString(field.Type, field.IsArray, field.ElemType)

		structField := reflect.StructField{
			Name: field.Name,
			Type: fieldType,
			Tag:  reflect.StructTag(`json:"` + jsonTag + `"`),
		}

		fields = append(fields, structField)
	}

	if len(fields) == 0 {
		return reflect.TypeOf(struct{}{})
	}

	return reflect.StructOf(fields)
}

// getReflectTypeFromString converts a type string to reflect.Type.
//
//nolint:gocyclo // Type mapping logic is intentionally centralized for maintainability
func (tc *TypeChecker) getReflectTypeFromString(typeName string, isArray bool, elemType string) reflect.Type {
	// Handle pointer types
	if typeName != "" && typeName[0] == '*' {
		typeName = typeName[1:]
	}

	// If it's an array type and starts with []
	if len(typeName) > 2 && typeName[:2] == "[]" {
		typeName = typeName[2:]
		isArray = true
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
		// Try to look up nested type
		if elemType != "" {
			nestedTypeInfo := tc.ExtractTypeInfo(elemType)
			if nestedTypeInfo != nil {
				baseType = tc.GetReflectType(nestedTypeInfo)
			} else {
				baseType = reflect.TypeOf((*interface{})(nil)).Elem()
			}
		} else {
			baseType = reflect.TypeOf((*interface{})(nil)).Elem()
		}
	}

	if isArray {
		return reflect.SliceOf(baseType)
	}

	return baseType
}

// extractJSONTagFromReflect extracts JSON tag from a reflect-style tag string.
func extractJSONTagFromReflect(tag string) string {
	// Use reflect.StructTag to parse the tag
	st := reflect.StructTag(tag)
	jsonTag := st.Get("json")

	if jsonTag == "" || jsonTag == "-" {
		return jsonTag
	}

	// Extract just the field name (before comma)
	for i := 0; i < len(jsonTag); i++ {
		if jsonTag[i] == ',' {
			return jsonTag[:i]
		}
	}

	return jsonTag
}
