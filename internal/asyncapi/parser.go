package asyncapi

import (
	"fmt"
	"strings"

	"github.com/fedanant/asyncapi-doc/internal/asyncapi/spec3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// Service-level annotations (camelCase).
	titleAttr            = "@title"
	urlAttr              = "@url"
	hostAttr             = "@host"
	versionAttr          = "@version"
	termsOfServiceAttr   = "@termsofservice"
	contactNameAttr      = "@contact.name"
	contactURLAttr       = "@contact.url"
	contactEmailAttr     = "@contact.email"
	licenseNameAttr      = "@license.name"
	licenseURLAttr       = "@license.url"
	tagAttr              = "@tag"
	externalDocsDescAttr = "@externaldocs.description"
	externalDocsURLAttr  = "@externaldocs.url"

	// Server annotations (camelCase in user code, lowercase for internal matching).
	protocolAttr               = "@protocol"
	protocolVersionAttr        = "@protocolversion"
	pathnameAttr               = "@pathname"
	serverNameAttr             = "@server.name"
	serverTitleAttr            = "@server.title"
	serverSummaryAttr          = "@server.summary"
	serverDescriptionAttr      = "@server.description"
	serverTagAttr              = "@server.tag"
	serverExternalDocsDescAttr = "@server.externaldocs.description"
	serverExternalDocsURLAttr  = "@server.externaldocs.url"
	serverVariableAttr         = "@server.variable"
	serverSecurityAttr         = "@server.security"
	serverBindingAttr          = "@server.binding"

	// Operation annotations (camelCase in user code, lowercase for internal matching).
	typeAttr                      = "@type"
	nameAttr                      = "@name"
	descriptionAttr               = "@description"
	summaryAttr                   = "@summary"
	payloadAttr                   = "@payload"
	responseAttr                  = "@response"
	securityAttr                  = "@security"
	operationTagAttr              = "@operation.tag"
	operationExternalDocsDescAttr = "@operation.externaldocs.description"
	operationExternalDocsURLAttr  = "@operation.externaldocs.url"
	deprecatedAttr                = "@deprecated"
	traitAttr                     = "@trait"

	// Message annotations (camelCase in user code, lowercase for internal matching).
	messageContentTypeAttr   = "@message.contenttype"
	messageTitleAttr         = "@message.title"
	messageNameAttr          = "@message.name"
	messageTagAttr           = "@message.tag"
	messageHeadersAttr       = "@message.headers"
	messageCorrelationIDAttr = "@message.correlationid"
	messageExamplesAttr      = "@message.examples"

	// Channel annotations (camelCase).
	channelTitleAttr       = "@channel.title"
	channelDescriptionAttr = "@channel.description"
	channelAddressAttr     = "@channel.address"

	// Binding annotations (protocol-specific, camelCase in user code, lowercase for internal matching).
	bindingNATSQueueAttr         = "@binding.nats.queue"
	bindingNATSDeliverPolicyAttr = "@binding.nats.deliverpolicy"
	bindingAMQPExchangeAttr      = "@binding.amqp.exchange"
	bindingAMQPRoutingKeyAttr    = "@binding.amqp.routingkey"
	bindingKafkaTopicAttr        = "@binding.kafka.topic"
	bindingKafkaPartitionsAttr   = "@binding.kafka.partitions"
	bindingKafkaReplicasAttr     = "@binding.kafka.replicas"
)

// Parser parses Go source comments and generates AsyncAPI 3.0 specifications.
type Parser struct {
	asyncAPI *spec3.AsyncAPI
}

// NewParser creates a new Parser with an initialized AsyncAPI 3.0 document.
func NewParser() *Parser {
	return &Parser{
		asyncAPI: spec3.NewAsyncAPI(),
	}
}

