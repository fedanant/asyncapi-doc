package asyncapi

import (
	"testing"

	"github.com/fedanant/asyncapi-doc/internal/asyncapi/spec3"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()

	if parser == nil {
		t.Fatal("NewParser returned nil")
	}

	if parser.asyncAPI == nil {
		t.Error("asyncApi should be initialized")
	}

	if parser.asyncAPI.Info.Title != "" {
		t.Error("Info.Title should be empty initially")
	}

	if parser.asyncAPI.Channels == nil {
		t.Error("Channels map should be initialized")
	}

	if parser.asyncAPI.Operations == nil {
		t.Error("Operations map should be initialized")
	}
}

func TestParseMain(t *testing.T) {
	tests := []struct {
		name         string
		comments     []string
		wantTitle    string
		wantVersion  string
		wantProtocol string
	}{
		{
			name: "parse basic API info",
			comments: []string{
				"@title Test API",
				"@version 1.0.0",
				"@protocol nats",
				"@url nats://localhost:4222",
			},
			wantTitle:    "Test API",
			wantVersion:  "1.0.0",
			wantProtocol: "nats",
		},
		{
			name: "parse with host instead of url",
			comments: []string{
				"@title Another API",
				"@version 2.0.0",
				"@protocol amqp",
				"@host localhost:5672",
			},
			wantTitle:    "Another API",
			wantVersion:  "2.0.0",
			wantProtocol: "amqp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			parser.ParseMain(tt.comments)

			if parser.asyncAPI.Info.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", parser.asyncAPI.Info.Title, tt.wantTitle)
			}

			if parser.asyncAPI.Info.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", parser.asyncAPI.Info.Version, tt.wantVersion)
			}

			if len(parser.asyncAPI.Servers) == 0 {
				t.Error("Expected at least one server to be created")
			}
		})
	}
}

