package asyncapi

import (
	"go/ast"
	"strings"

	"github.com/fedanant/asyncapi-doc/internal/asyncapi/spec3"
)

const (
	titleAttr       = "@title"
	urlAttr         = "@url"
	hostAttr        = "@host"
	versionAttr     = "@version"
	typeAttr        = "@type"
	nameAttr        = "@name"
	protocolAttr    = "@protocol"
	descriptionAttr = "@description"
	summaryAttr     = "@summary"
	payloadAttr     = "@payload"
	responseAttr    = "@response"
)

// Parser parses Go source comments and generates AsyncAPI 3.0 specifications.
type Parser struct {
	asyncApi *spec3.AsyncAPI
}

// NewParser creates a new Parser with an initialized AsyncAPI 3.0 document.
func NewParser() *Parser {
	return &Parser{
		asyncApi: spec3.NewAsyncAPI(),
	}
}

// ParseMain parses main function comments to extract API info and server configuration.
// In AsyncAPI 3.0, servers use 'host' instead of 'url'.
func (p *Parser) ParseMain(comments []string) {
	var protocol string
	var serverName string
	for i := range comments {
		commentLine := comments[i]
		attribute := strings.Split(commentLine, " ")[0]
		attr := strings.ToLower(attribute)
		value := strings.TrimSpace(commentLine[len(attribute):])
		switch attr {
		case titleAttr:
			p.asyncApi.Info.Title = value
			// Use title as default server name if not set
			if serverName == "" {
				serverName = strings.ReplaceAll(strings.ToLower(value), " ", "-")
			}
		case versionAttr:
			p.asyncApi.Info.Version = value
		case protocolAttr:
			protocol = value
		case urlAttr, hostAttr:
			// AsyncAPI 3.0 uses 'host' instead of 'url'
			// Support both @url and @host for backward compatibility
			if serverName == "" {
				serverName = "default"
			}
			// Strip protocol prefix from host if present (e.g., nats://localhost:4222 -> localhost:4222)
			host := value
			if idx := strings.Index(host, "://"); idx != -1 {
				host = host[idx+3:]
			}
			p.asyncApi.Servers[serverName] = spec3.Server{
				Host:     host,
				Protocol: protocol,
			}
		}
	}
}

// ParseOperation parses operation comments and processes them into AsyncAPI 3.0 structure.
func (p *Parser) ParseOperation(comments []string, astFile *ast.Package) {
	operation := NewOperation()
	for i := range comments {
		comment := comments[i]
		operation.ParseComment(comment, astFile)
	}
	p.proccessOperation(operation)
}

