package asyncapi

import (
	"go/ast"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/modern-go/reflect2"
	"github.com/swaggest/go-asyncapi/reflector/asyncapi-2.4.0"
	"github.com/swaggest/go-asyncapi/spec-2.4.0"
)

type Msg struct {
	Data interface{} `json:"data"`
}

type MsgResponse struct {
	Id       string      `json:"id"`
	Response interface{} `json:"response"`
}

type Operation struct {
	TypeOperation   string
	Name            string
	Message         *asyncapi.MessageSample
	MessageResponse *asyncapi.MessageSample
	Parameters      map[string]spec.Parameter
}

var paramsPattern = regexp.MustCompile("({(.+?)})")

func NewOperation() *Operation {
	return &Operation{
		TypeOperation:   "sub",
		Message:         &asyncapi.MessageSample{},
		MessageResponse: &asyncapi.MessageSample{},
		Parameters:      map[string]spec.Parameter{},
	}
}

func (operation *Operation) ParseComment(comment string, astFile *ast.Package) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "/"))
	if len(commentLine) == 0 {
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
		operation.ParsePayload(lineRemainder, astFile)
	case responseAttr:
		operation.ParseResponse(lineRemainder, astFile)
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
		operation.Parameters[name] = spec.Parameter{
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

func (operation *Operation) ParsePayload(name string, astFile *ast.Package) {
	typeSpec := GetByNameType(name, astFile)
	if typeSpec != nil {
		operation.Message.MessageSample = Msg{
			Data: typeSpec,
		}
	} else {
		log.Println("not found type payload", name)
	}
}

func (operation *Operation) ParseResponse(name string, astFile *ast.Package) {
	typeSpec := GetByNameType(name, astFile)
	if typeSpec != nil {
		operation.MessageResponse.MessageSample = MsgResponse{
			Response: typeSpec,
		}
	} else {
		log.Println("not found type response", name)
	}
}

func GetByNameType(typeName string, astFile *ast.Package) interface{} {
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

	typeInfo := ExtractTypeFromAST(typeName, astFile)
	if typeInfo != nil {
		instance := CreateReflectValue(typeInfo)
		if hasArray {
			sliceType := reflect.SliceOf(reflect.TypeOf(instance))
			return reflect.MakeSlice(sliceType, 0, 0).Interface()
		}
		return instance
	}

	if !strings.Contains(typeName, ".") {
		typeName = astFile.Name + "." + typeName
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
