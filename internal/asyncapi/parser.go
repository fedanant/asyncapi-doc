package asyncapi

import (
	"go/ast"
	"strings"

	"github.com/swaggest/go-asyncapi/reflector/asyncapi-2.4.0"
	"github.com/swaggest/go-asyncapi/spec-2.4.0"
)

const (
	titleAttr       = "@title"
	urlAttr         = "@url"
	versionAttr     = "@version"
	typeAttr        = "@type"
	nameAttr        = "@name"
	protocolAttr    = "@protocol"
	descriptionAttr = "@description"
	summaryAttr     = "@summary"
	payloadAttr     = "@payload"
	responseAttr    = "@response"
)

type Parser struct {
	asyncApi  *spec.AsyncAPI
	reflector *asyncapi.Reflector
}

func NewParser() *Parser {
	return &Parser{
		asyncApi:  &spec.AsyncAPI{},
		reflector: &asyncapi.Reflector{},
	}
}

func (p *Parser) ParseMain(comments []string) {
	var protocol string
	for i := range comments {
		commentLine := comments[i]
		attribute := strings.Split(commentLine, " ")[0]
		attr := strings.ToLower(attribute)
		value := strings.TrimSpace(commentLine[len(attribute):])
		switch attr {
		case titleAttr:
			p.asyncApi.Info.Title = value
		case versionAttr:
			p.asyncApi.Info.Version = value
		case protocolAttr:
			protocol = value
		case urlAttr:
			p.asyncApi.AddServer(p.asyncApi.Info.Title, spec.Server{
				URL:      value,
				Protocol: protocol,
			})
		}
		p.reflector.Schema = p.asyncApi
	}
}

func (p *Parser) ParseOperation(comments []string, astFile *ast.Package) {
	operation := NewOperation()
	for i := range comments {
		comment := comments[i]
		operation.ParseComment(comment, astFile)
	}
	p.proccessOperation(operation)
}

func (p *Parser) proccessOperation(operation *Operation) {
	channelBase := &spec.ChannelItem{
		Parameters: operation.Parameters,
	}
	if operation.TypeOperation == "pub" {
		p.reflector.AddChannel(asyncapi.ChannelInfo{
			Name:            operation.Name,
			Publish:         operation.Message,
			BaseChannelItem: channelBase,
		})
	} else if operation.TypeOperation == "sub" {
		if operation.MessageResponse.MessageSample != nil {
			p.reflector.AddChannel(asyncapi.ChannelInfo{
				Name:            operation.Name,
				Publish:         operation.Message,
				Subscribe:       operation.MessageResponse,
				BaseChannelItem: channelBase,
			})
		} else {
			p.reflector.AddChannel(asyncapi.ChannelInfo{
				Name:            operation.Name,
				Publish:         operation.Message,
				BaseChannelItem: channelBase,
			})
		}
	}
}

func (p *Parser) MarshalYAML() ([]byte, error) {
	return p.reflector.Schema.MarshalYAML()
}