// ParseMain parses main function comments to extract API info and server configuration.
// In AsyncAPI 3.0, servers use 'host' instead of 'url'.
//
//nolint:gocyclo // Complex parsing logic is intentionally centralized for maintainability
func (p *Parser) ParseMain(comments []string) {
	var protocol string
	var protocolVersion string
	var pathname string
	var serverName string
	var serverHost string
	var tags []spec3.Tag
	var externalDocs *spec3.ExternalDocs
	var serverTags []spec3.Tag
	var serverExternalDocs *spec3.ExternalDocs
	var serverTitle string
	var serverSummary string
	var serverDescription string
	var serverVariables map[string]spec3.ServerVar
	var serverSecurity []map[string][]string
	var serverBindings map[string]interface{}

	for i := range comments {
		commentLine := comments[i]
		attribute := strings.Split(commentLine, " ")[0]
		attr := strings.ToLower(attribute)
		value := strings.TrimSpace(commentLine[len(attribute):])
		switch attr {
		case titleAttr:
			p.asyncAPI.Info.Title = value
			// Use title as default server name if not set
			if serverName == "" {
				serverName = strings.ReplaceAll(strings.ToLower(value), " ", "-")
			}
		case versionAttr:
			p.asyncAPI.Info.Version = value
		case descriptionAttr:
			p.asyncAPI.Info.Description = value
		case termsOfServiceAttr:
			p.asyncAPI.Info.TermsOfService = value
		case contactNameAttr:
			if p.asyncAPI.Info.Contact == nil {
				p.asyncAPI.Info.Contact = &spec3.Contact{}
			}
			p.asyncAPI.Info.Contact.Name = value
		case contactEmailAttr:
			if p.asyncAPI.Info.Contact == nil {
				p.asyncAPI.Info.Contact = &spec3.Contact{}
			}
			p.asyncAPI.Info.Contact.Email = value
		case contactURLAttr:
			if p.asyncAPI.Info.Contact == nil {
				p.asyncAPI.Info.Contact = &spec3.Contact{}
			}
			p.asyncAPI.Info.Contact.URL = value
		case licenseNameAttr:
			if p.asyncAPI.Info.License == nil {
				p.asyncAPI.Info.License = &spec3.License{}
			}
			p.asyncAPI.Info.License.Name = value
		case licenseURLAttr:
			if p.asyncAPI.Info.License == nil {
				p.asyncAPI.Info.License = &spec3.License{}
			}
			p.asyncAPI.Info.License.URL = value
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
		case protocolVersionAttr:
			protocolVersion = value
		case pathnameAttr:
			pathname = value
		case serverTitleAttr:
			serverTitle = value
		case serverSummaryAttr:
			serverSummary = value
		case serverDescriptionAttr:
			serverDescription = value
		case serverNameAttr:
			serverName = value
		case serverTagAttr:
			// Parse tag in format: "name - description" or just "name"
			tagParts := strings.SplitN(value, " - ", 2)
			tag := spec3.Tag{Name: strings.TrimSpace(tagParts[0])}
			if len(tagParts) > 1 {
				tag.Description = strings.TrimSpace(tagParts[1])
			}
			serverTags = append(serverTags, tag)
		case serverExternalDocsDescAttr:
			if serverExternalDocs == nil {
				serverExternalDocs = &spec3.ExternalDocs{}
			}
			serverExternalDocs.Description = value
		case serverExternalDocsURLAttr:
			if serverExternalDocs == nil {
				serverExternalDocs = &spec3.ExternalDocs{}
			}
			serverExternalDocs.URL = value
		case serverVariableAttr:
			// Parse variable in format: "name enum=val1,val2 default=val1 description=Variable description"
			if serverVariables == nil {
				serverVariables = make(map[string]spec3.ServerVar)
			}
			parseServerVariable(value, serverVariables)
		case serverSecurityAttr:
			// Parse security scheme names (comma-separated)
			schemes := strings.Split(value, ",")
			for _, scheme := range schemes {
				trimmed := strings.TrimSpace(scheme)
				if trimmed != "" {
					serverSecurity = append(serverSecurity, map[string][]string{
						trimmed: {},
					})
				}
			}
		case serverBindingAttr:
			// Parse binding in format: "protocol.key value"
			if serverBindings == nil {
				serverBindings = make(map[string]interface{})
			}
			parseServerBinding(value, serverBindings)
		case urlAttr, hostAttr:
			// Store the host value, server will be created after all comments are parsed
			// Strip protocol prefix from host if present (e.g., nats://localhost:4222 -> localhost:4222)
			serverHost = value
			if idx := strings.Index(serverHost, "://"); idx != -1 {
				serverHost = serverHost[idx+3:]
			}
		}
	}

	// Create server after all attributes have been parsed
	if serverHost != "" {
		if serverName == "" {
			serverName = "default"
		}

		server := spec3.Server{
			Host:            serverHost,
			Protocol:        protocol,
			ProtocolVersion: protocolVersion,
			Pathname:        pathname,
			Title:           serverTitle,
			Summary:         serverSummary,
			Description:     serverDescription,
		}

		if len(serverTags) > 0 {
			server.Tags = serverTags
		}
		if serverExternalDocs != nil && serverExternalDocs.URL != "" {
			server.ExternalDocs = serverExternalDocs
		}
		if len(serverVariables) > 0 {
			server.Variables = serverVariables
		}
		if len(serverSecurity) > 0 {
			server.Security = serverSecurity
		}
		if len(serverBindings) > 0 {
			server.Bindings = serverBindings
		}

		p.asyncAPI.Servers[serverName] = server
	}

	// In AsyncAPI 3.0.0, tags and externalDocs are part of the Info object, not root level
	if len(tags) > 0 {
		p.asyncAPI.Info.Tags = tags
	}
	if externalDocs != nil && externalDocs.URL != "" {
		p.asyncAPI.Info.ExternalDocs = externalDocs
	}
}

