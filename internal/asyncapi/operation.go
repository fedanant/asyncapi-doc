package asyncapi

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/modern-go/reflect2"
)

type Msg struct {
	Data interface{} `json:"data"`
}

type MsgResponse struct {
	ID       string      `json:"id"`
	Response interface{} `json:"response"`
}

// MessageInfo holds message metadata for AsyncAPI 3.0 operations.
// Replaces the swaggest asyncapi.MessageSample for 3.0 compatibility.
type MessageInfo struct {
	Summary       string
	Description   string
	MessageSample interface{}
}

// ParameterInfo holds parameter metadata for AsyncAPI 3.0 channels.
// Maintains the Schema map for backward compatibility with how parameters are used.
type ParameterInfo struct {
	Schema map[string]interface{}
}

// Operation represents a parsed AsyncAPI operation from Go comments.
// Updated for AsyncAPI 3.0 compatibility.
type Operation struct {
	TypeOperation   string
	Name            string
	Message         *MessageInfo
	MessageResponse *MessageInfo
	Parameters      map[string]ParameterInfo
}

var paramsPattern = regexp.MustCompile("({(.+?)})")

func NewOperation() *Operation {
	return &Operation{
		TypeOperation:   "sub",
		Message:         &MessageInfo{},
		MessageResponse: &MessageInfo{},
		Parameters:      map[string]ParameterInfo{},
	}
}

func (operation *Operation) ParseComment(comment string, tc *TypeChecker) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "/"))
	if commentLine == "" {
		return nil
	}

	attribute := strings.Fields(commentLine)[0]
	lineRemainder, lowerAttribute := strings.TrimSpace(commentLine[len(attribute):]), strings.ToLower(attribute)
	switch lowerAttribute {
	case typeAttr:
		operation.ParseType(lineRemainder)
	case nameAttr:
		operation.ParseName(lineRemainder)
	case descriptionAttr:
		operation.ParseDescription(lineRemainder)
	case summaryAttr:
		operation.ParseSummary(lineRemainder)
	case payloadAttr:
		if err := operation.ParsePayload(lineRemainder, tc); err != nil {
			log.Printf("Warning: %v", err)
		}
	case responseAttr:
		if err := operation.ParseResponse(lineRemainder, tc); err != nil {
			log.Printf("Warning: %v", err)
		}
	}
	return nil
}

func (operation *Operation) ParseType(typeOperation string) {
	operation.TypeOperation = typeOperation
}

func (operation *Operation) ParseName(name string) {
	operation.Name = name
	params := paramsPattern.FindAllStringSubmatch(name, -1)
	for _, param := range params {
		name := param[2]
		operation.Parameters[name] = ParameterInfo{
			Schema: map[string]interface{}{
				"description": name,
				"type":        "string",
			},
		}
	}
}

func (operation *Operation) ParseDescription(description string) {
	operation.Message.Description = description
}

func (operation *Operation) ParseSummary(summary string) {
	operation.Message.Summary = summary
}

func (operation *Operation) ParsePayload(name string, tc *TypeChecker) error {
	typeSpec := GetByNameType(name, tc)
	if typeSpec != nil {
		operation.Message.MessageSample = Msg{
			Data: typeSpec,
		}
		return nil
	}
	return fmt.Errorf("payload type not found: %s", name)
}

func (operation *Operation) ParseResponse(name string, tc *TypeChecker) error {
	typeSpec := GetByNameType(name, tc)
	if typeSpec != nil {
		operation.MessageResponse.MessageSample = MsgResponse{
			Response: typeSpec,
		}
		return nil
	}
	return fmt.Errorf("response type not found: %s", name)
}

func GetByNameType(typeName string, tc *TypeChecker) interface{} {
	hasArray := false
	originalTypeName := typeName

	if strings.HasPrefix(typeName, "[]") {
		hasArray = true
		typeName = typeName[2:]
	}

	typeSpec := TransToReflectType(typeName)
	if typeSpec != nil {
		if hasArray {
			return []interface{}{typeSpec}
		}
		return typeSpec
	}

	// Use TypeChecker to extract type information
	typeInfo := tc.ExtractTypeInfo(typeName)
	if typeInfo != nil {
		reflectType := tc.GetReflectType(typeInfo)
		instance := reflect.New(reflectType).Elem()
		if hasArray {
			sliceType := reflect.SliceOf(reflectType)
			return reflect.MakeSlice(sliceType, 0, 0).Interface()
		}
		return instance.Interface()
	}

	// Try with package prefix
	if !strings.Contains(typeName, ".") && tc.pkg != nil {
		typeName = tc.pkg.Name() + "." + typeName
	}

	refType := reflect2.TypeByName(typeName)
	if refType != nil {
		if hasArray {
			return reflect.MakeSlice(reflect.SliceOf(refType.Type1()), 0, 10).Interface()
		}

		return refType.New()
	}

	log.Printf("warning: type '%s' not found, using empty struct", originalTypeName)
	return struct{}{}
}

func TransToReflectType(typeName string) interface{} {
	switch typeName {
	case "uint", "int", "uint8", "int8", "uint16", "int16", "byte", "uint32", "int32", "rune", "uint64", "int64":
		return int(0)
	case "float32", "float64":
		return float32(0)
	case "bool":
		return false
	case "string":
		return ""
	}

	return nil
}
