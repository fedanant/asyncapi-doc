// Package spec3 provides types for AsyncAPI 3.0.0 specification.
// This package was created because swaggest/go-asyncapi only supports AsyncAPI 2.x.
// Reference: https://www.asyncapi.com/docs/reference/specification/v3.0.0
package spec3

import "gopkg.in/yaml.v3"

// AsyncAPI represents the root object of an AsyncAPI 3.0.0 document.
type AsyncAPI struct {
	AsyncAPI           string               `json:"asyncapi" yaml:"asyncapi"`
	ID                 string               `json:"id,omitempty" yaml:"id,omitempty"`
	Info               Info                 `json:"info" yaml:"info"`
	Servers            map[string]Server    `json:"servers,omitempty" yaml:"servers,omitempty"`
	DefaultContentType string               `json:"defaultContentType,omitempty" yaml:"defaultContentType,omitempty"`
	Channels           map[string]Channel   `json:"channels,omitempty" yaml:"channels,omitempty"`
	Operations         map[string]Operation `json:"operations,omitempty" yaml:"operations,omitempty"`
	Components         *Components          `json:"components,omitempty" yaml:"components,omitempty"`
	Tags               []Tag                `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs       *ExternalDocs        `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// NewAsyncAPI creates a new AsyncAPI 3.0.0 document with default values.
func NewAsyncAPI() *AsyncAPI {
	return &AsyncAPI{
		AsyncAPI:   "3.0.0",
		Servers:    make(map[string]Server),
		Channels:   make(map[string]Channel),
		Operations: make(map[string]Operation),
		Components: &Components{
			Messages: make(map[string]Message),
			Schemas:  make(map[string]interface{}),
		},
	}
}

// Info provides metadata about the API.
type Info struct {
	Title          string   `json:"title" yaml:"title"`
	Version        string   `json:"version" yaml:"version"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License `json:"license,omitempty" yaml:"license,omitempty"`
}

// Contact information for the exposed API.
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// License information for the exposed API.
type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Server represents a server object in AsyncAPI 3.0.
// In 3.0, 'url' is replaced with 'host' and optional 'pathname'.
type Server struct {
	Host        string                 `json:"host" yaml:"host"`
	Protocol    string                 `json:"protocol" yaml:"protocol"`
	Pathname    string                 `json:"pathname,omitempty" yaml:"pathname,omitempty"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Title       string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Variables   map[string]ServerVar   `json:"variables,omitempty" yaml:"variables,omitempty"`
	Security    []map[string][]string  `json:"security,omitempty" yaml:"security,omitempty"`
	Tags        []Tag                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	Bindings    map[string]interface{} `json:"bindings,omitempty" yaml:"bindings,omitempty"`
}

// ServerVar represents a server variable for server URL template substitution.
type ServerVar struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default,omitempty" yaml:"default,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Examples    []string `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// Tag represents a tag object.
type Tag struct {
	Name         string        `json:"name" yaml:"name"`
	Description  string        `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// ExternalDocs allows referencing external documentation.
type ExternalDocs struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url"`
}

// Channel represents a channel in AsyncAPI 3.0.
// In 3.0, channels are separate from operations and only define the address and messages.
type Channel struct {
	Address     string                 `json:"address,omitempty" yaml:"address,omitempty"`
	Title       string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Messages    map[string]MessageRef  `json:"messages,omitempty" yaml:"messages,omitempty"`
	Parameters  map[string]Parameter   `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Servers     []Reference            `json:"servers,omitempty" yaml:"servers,omitempty"`
	Tags        []Tag                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	Bindings    map[string]interface{} `json:"bindings,omitempty" yaml:"bindings,omitempty"`
}

// Parameter represents a channel parameter.
type Parameter struct {
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Default     string   `json:"default,omitempty" yaml:"default,omitempty"`
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Examples    []string `json:"examples,omitempty" yaml:"examples,omitempty"`
	Location    string   `json:"location,omitempty" yaml:"location,omitempty"`
}

// Operation represents an operation in AsyncAPI 3.0.
// In 3.0, operations are separate from channels and define the action (send/receive).
type Operation struct {
	Action      OperationAction        `json:"action" yaml:"action"`
	Channel     Reference              `json:"channel" yaml:"channel"`
	Title       string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Messages    []Reference            `json:"messages,omitempty" yaml:"messages,omitempty"`
	Reply       *OperationReply        `json:"reply,omitempty" yaml:"reply,omitempty"`
	Traits      []Reference            `json:"traits,omitempty" yaml:"traits,omitempty"`
	Tags        []Tag                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	Bindings    map[string]interface{} `json:"bindings,omitempty" yaml:"bindings,omitempty"`
	Security    []map[string][]string  `json:"security,omitempty" yaml:"security,omitempty"`
}

// OperationAction represents the action type of an operation.
type OperationAction string

const (
	// ActionSend represents an outgoing message (equivalent to 2.x 'publish').
	ActionSend OperationAction = "send"
	// ActionReceive represents an incoming message (equivalent to 2.x 'subscribe').
	ActionReceive OperationAction = "receive"
)

// OperationReply represents the reply configuration for request/reply patterns.
type OperationReply struct {
	Address  *OperationReplyAddress `json:"address,omitempty" yaml:"address,omitempty"`
	Channel  *Reference             `json:"channel,omitempty" yaml:"channel,omitempty"`
	Messages []Reference            `json:"messages,omitempty" yaml:"messages,omitempty"`
}

// OperationReplyAddress represents the address for a reply.
type OperationReplyAddress struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Location    string `json:"location" yaml:"location"`
}

// Message represents a message object in AsyncAPI 3.0.
type Message struct {
	Name          string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Title         string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Summary       string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string                 `json:"description,omitempty" yaml:"description,omitempty"`
	ContentType   string                 `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Payload       interface{}            `json:"payload,omitempty" yaml:"payload,omitempty"`
	Headers       interface{}            `json:"headers,omitempty" yaml:"headers,omitempty"`
	CorrelationID *CorrelationID         `json:"correlationId,omitempty" yaml:"correlationId,omitempty"`
	Tags          []Tag                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	Bindings      map[string]interface{} `json:"bindings,omitempty" yaml:"bindings,omitempty"`
	Traits        []Reference            `json:"traits,omitempty" yaml:"traits,omitempty"`
}

