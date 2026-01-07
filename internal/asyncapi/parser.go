package asyncapi

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/fedanant/asyncapi-doc/internal/asyncapi/spec3"
)

const (
	titleAttr            = "@title"
	urlAttr              = "@url"
	hostAttr             = "@host"
	versionAttr          = "@version"
	typeAttr             = "@type"
	nameAttr             = "@name"
	protocolAttr         = "@protocol"
	descriptionAttr      = "@description"
	summaryAttr          = "@summary"
	payloadAttr          = "@payload"
	responseAttr         = "@response"
	termsOfServiceAttr   = "@termsofservice"
	contactNameAttr      = "@contact.name"
	contactURLAttr       = "@contact.url"
	contactEmailAttr     = "@contact.email"
	licenseNameAttr      = "@license.name"
	licenseURLAttr       = "@license.url"
	tagAttr              = "@tag"
	externalDocsDescAttr = "@externaldocs.description"
	externalDocsURLAttr  = "@externaldocs.url"
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
	var tags []spec3.Tag
	var externalDocs *spec3.ExternalDocs

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
		case descriptionAttr:
			p.asyncApi.Info.Description = value
		case termsOfServiceAttr:
			p.asyncApi.Info.TermsOfService = value
		case contactNameAttr:
			if p.asyncApi.Info.Contact == nil {
				p.asyncApi.Info.Contact = &spec3.Contact{}
			}
			p.asyncApi.Info.Contact.Name = value
		case contactEmailAttr:
			if p.asyncApi.Info.Contact == nil {
				p.asyncApi.Info.Contact = &spec3.Contact{}
			}
			p.asyncApi.Info.Contact.Email = value
		case contactURLAttr:
			if p.asyncApi.Info.Contact == nil {
				p.asyncApi.Info.Contact = &spec3.Contact{}
			}
			p.asyncApi.Info.Contact.URL = value
		case licenseNameAttr:
			if p.asyncApi.Info.License == nil {
				p.asyncApi.Info.License = &spec3.License{}
			}
			p.asyncApi.Info.License.Name = value
		case licenseURLAttr:
			if p.asyncApi.Info.License == nil {
				p.asyncApi.Info.License = &spec3.License{}
			}
			p.asyncApi.Info.License.URL = value
		case tagAttr:
			// Parse tag in format: "name - description" or just "name"
			tagParts := strings.SplitN(value, " - ", 2)
			tag := spec3.Tag{Name: strings.TrimSpace(tagParts[0])}
			if len(tagParts) > 1 {
				tag.Description = strings.TrimSpace(tagParts[1])
			}
			tags = append(tags, tag)
		case externalDocsDescAttr:
			if externalDocs == nil {
				externalDocs = &spec3.ExternalDocs{}
			}
			externalDocs.Description = value
		case externalDocsURLAttr:
			if externalDocs == nil {
				externalDocs = &spec3.ExternalDocs{}
			}
			externalDocs.URL = value
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

	// Store tags and externalDocs in AsyncAPI root level if present
	if len(tags) > 0 {
		p.asyncApi.Tags = tags
	}
	if externalDocs != nil && externalDocs.URL != "" {
		p.asyncApi.ExternalDocs = externalDocs
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

	channelName := toChannelName(operation.Name)
	messageName := channelName + "Message"

	action, operationName := p.determineActionAndName(operation.TypeOperation, channelName)
	channelParams := p.createChannelParameters(operation.Parameters)

	// Create and register the message
	p.createMessage(messageName, operation.Message)

	// Create and register the channel
	p.createChannel(channelName, operation.Name, messageName, channelParams)

	// Create the operation
	op := p.createOperation(action, channelName, messageName, operation.Message)

	// Handle request-reply pattern
	if operation.TypeOperation == "request" && operation.MessageResponse != nil && operation.MessageResponse.MessageSample != nil {
		p.addReplyConfiguration(&op, channelName, operation, channelParams)
	}

	p.asyncApi.Operations[operationName] = op
}

// determineActionAndName returns the action and operation name based on operation type.
func (p *Parser) determineActionAndName(opType string, channelName string) (spec3.OperationAction, string) {
	switch opType {
	case "pub":
		return spec3.ActionSend, "publish" + strings.Title(channelName)
	case "sub":
		return spec3.ActionReceive, "subscribe" + strings.Title(channelName)
	case "request":
		return spec3.ActionSend, "request" + strings.Title(channelName)
	default:
		return spec3.ActionReceive, "subscribe" + strings.Title(channelName)
	}
}

// createChannelParameters converts operation parameters to channel parameters.
func (p *Parser) createChannelParameters(params map[string]ParameterInfo) map[string]spec3.Parameter {
	channelParams := make(map[string]spec3.Parameter)
	for paramName, param := range params {
		channelParams[paramName] = spec3.Parameter{
			Description: getSchemaDescription(param.Schema),
		}
	}
	return channelParams
}

// createMessage creates and registers a message in the components section.
func (p *Parser) createMessage(messageName string, msgInfo *MessageInfo) {
	message := spec3.Message{
		Name:        messageName,
		Summary:     msgInfo.Summary,
		Description: msgInfo.Description,
	}

	if msgInfo.MessageSample != nil {
		schemaName := messageName + "Payload"
		schema := GenerateJSONSchema(msgInfo.MessageSample)
		p.asyncApi.Components.Schemas[schemaName] = schema
		message.Payload = map[string]interface{}{
			"$ref": "#/components/schemas/" + schemaName,
		}
	}

	p.asyncApi.Components.Messages[messageName] = message
}

// createChannel creates and registers a channel.
func (p *Parser) createChannel(channelName, address, messageName string, params map[string]spec3.Parameter) {
	channel := spec3.Channel{
		Address: address,
		Messages: map[string]spec3.MessageRef{
			messageName: {
				Ref: "#/components/messages/" + messageName,
			},
		},
	}

	if len(params) > 0 {
		channel.Parameters = params
	}

	p.asyncApi.Channels[channelName] = channel
}

// createOperation creates an operation structure.
func (p *Parser) createOperation(action spec3.OperationAction, channelName, messageName string, msgInfo *MessageInfo) spec3.Operation {
	return spec3.Operation{
		Action: action,
		Channel: spec3.Reference{
			Ref: "#/channels/" + channelName,
		},
		Summary:     msgInfo.Summary,
		Description: msgInfo.Description,
		Messages: []spec3.Reference{
			{Ref: "#/channels/" + channelName + "/messages/" + messageName},
		},
	}
}

// addReplyConfiguration adds reply channel and message for request-reply pattern.
func (p *Parser) addReplyConfiguration(op *spec3.Operation, channelName string, operation *Operation, channelParams map[string]spec3.Parameter) {
	replyChannelName := channelName + "Reply"
	replyMessageName := replyChannelName + "Message"

	// Create and register reply message
	p.createMessage(replyMessageName, operation.MessageResponse)

	// Create and register reply channel
	p.createChannel(replyChannelName, operation.Name+"/reply", replyMessageName, channelParams)

	// Set reply configuration on operation
	op.Reply = &spec3.OperationReply{
		Channel: &spec3.Reference{
			Ref: "#/channels/" + replyChannelName,
		},
		Messages: []spec3.Reference{
			{Ref: "#/channels/" + replyChannelName + "/messages/" + replyMessageName},
		},
	}
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

// Validate checks that the parser has collected required API information.
func (p *Parser) Validate() error {
	if p.asyncApi.Info.Title == "" {
		return fmt.Errorf("missing required @title annotation in API comments")
	}
	if p.asyncApi.Info.Version == "" {
		return fmt.Errorf("missing required @version annotation in API comments")
	}
	if len(p.asyncApi.Servers) == 0 {
		return fmt.Errorf("missing required server configuration (@url or @host and @protocol)")
	}
	return nil
}

// MarshalYAML serializes the AsyncAPI 3.0 document to YAML format.
func (p *Parser) MarshalYAML() ([]byte, error) {
	return p.asyncApi.MarshalYAML()
}