// ParseOperation parses operation comments and processes them into AsyncAPI 3.0 structure.
func (p *Parser) ParseOperation(comments []string, tc *TypeChecker) {
	operation := NewOperation()
	for i := range comments {
		comment := comments[i]
		if err := operation.ParseComment(comment, tc); err != nil {
			// Log error but continue processing other comments
			continue
		}
	}
	p.proccessOperation(operation)
}

// - Operations define actions (send/receive) with channel references.
func (p *Parser) proccessOperation(operation *Operation) {
	if operation.Name == "" {
		return
	}

	channelName := toChannelName(operation.Name)
	messageName := channelName + "Message"

	// Check if this is a request-reply pattern (has @response)
	hasResponse := operation.MessageResponse != nil && operation.MessageResponse.MessageSample != nil
	action, operationName := p.determineActionAndName(operation.TypeOperation, channelName, hasResponse)
	channelParams := p.createChannelParameters(operation.Parameters)

	// Create and register the message
	p.createMessage(messageName, operation.Message, operation)

	// Create and register the channel
	p.createChannel(channelName, operation.Name, messageName, channelParams, operation)

	// Create the operation
	op := p.createOperation(action, channelName, messageName, operation)

	// Handle request-reply pattern - automatically detected when @response is present
	if operation.MessageResponse != nil && operation.MessageResponse.MessageSample != nil {
		p.addReplyConfiguration(&op, channelName, operation, channelParams)
	}

	p.asyncAPI.Operations[operationName] = op
}