// MessageRef can be either a direct Message or a Reference.
type MessageRef struct {
	Ref     string   `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Message *Message `json:"-" yaml:"-"`
}

// CorrelationID specifies an identifier for message correlation.
type CorrelationID struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Location    string `json:"location" yaml:"location"`
}

// Reference represents a $ref to another object.
type Reference struct {
	Ref string `json:"$ref" yaml:"$ref"`
}

// Components holds reusable objects for the specification.
type Components struct {
	Schemas           map[string]interface{}           `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Servers           map[string]Server                `json:"servers,omitempty" yaml:"servers,omitempty"`
	Channels          map[string]Channel               `json:"channels,omitempty" yaml:"channels,omitempty"`
	Operations        map[string]Operation             `json:"operations,omitempty" yaml:"operations,omitempty"`
	Messages          map[string]Message               `json:"messages,omitempty" yaml:"messages,omitempty"`
	SecuritySchemes   map[string]SecurityScheme        `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Parameters        map[string]Parameter             `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	CorrelationIDs    map[string]CorrelationID         `json:"correlationIds,omitempty" yaml:"correlationIds,omitempty"`
	OperationTraits   map[string]OperationTrait        `json:"operationTraits,omitempty" yaml:"operationTraits,omitempty"`
	MessageTraits     map[string]MessageTrait          `json:"messageTraits,omitempty" yaml:"messageTraits,omitempty"`
	Replies           map[string]OperationReply        `json:"replies,omitempty" yaml:"replies,omitempty"`
	ReplyAddresses    map[string]OperationReplyAddress `json:"replyAddresses,omitempty" yaml:"replyAddresses,omitempty"`
	ServerBindings    map[string]interface{}           `json:"serverBindings,omitempty" yaml:"serverBindings,omitempty"`
	ChannelBindings   map[string]interface{}           `json:"channelBindings,omitempty" yaml:"channelBindings,omitempty"`
	OperationBindings map[string]interface{}           `json:"operationBindings,omitempty" yaml:"operationBindings,omitempty"`
	MessageBindings   map[string]interface{}           `json:"messageBindings,omitempty" yaml:"messageBindings,omitempty"`
}

// SecurityScheme defines a security scheme.
type SecurityScheme struct {
	Type             string      `json:"type" yaml:"type"`
	Description      string      `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty"`
	In               string      `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty" yaml:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
	Scopes           []string    `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

// OAuthFlows defines OAuth flows configuration.
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow represents a single OAuth flow configuration.
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	AvailableScopes  map[string]string `json:"availableScopes,omitempty" yaml:"availableScopes,omitempty"`
}

// OperationTrait represents an operation trait for reuse.
type OperationTrait struct {
	Title       string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []Tag                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	Bindings    map[string]interface{} `json:"bindings,omitempty" yaml:"bindings,omitempty"`
	Security    []map[string][]string  `json:"security,omitempty" yaml:"security,omitempty"`
}

// MessageTrait represents a message trait for reuse.
type MessageTrait struct {
	Name          string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Title         string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Summary       string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string                 `json:"description,omitempty" yaml:"description,omitempty"`
	ContentType   string                 `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       interface{}            `json:"headers,omitempty" yaml:"headers,omitempty"`
	CorrelationID *CorrelationID         `json:"correlationId,omitempty" yaml:"correlationId,omitempty"`
	Tags          []Tag                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	Bindings      map[string]interface{} `json:"bindings,omitempty" yaml:"bindings,omitempty"`
}

// MarshalYAML serializes the AsyncAPI document to YAML format.
func (a *AsyncAPI) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(a)
}
