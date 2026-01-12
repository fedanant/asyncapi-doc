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
// Updated for AsyncAPI 3.0 compatibility with extended annotations support.
type Operation struct {
	TypeOperation   string
	Name            string
	Message         *MessageInfo
	MessageResponse *MessageInfo
	Parameters      map[string]ParameterInfo

	// Extended operation fields
	Security      []string               // @security
	OperationTags []string               // @operation.tag
	Deprecated    bool                   // @deprecated
	ExternalDocs  *ExternalDocsInfo      // @operation.externaldocs.*
	Bindings      map[string]interface{} // @binding.*

	// Channel metadata
	ChannelTitle       string // @channel.title
	ChannelDescription string // @channel.description

	// Message metadata
	MessageContentType   string   // @message.contenttype
	MessageTitle         string   // @message.title
	MessageTags          []string // @message.tag
	MessageHeaders       string   // @message.headers (type name)
	MessageCorrelationID string   // @message.correlationid
}

// ExternalDocsInfo holds external documentation metadata
type ExternalDocsInfo struct {
	Description string
	URL         string
}

var paramsPattern = regexp.MustCompile("({(.+?)})")

func NewOperation() *Operation {
	return &Operation{
		TypeOperation:   "sub",
		Message:         &MessageInfo{},
		MessageResponse: &MessageInfo{},
		Parameters:      map[string]ParameterInfo{},
		Security:        []string{},
		OperationTags:   []string{},
		MessageTags:     []string{},
		Bindings:        make(map[string]interface{}),
		Deprecated:      false,
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
	// Extended operation annotations
	case securityAttr:
		operation.ParseSecurity(lineRemainder)
	case operationTagAttr:
		operation.ParseOperationTag(lineRemainder)
	case deprecatedAttr:
		operation.ParseDeprecated(lineRemainder)
	case operationExternalDocsDescAttr:
		operation.ParseOperationExternalDocsDesc(lineRemainder)
	case operationExternalDocsURLAttr:
		operation.ParseOperationExternalDocsURL(lineRemainder)
	// Message annotations
	case messageContentTypeAttr:
		operation.MessageContentType = lineRemainder
	case messageTitleAttr:
		operation.MessageTitle = lineRemainder
	case messageTagAttr:
		operation.ParseMessageTag(lineRemainder)
	case messageHeadersAttr:
		operation.MessageHeaders = lineRemainder
	case messageCorrelationIDAttr:
		operation.MessageCorrelationID = lineRemainder
	// Channel annotations
	case channelTitleAttr:
		operation.ChannelTitle = lineRemainder
	case channelDescriptionAttr:
		operation.ChannelDescription = lineRemainder
	// Binding annotations
	case bindingNATSQueueAttr:
		operation.ParseBindingNATS("queue", lineRemainder)
	case bindingNATSDeliverPolicyAttr:
		operation.ParseBindingNATS("deliverPolicy", lineRemainder)
	case bindingAMQPExchangeAttr:
		operation.ParseBindingAMQP("exchange", lineRemainder)
	case bindingAMQPRoutingKeyAttr:
		operation.ParseBindingAMQP("routingKey", lineRemainder)
	case bindingKafkaTopicAttr:
		operation.ParseBindingKafka("topic", lineRemainder)
	case bindingKafkaPartitionsAttr:
		operation.ParseBindingKafka("partitions", lineRemainder)
	case bindingKafkaReplicasAttr:
		operation.ParseBindingKafka("replicas", lineRemainder)
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

// ParseSecurity parses comma-separated security scheme names
func (operation *Operation) ParseSecurity(value string) {
	schemes := strings.Split(value, ",")
	for _, scheme := range schemes {
		trimmed := strings.TrimSpace(scheme)
		if trimmed != "" {
			operation.Security = append(operation.Security, trimmed)
		}
	}
}

// ParseOperationTag adds an operation tag
func (operation *Operation) ParseOperationTag(value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" {
		operation.OperationTags = append(operation.OperationTags, trimmed)
	}
}

// ParseDeprecated marks the operation as deprecated
func (operation *Operation) ParseDeprecated(value string) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	operation.Deprecated = trimmed == "true" || trimmed == ""
}

// ParseOperationExternalDocsDesc sets the external docs description
func (operation *Operation) ParseOperationExternalDocsDesc(value string) {
	if operation.ExternalDocs == nil {
		operation.ExternalDocs = &ExternalDocsInfo{}
	}
	operation.ExternalDocs.Description = strings.TrimSpace(value)
}

// ParseOperationExternalDocsURL sets the external docs URL
func (operation *Operation) ParseOperationExternalDocsURL(value string) {
	if operation.ExternalDocs == nil {
		operation.ExternalDocs = &ExternalDocsInfo{}
	}
	operation.ExternalDocs.URL = strings.TrimSpace(value)
}

// ParseMessageTag adds a message tag
func (operation *Operation) ParseMessageTag(value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" {
		operation.MessageTags = append(operation.MessageTags, trimmed)
	}
}

// ParseBindingNATS parses NATS-specific binding properties
func (operation *Operation) ParseBindingNATS(key, value string) {
	if operation.Bindings["nats"] == nil {
		operation.Bindings["nats"] = make(map[string]interface{})
	}
	natsBinding := operation.Bindings["nats"].(map[string]interface{})
	natsBinding[key] = strings.TrimSpace(value)
}

// ParseBindingAMQP parses AMQP-specific binding properties
func (operation *Operation) ParseBindingAMQP(key, value string) {
	if operation.Bindings["amqp"] == nil {
		operation.Bindings["amqp"] = make(map[string]interface{})
	}
	amqpBinding := operation.Bindings["amqp"].(map[string]interface{})
	amqpBinding[key] = strings.TrimSpace(value)
}

// ParseBindingKafka parses Kafka-specific binding properties
func (operation *Operation) ParseBindingKafka(key, value string) {
	if operation.Bindings["kafka"] == nil {
		operation.Bindings["kafka"] = make(map[string]interface{})
	}
	kafkaBinding := operation.Bindings["kafka"].(map[string]interface{})

	// Handle numeric fields
	trimmed := strings.TrimSpace(value)
	switch key {
	case "partitions", "replicas":
		kafkaBinding[key] = trimmed // Store as string, can be converted later if needed
	default:
		kafkaBinding[key] = trimmed
	}
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
