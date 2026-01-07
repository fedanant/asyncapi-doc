package asyncapi

import (
	"reflect"
	"testing"
	"time"
)

func TestGenerateJSONSchema_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantType string
	}{
		{"nil value", nil, "object"},
		{"string", "test", "string"},
		{"bool", true, "boolean"},
		{"int", 42, "integer"},
		{"int64", int64(42), "integer"},
		{"float32", float32(3.14), "number"},
		{"float64", 3.14159, "number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateJSONSchema(tt.input)

			if schema == nil {
				t.Fatal("GenerateJSONSchema returned nil")
			}

			schemaType, ok := schema["type"].(string)
			if !ok {
				t.Errorf("Schema type is not a string: %v", schema["type"])
				return
			}

			if schemaType != tt.wantType {
				t.Errorf("Type = %q, want %q", schemaType, tt.wantType)
			}
		})
	}
}

func TestGenerateJSONSchema_Struct(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email,omitempty"`
	}

	input := TestStruct{
		Name:  "John",
		Age:   30,
		Email: "john@example.com",
	}

	schema := GenerateJSONSchema(input)

	if schema["type"] != "object" {
		t.Errorf("Type = %v, want 'object'", schema["type"])
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties is not a map")
	}

	if len(properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(properties))
	}

	// Check name property
	nameSchema, ok := properties["name"].(map[string]interface{})
	if !ok || nameSchema["type"] != "string" {
		t.Errorf("name property type = %v, want 'string'", nameSchema["type"])
	}

	// Check age property
	ageSchema, ok := properties["age"].(map[string]interface{})
	if !ok || ageSchema["type"] != "integer" {
		t.Errorf("age property type = %v, want 'integer'", ageSchema["type"])
	}

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Required is not a string slice")
	}

	// name and age should be required, email should not (has omitempty)
	if len(required) != 2 {
		t.Errorf("Expected 2 required fields, got %d: %v", len(required), required)
	}
}

func TestGenerateJSONSchema_TimeField(t *testing.T) {
	type EventStruct struct {
		Timestamp time.Time `json:"timestamp"`
	}

	input := EventStruct{
		Timestamp: time.Now(),
	}

	schema := GenerateJSONSchema(input)

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties is not a map")
	}

	timestampSchema, ok := properties["timestamp"].(map[string]interface{})
	if !ok {
		t.Fatal("timestamp property not found")
	}

	if timestampSchema["type"] != "string" {
		t.Errorf("timestamp type = %v, want 'string'", timestampSchema["type"])
	}

	if timestampSchema["format"] != "date-time" {
		t.Errorf("timestamp format = %v, want 'date-time'", timestampSchema["format"])
	}
}

func TestGenerateJSONSchema_Array(t *testing.T) {
	type Item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	input := []Item{
		{ID: "1", Name: "Item 1"},
		{ID: "2", Name: "Item 2"},
	}

	schema := GenerateJSONSchema(input)

	if schema["type"] != "array" {
		t.Errorf("Type = %v, want 'array'", schema["type"])
	}

	items, ok := schema["items"].(map[string]interface{})
	if !ok {
		t.Fatal("Items is not a map")
	}

	if items["type"] != "object" {
		t.Errorf("Items type = %v, want 'object'", items["type"])
	}

	properties, ok := items["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Items properties is not a map")
	}

	if len(properties) != 2 {
		t.Errorf("Expected 2 properties in items, got %d", len(properties))
	}
}

func TestGenerateJSONSchema_EmptyArray(t *testing.T) {
	type Item struct {
		ID string `json:"id"`
	}

	input := []Item{}

	schema := GenerateJSONSchema(input)

	if schema["type"] != "array" {
		t.Errorf("Type = %v, want 'array'", schema["type"])
	}

	items, ok := schema["items"].(map[string]interface{})
	if !ok {
		t.Fatal("Items is not a map")
	}

	// Should still generate object schema for empty array
	if items["type"] != "object" {
		t.Errorf("Items type = %v, want 'object'", items["type"])
	}
}

func TestGenerateJSONSchema_Pointer(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}

	value := TestStruct{Name: "test"}
	input := &value

	schema := GenerateJSONSchema(input)

	if schema["type"] != "object" {
		t.Errorf("Type = %v, want 'object'", schema["type"])
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties is not a map")
	}

	if len(properties) != 1 {
		t.Errorf("Expected 1 property, got %d", len(properties))
	}
}

func TestGenerateJSONSchema_NilPointer(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}

	var input *TestStruct = nil

	schema := GenerateJSONSchema(input)

	if schema["type"] != "object" {
		t.Errorf("Type = %v, want 'object'", schema["type"])
	}
}