func TestParseMainWithInfoAnnotations(t *testing.T) {
	tests := []struct {
		name                string
		comments            []string
		wantDescription     string
		wantTermsOfService  string
		wantContactName     string
		wantContactEmail    string
		wantContactURL      string
		wantLicenseName     string
		wantLicenseURL      string
		wantTagsCount       int
		wantExternalDocsURL string
	}{
		{
			name: "parse all info annotations",
			comments: []string{
				"@title Complete API",
				"@version 1.0.0",
				"@description This is a comprehensive API for testing",
				"@termsOfService https://example.com/terms",
				"@contact.name API Support Team",
				"@contact.email support@example.com",
				"@contact.url https://example.com/support",
				"@license.name Apache 2.0",
				"@license.url https://www.apache.org/licenses/LICENSE-2.0.html",
				"@tag users - User management operations",
				"@tag orders - Order processing",
				"@externalDocs.description Find more info here",
				"@externalDocs.url https://docs.example.com",
				"@protocol nats",
				"@url nats://localhost:4222",
			},
			wantDescription:     "This is a comprehensive API for testing",
			wantTermsOfService:  "https://example.com/terms",
			wantContactName:     "API Support Team",
			wantContactEmail:    "support@example.com",
			wantContactURL:      "https://example.com/support",
			wantLicenseName:     "Apache 2.0",
			wantLicenseURL:      "https://www.apache.org/licenses/LICENSE-2.0.html",
			wantTagsCount:       2,
			wantExternalDocsURL: "https://docs.example.com",
		},
		{
			name: "parse partial info annotations",
			comments: []string{
				"@title Minimal API",
				"@version 1.0.0",
				"@description A minimal API",
				"@license.name MIT",
				"@protocol nats",
				"@url nats://localhost:4222",
			},
			wantDescription: "A minimal API",
			wantLicenseName: "MIT",
			wantTagsCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			parser.ParseMain(tt.comments)

			if parser.asyncAPI.Info.Description != tt.wantDescription {
				t.Errorf("Description = %q, want %q", parser.asyncAPI.Info.Description, tt.wantDescription)
			}

			if parser.asyncAPI.Info.TermsOfService != tt.wantTermsOfService {
				t.Errorf("TermsOfService = %q, want %q", parser.asyncAPI.Info.TermsOfService, tt.wantTermsOfService)
			}

			// Test contact fields
			if tt.wantContactName != "" || tt.wantContactEmail != "" || tt.wantContactURL != "" {
				if parser.asyncAPI.Info.Contact == nil {
					t.Fatal("Contact should not be nil")
				}
				if parser.asyncAPI.Info.Contact.Name != tt.wantContactName {
					t.Errorf("Contact.Name = %q, want %q", parser.asyncAPI.Info.Contact.Name, tt.wantContactName)
				}
				if parser.asyncAPI.Info.Contact.Email != tt.wantContactEmail {
					t.Errorf("Contact.Email = %q, want %q", parser.asyncAPI.Info.Contact.Email, tt.wantContactEmail)
				}
				if parser.asyncAPI.Info.Contact.URL != tt.wantContactURL {
					t.Errorf("Contact.URL = %q, want %q", parser.asyncAPI.Info.Contact.URL, tt.wantContactURL)
				}
			}

			// Test license fields
			if tt.wantLicenseName != "" || tt.wantLicenseURL != "" {
				if parser.asyncAPI.Info.License == nil {
					t.Fatal("License should not be nil")
				}
				if parser.asyncAPI.Info.License.Name != tt.wantLicenseName {
					t.Errorf("License.Name = %q, want %q", parser.asyncAPI.Info.License.Name, tt.wantLicenseName)
				}
				if parser.asyncAPI.Info.License.URL != tt.wantLicenseURL {
					t.Errorf("License.URL = %q, want %q", parser.asyncAPI.Info.License.URL, tt.wantLicenseURL)
				}
			}

			// Test tags (now in Info object per AsyncAPI 3.0.0 spec)
			if len(parser.asyncAPI.Info.Tags) != tt.wantTagsCount {
				t.Errorf("Info.Tags count = %d, want %d", len(parser.asyncAPI.Info.Tags), tt.wantTagsCount)
			}

			if tt.wantTagsCount > 0 {
				if parser.asyncAPI.Info.Tags[0].Name != "users" {
					t.Errorf("First tag name = %q, want %q", parser.asyncAPI.Info.Tags[0].Name, "users")
				}
				if parser.asyncAPI.Info.Tags[0].Description != "User management operations" {
					t.Errorf("First tag description = %q, want %q", parser.asyncAPI.Info.Tags[0].Description, "User management operations")
				}
			}

			// Test external docs (now in Info object per AsyncAPI 3.0.0 spec)
			if tt.wantExternalDocsURL != "" {
				if parser.asyncAPI.Info.ExternalDocs == nil {
					t.Fatal("Info.ExternalDocs should not be nil")
				}
				if parser.asyncAPI.Info.ExternalDocs.URL != tt.wantExternalDocsURL {
					t.Errorf("Info.ExternalDocs.URL = %q, want %q", parser.asyncAPI.Info.ExternalDocs.URL, tt.wantExternalDocsURL)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Parser)
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid parser",
			setup: func(p *Parser) {
				p.asyncAPI.Info.Title = "Test API"
				p.asyncAPI.Info.Version = "1.0.0"
				p.asyncAPI.Servers["default"] = spec3.Server{
					Host:     "localhost:4222",
					Protocol: "nats",
				}
			},
			wantErr: false,
		},
		{
			name: "missing title",
			setup: func(p *Parser) {
				p.asyncAPI.Info.Version = "1.0.0"
				p.asyncAPI.Servers["default"] = spec3.Server{
					Host:     "localhost:4222",
					Protocol: "nats",
				}
			},
			wantErr: true,
			errMsg:  "missing required @title annotation in API comments",
		},
		{
			name: "missing version",
			setup: func(p *Parser) {
				p.asyncAPI.Info.Title = "Test API"
				p.asyncAPI.Servers["default"] = spec3.Server{
					Host:     "localhost:4222",
					Protocol: "nats",
				}
			},
			wantErr: true,
			errMsg:  "missing required @version annotation in API comments",
		},
		{
			name: "missing server",
			setup: func(p *Parser) {
				p.asyncAPI.Info.Title = "Test API"
				p.asyncAPI.Info.Version = "1.0.0"
			},
			wantErr: true,
			errMsg:  "missing required server configuration (@url or @host and @protocol)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			tt.setup(parser)

			err := parser.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Validate() error message = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestToChannelName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user.created", "userCreated"},
		{"order.placed", "orderPlaced"},
		{"user.{id}.updated", "userIdUpdated"},
		{"events.{region}.{warehouse}.inventory", "eventsRegionWarehouseInventory"},
		{"simple", "simple"},
		{"with-dashes", "withDashes"},
		{"with_underscores", "withUnderscores"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toChannelName(tt.input)
			if got != tt.want {
				t.Errorf("toChannelName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetermineActionAndName(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name        string
		opType      string
		channelName string
		hasResponse bool
		wantAction  spec3.OperationAction
		wantName    string
	}{
		{"publish operation", "pub", "userCreated", false, spec3.ActionSend, "publishUserCreated"},
		{"subscribe operation", "sub", "userUpdated", false, spec3.ActionReceive, "subscribeUserUpdated"},
		{"request-reply with response", "sub", "getUser", true, spec3.ActionSend, "requestGetUser"},
		{"request-reply overrides pub", "pub", "getUser", true, spec3.ActionSend, "requestGetUser"},
		{"unknown defaults to subscribe", "unknown", "someChannel", false, spec3.ActionReceive, "subscribeSomeChannel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, name := parser.determineActionAndName(tt.opType, tt.channelName, tt.hasResponse)

			if action != tt.wantAction {
				t.Errorf("action = %v, want %v", action, tt.wantAction)
			}

			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
		})
	}
}

func TestCreateChannelParameters(t *testing.T) {
	parser := NewParser()

	params := map[string]ParameterInfo{
		"userId": {
			Schema: map[string]interface{}{
				"type":        "string",
				"description": "User ID",
			},
		},
		"orderId": {
			Schema: map[string]interface{}{
				"type":        "string",
				"description": "Order ID",
			},
		},
	}

	result := parser.createChannelParameters(params)

	if len(result) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(result))
	}

	if result["userId"].Description != "User ID" {
		t.Errorf("userId description = %q, want %q", result["userId"].Description, "User ID")
	}

	if result["orderId"].Description != "Order ID" {
		t.Errorf("orderId description = %q, want %q", result["orderId"].Description, "Order ID")
	}
}

func TestCreateMessage(t *testing.T) {
	parser := NewParser()

	msgInfo := &MessageInfo{
		Summary:     "User created event",
		Description: "Triggered when a user is created",
		MessageSample: struct {
			UserID string `json:"userId"`
			Email  string `json:"email"`
		}{},
	}

	parser.createMessage("userCreatedMessage", msgInfo)

	msg, exists := parser.asyncAPI.Components.Messages["userCreatedMessage"]
	if !exists {
		t.Fatal("Message was not created")
	}

	if msg.Summary != msgInfo.Summary {
		t.Errorf("Summary = %q, want %q", msg.Summary, msgInfo.Summary)
	}

	if msg.Description != msgInfo.Description {
		t.Errorf("Description = %q, want %q", msg.Description, msgInfo.Description)
	}

	if msg.Payload == nil {
		t.Error("Payload should not be nil")
	}
}

func TestCreateChannel(t *testing.T) {
	parser := NewParser()

	params := map[string]spec3.Parameter{
		"userId": {Description: "User ID"},
	}

	parser.createChannel("userCreated", "user.created", "userCreatedMessage", params)

	channel, exists := parser.asyncAPI.Channels["userCreated"]
	if !exists {
		t.Fatal("Channel was not created")
	}

	if channel.Address != "user.created" {
		t.Errorf("Address = %q, want %q", channel.Address, "user.created")
	}

	if len(channel.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(channel.Parameters))
	}

	if _, hasMsg := channel.Messages["userCreatedMessage"]; !hasMsg {
		t.Error("Expected message reference in channel")
	}
}

func TestCreateOperation(t *testing.T) {
	parser := NewParser()

	msgInfo := &MessageInfo{
		Summary:     "Test summary",
		Description: "Test description",
	}

	op := parser.createOperation(spec3.ActionSend, "testChannel", "testMessage", msgInfo)

	if op.Action != spec3.ActionSend {
		t.Errorf("Action = %v, want %v", op.Action, spec3.ActionSend)
	}

	if op.Summary != msgInfo.Summary {
		t.Errorf("Summary = %q, want %q", op.Summary, msgInfo.Summary)
	}

	if op.Description != msgInfo.Description {
		t.Errorf("Description = %q, want %q", op.Description, msgInfo.Description)
	}

	if len(op.Messages) != 1 {
		t.Errorf("Expected 1 message reference, got %d", len(op.Messages))
	}
}
