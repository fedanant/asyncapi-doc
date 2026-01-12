package asyncapi

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestNewOperation(t *testing.T) {
	op := NewOperation()

	if op == nil {
		t.Fatal("NewOperation returned nil")
	}

	if op.TypeOperation != "sub" {
		t.Errorf("Default TypeOperation = %q, want %q", op.TypeOperation, "sub")
	}

	if op.Message == nil {
		t.Error("Message should be initialized")
	}

	if op.MessageResponse == nil {
		t.Error("MessageResponse should be initialized")
	}

	if op.Parameters == nil {
		t.Error("Parameters should be initialized")
	}
}

func TestParseType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"pub", "pub"},
		{"sub", "sub"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			op := NewOperation()
			op.ParseType(tt.input)

			if op.TypeOperation != tt.want {
				t.Errorf("TypeOperation = %q, want %q", op.TypeOperation, tt.want)
			}
		})
	}
}

func TestParseName(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantName   string
		wantParams int
	}{
		{
			name:       "simple name",
			input:      "user.created",
			wantName:   "user.created",
			wantParams: 0,
		},
		{
			name:       "name with one parameter",
			input:      "user.{userId}.updated",
			wantName:   "user.{userId}.updated",
			wantParams: 1,
		},
		{
			name:       "name with multiple parameters",
			input:      "order.{orderId}.item.{itemId}",
			wantName:   "order.{orderId}.item.{itemId}",
			wantParams: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewOperation()
			op.ParseName(tt.input)

			if op.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", op.Name, tt.wantName)
			}

			if len(op.Parameters) != tt.wantParams {
				t.Errorf("Parameters count = %d, want %d", len(op.Parameters), tt.wantParams)
			}
		})
	}
}

func TestParseDescription(t *testing.T) {
	op := NewOperation()
	description := "This is a test description"

	op.ParseDescription(description)

	if op.Message.Description != description {
		t.Errorf("Description = %q, want %q", op.Message.Description, description)
	}
}

func TestParseSummary(t *testing.T) {
	op := NewOperation()
	summary := "Test summary"

	op.ParseSummary(summary)

	if op.Message.Summary != summary {
		t.Errorf("Summary = %q, want %q", op.Message.Summary, summary)
	}
}

func TestParseComment(t *testing.T) {
	// Create a simple test package with a type
	src := `
package testpkg

type TestEvent struct {
	ID   string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	tc, err := NewTypeChecker(fset, []*ast.File{file}, "testpkg")
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	tests := []struct {
		name    string
		comment string
		check   func(*testing.T, *Operation)
	}{
		{
			name:    "parse type attribute",
			comment: "@type pub",
			check: func(t *testing.T, op *Operation) {
				if op.TypeOperation != "pub" {
					t.Errorf("TypeOperation = %q, want %q", op.TypeOperation, "pub")
				}
			},
		},
		{
			name:    "parse name attribute",
			comment: "@name user.created",
			check: func(t *testing.T, op *Operation) {
				if op.Name != "user.created" {
					t.Errorf("Name = %q, want %q", op.Name, "user.created")
				}
			},
		},
		{
			name:    "parse summary attribute",
			comment: "@summary User created event",
			check: func(t *testing.T, op *Operation) {
				if op.Message.Summary != "User created event" {
					t.Errorf("Summary = %q, want %q", op.Message.Summary, "User created event")
				}
			},
		},
		{
			name:    "parse description attribute",
			comment: "@description This is a description",
			check: func(t *testing.T, op *Operation) {
				if op.Message.Description != "This is a description" {
					t.Errorf("Description = %q, want %q", op.Message.Description, "This is a description")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewOperation()
			err := op.ParseComment(tt.comment, tc)
			if err != nil {
				t.Errorf("ParseComment() error = %v", err)
			}
			tt.check(t, op)
		})
	}
}

func TestParseCommentWithEmptyLine(t *testing.T) {
	op := NewOperation()
	err := op.ParseComment("", nil)

	if err != nil {
		t.Errorf("ParseComment with empty string should not error, got: %v", err)
	}
}

func TestTransToReflectType(t *testing.T) {
	tests := []struct {
		typeName string
		wantNil  bool
	}{
		{"int", false},
		{"uint", false},
		{"int8", false},
		{"uint8", false},
		{"int16", false},
		{"uint16", false},
		{"int32", false},
		{"uint32", false},
		{"int64", false},
		{"uint64", false},
		{"float32", false},
		{"float64", false},
		{"bool", false},
		{"string", false},
		{"unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := TransToReflectType(tt.typeName)

			if tt.wantNil && result != nil {
				t.Errorf("TransToReflectType(%q) = %v, want nil", tt.typeName, result)
			}

			if !tt.wantNil && result == nil {
				t.Errorf("TransToReflectType(%q) = nil, want non-nil", tt.typeName)
			}
		})
	}
}

func TestParseNameWithParameters(t *testing.T) {
	op := NewOperation()
	op.ParseName("user.{userId}.order.{orderId}")

	if len(op.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(op.Parameters))
	}

	if _, exists := op.Parameters["userId"]; !exists {
		t.Error("Parameter 'userId' should exist")
	}

	if _, exists := op.Parameters["orderId"]; !exists {
		t.Error("Parameter 'orderId' should exist")
	}

	// Check parameter schema
	if param, exists := op.Parameters["userId"]; exists {
		if param.Schema == nil {
			t.Error("Parameter schema should not be nil")
		}

		if paramType, ok := param.Schema["type"].(string); !ok || paramType != "string" {
			t.Errorf("Parameter type = %v, want 'string'", param.Schema["type"])
		}
	}
}

func TestParsePayloadWithInvalidType(t *testing.T) {
	op := NewOperation()

	// Create empty type checker
	fset := token.NewFileSet()
	tc, err := NewTypeChecker(fset, []*ast.File{}, "testpkg")
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	// Note: GetByNameType returns struct{}{} for unknown types instead of nil
	// So ParsePayload will succeed but with an empty struct
	// This test documents the current behavior
	err = op.ParsePayload("NonExistentType", tc)

	// The function returns nil error because GetByNameType always returns a value
	if err != nil {
		t.Logf("Got error (expected due to current implementation): %v", err)
	}

	// Verify that some message sample was set (even if it's empty struct)
	if op.Message.MessageSample == nil {
		t.Error("MessageSample should be set even for unknown types")
	}
}

func TestParseResponseWithInvalidType(t *testing.T) {
	op := NewOperation()

	// Create empty type checker
	fset := token.NewFileSet()
	tc, err := NewTypeChecker(fset, []*ast.File{}, "testpkg")
	if err != nil {
		t.Fatalf("Failed to create type checker: %v", err)
	}

	// Note: GetByNameType returns struct{}{} for unknown types instead of nil
	// So ParseResponse will succeed but with an empty struct
	err = op.ParseResponse("NonExistentType", tc)

	// The function returns nil error because GetByNameType always returns a value
	if err != nil {
		t.Logf("Got error (expected due to current implementation): %v", err)
	}

	// Verify that some message sample was set (even if it's empty struct)
	if op.MessageResponse.MessageSample == nil {
		t.Error("MessageSample should be set even for unknown types")
	}
}