func TestGenerateJSONSchema_MsgWrapper(t *testing.T) {
	type UserEvent struct {
		UserID string `json:"userId"`
		Email  string `json:"email"`
	}

	input := Msg{
		Data: UserEvent{
			UserID: "123",
			Email:  "test@example.com",
		},
	}

	schema := GenerateJSONSchema(input)

	// Should unwrap and return only the inner schema
	if schema["type"] != "object" {
		t.Errorf("Type = %v, want 'object'", schema["type"])
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties is not a map")
	}

	// Should have userId and email, not Data
	if _, hasData := properties["data"]; hasData {
		t.Error("Schema should not contain 'data' field from wrapper")
	}

	if _, hasUserID := properties["userId"]; !hasUserID {
		t.Error("Schema should contain 'userId' field from unwrapped data")
	}

	if _, hasEmail := properties["email"]; !hasEmail {
		t.Error("Schema should contain 'email' field from unwrapped data")
	}
}

func TestGenerateJSONSchema_MsgResponseWrapper(t *testing.T) {
	type UserResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	input := MsgResponse{
		Id: "request-123",
		Response: UserResponse{
			Success: true,
			Message: "OK",
		},
	}

	schema := GenerateJSONSchema(input)

	// Should unwrap and return only the Response schema
	if schema["type"] != "object" {
		t.Errorf("Type = %v, want 'object'", schema["type"])
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties is not a map")
	}

	// Should have success and message, not Id or Response
	if _, hasId := properties["id"]; hasId {
		t.Error("Schema should not contain 'id' field from wrapper")
	}

	if _, hasResponse := properties["response"]; hasResponse {
		t.Error("Schema should not contain 'response' field from wrapper")
	}

	if _, hasSuccess := properties["success"]; !hasSuccess {
		t.Error("Schema should contain 'success' field from unwrapped response")
	}

	if _, hasMessage := properties["message"]; !hasMessage {
		t.Error("Schema should contain 'message' field from unwrapped response")
	}
}

func TestGenerateJSONSchema_NestedStruct(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type User struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	input := User{
		Name: "John",
		Address: Address{
			Street: "123 Main St",
			City:   "Springfield",
		},
	}

	schema := GenerateJSONSchema(input)

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties is not a map")
	}

	addressSchema, ok := properties["address"].(map[string]interface{})
	if !ok {
		t.Fatal("address property not found or not a map")
	}

	if addressSchema["type"] != "object" {
		t.Errorf("address type = %v, want 'object'", addressSchema["type"])
	}

	addressProps, ok := addressSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("address properties not found")
	}

	if len(addressProps) != 2 {
		t.Errorf("Expected 2 properties in address, got %d", len(addressProps))
	}
}

func TestGenerateJSONSchema_JSONTagSkipped(t *testing.T) {
	type TestStruct struct {
		Public  string `json:"public"`
		Skipped string `json:"-"`
		NoTag   string
	}

	input := TestStruct{
		Public:  "visible",
		Skipped: "hidden",
		NoTag:   "also hidden",
	}

	schema := GenerateJSONSchema(input)

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties is not a map")
	}

	if len(properties) != 1 {
		t.Errorf("Expected 1 property, got %d", len(properties))
	}

	if _, hasPublic := properties["public"]; !hasPublic {
		t.Error("Should have 'public' property")
	}

	if _, hasSkipped := properties["skipped"]; hasSkipped {
		t.Error("Should not have 'skipped' property (json:\"-\")")
	}
}

func TestGenerateSchemaForType(t *testing.T) {
	tests := []struct {
		name     string
		typ      reflect.Type
		wantType string
	}{
		{"string type", reflect.TypeOf(""), "string"},
		{"bool type", reflect.TypeOf(true), "boolean"},
		{"int type", reflect.TypeOf(0), "integer"},
		{"float type", reflect.TypeOf(0.0), "number"},
		{"time type", reflect.TypeOf(time.Time{}), "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := generateSchemaForType(tt.typ)

			schemaType, ok := schema["type"].(string)
			if !ok {
				t.Errorf("Type is not a string: %v", schema["type"])
				return
			}

			if schemaType != tt.wantType {
				t.Errorf("Type = %q, want %q", schemaType, tt.wantType)
			}

			// Special check for time.Time format
			if tt.name == "time type" {
				if format, ok := schema["format"].(string); !ok || format != "date-time" {
					t.Errorf("Format = %v, want 'date-time'", schema["format"])
				}
			}
		})
	}
}

func TestGenerateMapSchema(t *testing.T) {
	input := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	schema := GenerateJSONSchema(input)

	if schema["type"] != "object" {
		t.Errorf("Type = %v, want 'object'", schema["type"])
	}

	if _, hasAdditionalProps := schema["additionalProperties"]; !hasAdditionalProps {
		t.Error("Map schema should have additionalProperties")
	}
}