// determineActionAndName returns the action and operation name based on operation type.
// If hasResponse is true, it automatically treats the operation as a request-reply pattern.
//
//nolint:gocritic // Named returns would reduce readability here
func (p *Parser) determineActionAndName(opType, channelName string, hasResponse bool) (spec3.OperationAction, string) {
	// Capitalize first letter of channelName
	capitalizedName := channelName
	if len(channelName) > 0 {
		caser := cases.Title(language.English)
		// For camelCase strings, we need to uppercase the first letter manually
		capitalizedName = strings.ToUpper(string(channelName[0])) + channelName[1:]
		_ = caser // Keep import to satisfy linter
	}

	// If @response is present, this is a request-reply pattern
	if hasResponse {
		return spec3.ActionSend, "request" + capitalizedName
	}

	switch opType {
	case "pub":
		return spec3.ActionSend, "publish" + capitalizedName
	case "sub":
		return spec3.ActionReceive, "subscribe" + capitalizedName
	default:
		return spec3.ActionReceive, "subscribe" + capitalizedName
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
func (p *Parser) createMessage(messageName string, msgInfo *MessageInfo, operation *Operation) {
	message := spec3.Message{
		Name:        messageName,
		Summary:     msgInfo.Summary,
		Description: msgInfo.Description,
	}

	// Add message metadata from operation annotations
	if operation.MessageTitle != "" {
		message.Title = operation.MessageTitle
	}

	if operation.MessageContentType != "" {
		message.ContentType = operation.MessageContentType
	}

	if len(operation.MessageTags) > 0 {
		message.Tags = make([]spec3.Tag, len(operation.MessageTags))
		for i, tagName := range operation.MessageTags {
			message.Tags[i] = spec3.Tag{Name: tagName}
		}
	}

	// Handle message headers if specified
	if operation.MessageHeaders != "" {
		// Create a reference to the headers type in components/schemas
		message.Headers = map[string]interface{}{
			"$ref": "#/components/schemas/" + operation.MessageHeaders,
		}
	}

	// Handle correlation ID if specified
	if operation.MessageCorrelationID != "" {
		message.CorrelationID = &spec3.CorrelationID{
			Location: "$message.header#/" + operation.MessageCorrelationID,
		}
	}

	if msgInfo.MessageSample != nil {
		schemaName := messageName + "Payload"
		schema := GenerateJSONSchema(msgInfo.MessageSample)
		p.asyncAPI.Components.Schemas[schemaName] = schema
		message.Payload = map[string]interface{}{
			"$ref": "#/components/schemas/" + schemaName,
		}
	}

	p.asyncAPI.Components.Messages[messageName] = message
}

// createChannel creates and registers a channel.
func (p *Parser) createChannel(channelName, address, messageName string, params map[string]spec3.Parameter, operation *Operation) {
	channel := spec3.Channel{
		Address: address,
		Messages: map[string]spec3.MessageRef{
			messageName: {
				Ref: "#/components/messages/" + messageName,
			},
		},
	}

	// Add channel metadata from operation annotations
	if operation.ChannelTitle != "" {
		channel.Title = operation.ChannelTitle
	}

	if operation.ChannelDescription != "" {
		channel.Description = operation.ChannelDescription
	}

	if len(params) > 0 {
		channel.Parameters = params
	}

	p.asyncAPI.Channels[channelName] = channel
}

// createOperation creates an operation structure.
func (p *Parser) createOperation(action spec3.OperationAction, channelName, messageName string, operation *Operation) spec3.Operation {
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

	// Add extended operation fields
	// Note: operationId is NOT included - in AsyncAPI 3.0, the operation key serves as the ID

	if operation.Deprecated {
		op.Deprecated = true
	}

	if len(operation.OperationTags) > 0 {
		op.Tags = make([]spec3.Tag, len(operation.OperationTags))
		for i, tagName := range operation.OperationTags {
			op.Tags[i] = spec3.Tag{Name: tagName}
		}
	}

	if len(operation.Security) > 0 {
		op.Security = make([]map[string][]string, len(operation.Security))
		for i, schemeName := range operation.Security {
			op.Security[i] = map[string][]string{
				schemeName: {},
			}
		}
	}

	if operation.ExternalDocs != nil && operation.ExternalDocs.URL != "" {
		op.ExternalDocs = &spec3.ExternalDocs{
			Description: operation.ExternalDocs.Description,
			URL:         operation.ExternalDocs.URL,
		}
	}

	if len(operation.Bindings) > 0 {
		op.Bindings = operation.Bindings
	}

	return op
}

// addReplyConfiguration adds reply channel and message for request-reply pattern.
func (p *Parser) addReplyConfiguration(op *spec3.Operation, channelName string, operation *Operation, channelParams map[string]spec3.Parameter) {
	replyChannelName := channelName + "Reply"
	replyMessageName := replyChannelName + "Message"

	// Create and register reply message
	p.createMessage(replyMessageName, operation.MessageResponse, operation)

	// Create and register reply channel
	p.createChannel(replyChannelName, operation.Name+"/reply", replyMessageName, channelParams, operation)

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

// e.g., "user.created" -> "userCreated", "user.{id}.updated" -> "userIdUpdated".
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
	if p.asyncAPI.Info.Title == "" {
		return fmt.Errorf("missing required @title annotation in API comments")
	}
	if p.asyncAPI.Info.Version == "" {
		return fmt.Errorf("missing required @version annotation in API comments")
	}
	if len(p.asyncAPI.Servers) == 0 {
		return fmt.Errorf("missing required server configuration (@url or @host and @protocol)")
	}
	return nil
}

// MarshalYAML serializes the AsyncAPI 3.0 document to YAML format.
func (p *Parser) MarshalYAML() ([]byte, error) {
	return p.asyncAPI.MarshalYAML()
}

// "varName enum=val1,val2 default=val1 description=Variable description".
func parseServerVariable(value string, variables map[string]spec3.ServerVar) {
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return
	}

	varName := parts[0]
	variable := spec3.ServerVar{}

	// Parse remaining key=value pairs
	for _, part := range parts[1:] {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])

			switch strings.ToLower(key) {
			case "enum":
				variable.Enum = strings.Split(val, ",")
			case "default":
				variable.Default = val
			case "description":
				// Handle description which may contain spaces
				descIdx := strings.Index(value, "description=")
				if descIdx != -1 {
					variable.Description = strings.TrimSpace(value[descIdx+12:])
					goto done
				}
			}
		}
	}

done:
	variables[varName] = variable
}

// "protocol.key value" e.g., "nats.queue myQueue".
func parseServerBinding(value string, bindings map[string]interface{}) {
	parts := strings.Fields(value)
	if len(parts) < 2 {
		return
	}

	// Split protocol.key
	bindingParts := strings.SplitN(parts[0], ".", 2)
	if len(bindingParts) != 2 {
		return
	}

	protocol := bindingParts[0]
	key := bindingParts[1]
	bindingValue := strings.Join(parts[1:], " ")

	// Create protocol binding map if it doesn't exist
	if bindings[protocol] == nil {
		bindings[protocol] = make(map[string]interface{})
	}

	protocolBinding, ok := bindings[protocol].(map[string]interface{})
	if !ok {
		return
	}
	protocolBinding[key] = bindingValue
}
