package asyncapi

import "go/ast"

// TypeSpecDef the whole information of a typeSpec.
type TypeSpecDef struct {
	// ast file where TypeSpec is
	File *ast.File

	// the TypeSpec of this type definition
	TypeSpec *ast.TypeSpec

	// path of package starting from under ${GOPATH}/src or from module path in go.mod
	PkgPath    string
	ParentSpec ast.Decl
	Payload    interface{}

	NotUnique bool
}

func (t *TypeSpecDef) Name() string {
	if t.TypeSpec != nil {
		return t.TypeSpec.Name.Name
	}

	return ""
}