// proccessOperation converts an Operation into AsyncAPI 3.0 channels and operations.
// In AsyncAPI 3.0, channels and operations are separate sections:
// - Channels define addresses and messages
// - Operations define actions (send/receive) with channel references
func (p *Parser) proccessOperation(operation *Operation) {
	if operation.Name == "" {
		return
	}

	// Generate channel and operation names from the operation name
	// e.g., "user.created" -> channelName: "userCreated", operationName: "userCreatedOp"
	channelName := toChannelName(operation.Name)
	messageName := channelName + "Message"
	operationName := channelName

	// Determine the action based on operation type
	// In AsyncAPI 3.0: pub -> send, sub -> receive
	var action spec3.OperationAction
	switch operation.TypeOperation {
	case "pub":
		action = spec3.ActionSend
		operationName = "publish" + strings.Title(channelName)
	case "sub":
		action = spec3.ActionReceive
		operationName = "subscribe" + strings.Title(channelName)
	case "request":
		// Request type: send with reply configuration
		action = spec3.ActionSend
		operationName = "request" + strings.Title(channelName)
	default:
		// Default to receive (subscribe) for unknown types
		action = spec3.ActionReceive
		operationName = "subscribe" + strings.Title(channelName)
	}

	// Create the message in components
	message := spec3.Message{
		Name:        messageName,
		Summary:     operation.Message.Summary,
		Description: operation.Message.Description,
	}

	// Set payload if available - create schema and use $ref
	if operation.Message.MessageSample != nil {
		schemaName := messageName + "Payload"
		schema := GenerateJSONSchema(operation.Message.MessageSample)
		p.asyncApi.Components.Schemas[schemaName] = schema

		// Use $ref to reference the schema
		message.Payload = map[string]interface{}{
			"$ref": "#/components/schemas/" + schemaName,
		}
	}

	p.asyncApi.Components.Messages[messageName] = message

	// Create channel parameters from operation parameters
	channelParams := make(map[string]spec3.Parameter)
	for paramName, param := range operation.Parameters {
		channelParams[paramName] = spec3.Parameter{
			Description: getSchemaDescription(param.Schema),
		}
	}

	// Create the channel
	channel := spec3.Channel{
		Address: operation.Name,
		Messages: map[string]spec3.MessageRef{
			messageName: {
				Ref: "#/components/messages/" + messageName,
			},
		},
	}

	// Only add parameters if there are any
	if len(channelParams) > 0 {
		channel.Parameters = channelParams
	}

	p.asyncApi.Channels[channelName] = channel

	// Create the operation
	op := spec3.Operation{
		Action: action,
		Channel: spec3.Reference{
			Ref: "#/channels/" + channelName,
		},
		Summary:     operation.Message.Summary,
		Description: operation.Message.Description,
		Messages: []spec3.Reference{
			{Ref: "#/channels/" + channelName + "/messages/" + messageName},
		},
	}

	// Handle request type: add reply configuration
	if operation.TypeOperation == "request" && operation.MessageResponse != nil && operation.MessageResponse.MessageSample != nil {
		// Create reply channel and message names
		replyChannelName := channelName + "Reply"
		replyMessageName := replyChannelName + "Message"

		// Create the reply message in components - create schema and use $ref
		replySchemaName := replyMessageName + "Payload"
		replySchema := GenerateJSONSchema(operation.MessageResponse.MessageSample)
		p.asyncApi.Components.Schemas[replySchemaName] = replySchema

		replyMessage := spec3.Message{
			Name:        replyMessageName,
			Summary:     operation.MessageResponse.Summary,
			Description: operation.MessageResponse.Description,
			Payload: map[string]interface{}{
				"$ref": "#/components/schemas/" + replySchemaName,
			},
		}
		p.asyncApi.Components.Messages[replyMessageName] = replyMessage

		// Create the reply channel
		replyChannel := spec3.Channel{
			Address: operation.Name + "/reply",
			Messages: map[string]spec3.MessageRef{
				replyMessageName: {
					Ref: "#/components/messages/" + replyMessageName,
				},
			},
		}

		// Add parameters to reply channel if the original channel has them
		if len(channelParams) > 0 {
			replyChannel.Parameters = channelParams
		}

		p.asyncApi.Channels[replyChannelName] = replyChannel

		// Set the reply configuration on the operation
		op.Reply = &spec3.OperationReply{
			Channel: &spec3.Reference{
				Ref: "#/channels/" + replyChannelName,
			},
			Messages: []spec3.Reference{
				{Ref: "#/channels/" + replyChannelName + "/messages/" + replyMessageName},
			},
		}
	}

	p.asyncApi.Operations[operationName] = op
}

// toChannelName converts a channel address to a valid channel name.
// e.g., "user.created" -> "userCreated", "user.{id}.updated" -> "userIdUpdated"
func toChannelName(address string) string {
	// Remove parameter braces and convert to camelCase
	result := strings.Builder{}
	capitalizeNext := false

	for _, char := range address {
		switch char {
		case '.', '-', '_', '{', '}':
			capitalizeNext = true
		default:
			if capitalizeNext && result.Len() > 0 {
				result.WriteRune(toUpper(char))
			} else {
				result.WriteRune(char)
			}
			capitalizeNext = false
		}
	}

	return result.String()
}

// toUpper converts a rune to uppercase.
func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

// getSchemaDescription extracts description from a schema map.
func getSchemaDescription(schema map[string]interface{}) string {
	if schema == nil {
		return ""
	}
	if desc, ok := schema["description"].(string); ok {
		return desc
	}
	return ""
}

// MarshalYAML serializes the AsyncAPI 3.0 document to YAML format.
func (p *Parser) MarshalYAML() ([]byte, error) {
	return p.asyncApi.MarshalYAML()
}
